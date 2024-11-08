package aws_provider

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2Type "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	eksType "github.com/aws/aws-sdk-go-v2/service/eks/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3type "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"

	"rmk/util"
)

const (
	AWSClusterProvider = "aws"

	AWSConfigTemplateFile = `[profile {{ .Profile }}]
region = {{ .Region }}
output = {{ .Output }}
`

	AWSCredentialsTemplateFile = `[{{ .Profile }}]
aws_access_key_id = {{ .AwsCredentialsProfile.AccessKeyID }}
aws_secret_access_key = {{ .AwsCredentialsProfile.SecretAccessKey }}
{{- if .AwsCredentialsProfile.SessionToken }}
aws_session_token = {{ .AwsCredentialsProfile.SessionToken }}
{{- end }}
`
)

type AwsConfigure struct {
	*MFAToken             `yaml:"-"`
	AccountID             string `yaml:"account_id,omitempty"`
	AwsCredentialsProfile `yaml:"-"`
	ConfigSource          string          `yaml:"config-source"`
	CredentialsSource     string          `yaml:"credentials-source"`
	Ctx                   context.Context `yaml:"-"`
	IAMUserName           string          `yaml:"user-name,omitempty"`
	MFADeviceSerialNumber string          `yaml:"mfa-device,omitempty"`
	MFAProfileCredentials aws.Credentials `yaml:"-"`
	Output                string          `yaml:"output,omitempty"`
	Profile               string          `yaml:"profile,omitempty"`
	Region                string          `json:"aws-region,omitempty" yaml:"region,omitempty"`
}

type MFAToken struct {
	AccessKeyID     string
	Expiration      time.Time
	SecretAccessKey string
	SessionToken    string
}

type AwsCredentialsProfile struct {
	AccessKeyID     string `json:"aws-access-key-id,omitempty" yaml:"-"`
	SecretAccessKey string `json:"aws-secret-access-key,omitempty" yaml:"-"`
	SessionToken    string `yaml:"-"`
}

func (a *AwsConfigure) AWSSharedConfigFile(profile string) []string {
	return []string{config.DefaultSharedConfigFilename() + "_" + profile}
}

func (a *AwsConfigure) AWSSharedCredentialsFile(profile string) []string {
	return []string{config.DefaultSharedCredentialsFilename() + "_" + profile}
}

func NewAwsConfigure(ctx context.Context, profile string) *AwsConfigure {
	return &AwsConfigure{Ctx: ctx, Output: "text", Profile: profile}
}

func (a *AwsConfigure) errorProxy(cfg aws.Config, err error) (aws.Config, error) {
	for _, val := range cfg.ConfigSources {
		switch result := val.(type) {
		case config.SharedConfig:
			if len(result.Profile) == 0 && len(result.Region) == 0 {
				return aws.Config{}, fmt.Errorf("AWS profile by name %s does not exist, will be created", a.Profile)
			}
		}
	}

	return cfg, err
}

// configOptions - forming custom paths to AWS credentials and profile
func (a *AwsConfigure) configOptions() []func(options *config.LoadOptions) error {
	return []func(options *config.LoadOptions) error{
		config.WithSharedConfigFiles(a.AWSSharedConfigFile(a.Profile)),
		config.WithSharedCredentialsFiles(a.AWSSharedCredentialsFile(a.Profile)),
		config.WithSharedConfigProfile(a.Profile),
	}
}

func getTagStructName(i interface{}, name string) error {
	if field, ok := reflect.TypeOf(i).Elem().FieldByName(name); ok {
		return fmt.Errorf("profile option %s required", strings.TrimSuffix(field.Tag.Get("json"), ",omitempty"))
	} else {
		return fmt.Errorf("field with name %s not defined", name)
	}
}

// ValidateAWSCredentials will validate the required parameters for AWS authentication
func (a *AwsConfigure) ValidateAWSCredentials() error {
	if len(a.AwsCredentialsProfile.AccessKeyID) == 0 {
		return getTagStructName(&a.AwsCredentialsProfile, "AccessKeyID")
	}

	if len(a.AwsCredentialsProfile.SecretAccessKey) == 0 {
		return getTagStructName(&a.AwsCredentialsProfile, "SecretAccessKey")
	}

	if len(a.Region) == 0 {
		return getTagStructName(a, "Region")
	}

	return nil
}

// RenderAWSConfigProfile will render the AWS profile.
func (a *AwsConfigure) RenderAWSConfigProfile(temp string) ([]byte, error) {
	tmpl, err := template.New("AWS config Profile").Parse(temp)
	if err != nil {
		return nil, err
	}

	var credsFileStr bytes.Buffer
	err = tmpl.Execute(&credsFileStr, a)
	if err != nil {
		return nil, err
	}

	return credsFileStr.Bytes(), nil
}

// RenderBase64EncodedAWSConfigProfile will render the AWS profile, encoded in base 64.
func (a *AwsConfigure) RenderBase64EncodedAWSConfigProfile(temp string) (string, error) {
	configProfile, err := a.RenderAWSConfigProfile(temp)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(configProfile), nil
}

func (a *AwsConfigure) ReadAWSConfigProfile() error {
	cfg, err := config.LoadDefaultConfig(a.Ctx, a.configOptions()...)
	if err != nil {
		return err
	}

	for _, val := range cfg.ConfigSources {
		switch result := val.(type) {
		case config.SharedConfig:
			a.AwsCredentialsProfile.AccessKeyID = result.Credentials.AccessKeyID
			a.AwsCredentialsProfile.SecretAccessKey = result.Credentials.SecretAccessKey
			a.AwsCredentialsProfile.SessionToken = result.Credentials.SessionToken
			a.Region = result.Region
		}
	}

	return nil
}

func (a *AwsConfigure) WriteAWSConfigProfile() error {
	var (
		err         error
		sharedFiles = make(map[string][]byte)
	)

	sharedFiles[a.ConfigSource], err = a.RenderAWSConfigProfile(AWSConfigTemplateFile)
	if err != nil {
		return err
	}

	sharedFiles[a.CredentialsSource], err = a.RenderAWSConfigProfile(AWSCredentialsTemplateFile)
	if err != nil {
		return err
	}

	for key, val := range sharedFiles {
		if err := os.MkdirAll(filepath.Dir(key), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(key, val, 0644); err != nil {
			return err
		}
	}

	return nil
}

func (a *AwsConfigure) GetUserName() error {
	cfg, err := a.errorProxy(config.LoadDefaultConfig(a.Ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	user, err := iam.NewFromConfig(cfg).GetUser(a.Ctx, &iam.GetUserInput{})
	if err != nil {
		return err
	}

	a.IAMUserName = aws.ToString(user.User.UserName)

	return nil
}

func (a *AwsConfigure) GetMFACredentials() error {
	cfg, err := a.errorProxy(config.LoadDefaultConfig(a.Ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	if a.MFAProfileCredentials, err = cfg.Credentials.Retrieve(a.Ctx); err != nil {
		return err
	}

	return nil
}

func (a *AwsConfigure) GetMFADevicesSerialNumbers() error {
	var serialNumbers = make(map[string]string)

	if err := a.GetUserName(); err != nil {
		return err
	}

	if err := a.GetMFACredentials(); err != nil {
		return err
	}

	cfg, err := a.errorProxy(config.LoadDefaultConfig(a.Ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	mfaDevices, err := iam.NewFromConfig(cfg).ListMFADevices(a.Ctx,
		&iam.ListMFADevicesInput{UserName: aws.String(a.IAMUserName)})
	if err != nil {
		return err
	}

	if len(mfaDevices.MFADevices) > 1 {
		for key, val := range mfaDevices.MFADevices {
			fmt.Printf("%d. - MFA Device SerialNumber: %s\n", key+1, aws.ToString(val.SerialNumber))
			serialNumbers[strconv.Itoa(key+1)] = aws.ToString(val.SerialNumber)
		}

		if _, ok := serialNumbers[util.ReadStdin("number SerialNumber")]; ok {
			a.MFADeviceSerialNumber = serialNumbers[util.ReadStdin("number SerialNumber")]
		} else {
			return fmt.Errorf("incorrectly specified number SerialNumber")
		}
	} else if len(mfaDevices.MFADevices) == 1 {
		a.MFADeviceSerialNumber = aws.ToString(mfaDevices.MFADevices[0].SerialNumber)
	}

	return nil
}

func (a *AwsConfigure) GetMFASessionToken() error {
	cfg, err := a.errorProxy(config.LoadDefaultConfig(a.Ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	if err := a.GetMFADevicesSerialNumbers(); err != nil {
		return err
	}

	token, err := sts.NewFromConfig(cfg).GetSessionToken(a.Ctx, &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(43200),
		SerialNumber:    aws.String(a.MFADeviceSerialNumber),
		TokenCode:       aws.String(util.ReadStdin("TOTP")),
	})
	if err != nil {
		return err
	}

	a.MFAToken = &MFAToken{
		AccessKeyID:     aws.ToString(token.Credentials.AccessKeyId),
		Expiration:      aws.ToTime(token.Credentials.Expiration),
		SecretAccessKey: aws.ToString(token.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(token.Credentials.SessionToken),
	}

	return nil
}

func (a *AwsConfigure) GetAwsConfigure(profile string) (bool, error) {
	cfg, err := a.errorProxy(config.LoadDefaultConfig(a.Ctx,
		config.WithSharedConfigFiles(a.AWSSharedConfigFile(profile)),
		config.WithSharedCredentialsFiles(a.AWSSharedCredentialsFile(profile)),
		config.WithSharedConfigProfile(profile),
	))
	if err != nil {
		return true, err
	}

	client := sts.NewFromConfig(cfg)
	identity, err := client.GetCallerIdentity(a.Ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return false, err
	}

	a.Region = cfg.Region
	a.AccountID = aws.ToString(identity.Account)

	return true, nil
}

func (a *AwsConfigure) GetAWSClusterContext(clusterName string) ([]byte, error) {
	cfg, err := a.errorProxy(config.LoadDefaultConfig(a.Ctx, a.configOptions()...))
	if err != nil {
		return nil, err
	}

	client := eks.NewFromConfig(cfg)
	cluster, err := client.DescribeCluster(a.Ctx, &eks.DescribeClusterInput{Name: aws.String(clusterName)})
	if err != nil {
		return nil, err
	}

	return a.generateUserKubeconfig(cluster.Cluster)
}

func (a *AwsConfigure) generateUserKubeconfig(cluster *eksType.Cluster) ([]byte, error) {
	var execEnvVars []api.ExecEnvVar

	clusterName := aws.ToString(cluster.Name)
	userName := a.getKubeConfigUserName(clusterName)

	cfg, err := a.generateBaseKubeConfig(cluster)
	if err != nil {
		return nil, fmt.Errorf("creating base kubeconfig: %w", err)
	}

	execEnvVars = append(execEnvVars,
		api.ExecEnvVar{Name: "AWS_PROFILE", Value: a.Profile},
		api.ExecEnvVar{Name: "AWS_CONFIG_FILE", Value: strings.Join(a.AWSSharedConfigFile(a.Profile), "")},
		api.ExecEnvVar{Name: "AWS_SHARED_CREDENTIALS_FILE", Value: strings.Join(a.AWSSharedCredentialsFile(a.Profile), "")},
	)

	// Version v1alpha1 was removed in Kubernetes v1.23.
	// Version v1 was released in Kubernetes v1.23.
	// Version v1beta1 was selected as it has the widest range of support
	// This should be changed to v1 once EKS no longer supports Kubernetes <v1.23
	execConfig := &api.ExecConfig{
		APIVersion: "client.authentication.k8s.io/v1beta1",
		Args: []string{
			"token",
			"-i",
			clusterName,
		},
		Command: "aws-iam-authenticator",
		Env:     execEnvVars,
	}

	cfg.AuthInfos = map[string]*api.AuthInfo{
		userName: {
			Exec: execConfig,
		},
	}

	out, err := clientcmd.Write(*cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize kubeconfig to yaml: %w", err)
	}

	return out, nil
}

func (a *AwsConfigure) generateBaseKubeConfig(cluster *eksType.Cluster) (*api.Config, error) {
	clusterName := aws.ToString(cluster.Name)
	contextName := aws.ToString(cluster.Name)
	userName := a.getKubeConfigUserName(clusterName)

	certData, err := base64.StdEncoding.DecodeString(aws.ToString(cluster.CertificateAuthority.Data))
	if err != nil {
		return nil, fmt.Errorf("decoding cluster CA cert: %w", err)
	}

	cfg := &api.Config{
		APIVersion: api.SchemeGroupVersion.Version,
		Clusters: map[string]*api.Cluster{
			clusterName: {
				Server:                   aws.ToString(cluster.Endpoint),
				CertificateAuthorityData: certData,
			},
		},
		Contexts: map[string]*api.Context{
			contextName: {
				Cluster:  clusterName,
				AuthInfo: userName,
			},
		},
		CurrentContext: contextName,
	}

	return cfg, nil
}

func (a *AwsConfigure) getKubeConfigUserName(clusterName string) string {
	return fmt.Sprintf("%s-capi-admin", clusterName)
}

func (a *AwsConfigure) CreateEC2SSHKey(clusterName string) error {
	var respError smithy.APIError

	cfg, err := a.errorProxy(config.LoadDefaultConfig(a.Ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	params := &ec2.CreateKeyPairInput{
		KeyName: aws.String(clusterName),
		DryRun:  aws.Bool(false),
		KeyType: ec2Type.KeyTypeEd25519,
		TagSpecifications: append([]ec2Type.TagSpecification{}, ec2Type.TagSpecification{
			ResourceType: ec2Type.ResourceTypeKeyPair,
			Tags: append([]ec2Type.Tag{}, ec2Type.Tag{
				Key:   aws.String("kubernetes.io/cluster/" + clusterName),
				Value: aws.String("owned"),
			}),
		}),
	}

	client := ec2.NewFromConfig(cfg)
	sshKey, err := client.CreateKeyPair(a.Ctx, params)
	if err != nil {
		if errors.As(err, &respError) && respError.ErrorCode() != "InvalidKeyPair.Duplicate" {
			return err
		}
	}

	if sshKey != nil {
		zap.S().Infof("created SSHKey %s with id %s", aws.ToString(sshKey.KeyName), aws.ToString(sshKey.KeyPairId))
	}

	return nil
}

func (a *AwsConfigure) DeleteEC2SSHKey(clusterName string) error {
	cfg, err := a.errorProxy(config.LoadDefaultConfig(a.Ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	params := &ec2.DeleteKeyPairInput{
		KeyName: aws.String(clusterName),
		DryRun:  aws.Bool(false),
	}

	client := ec2.NewFromConfig(cfg)
	sshKey, err := client.DeleteKeyPair(a.Ctx, params)
	if err != nil {
		return err
	}

	if sshKey.KeyPairId != nil {
		zap.S().Infof("deleted SSHKey %s with id %s", clusterName, aws.ToString(sshKey.KeyPairId))
	}

	return nil
}

func (a *AwsConfigure) CreateBucket(bucketName string) error {
	var (
		respError      s3.ResponseError
		bucketExist    *s3type.BucketAlreadyExists
		bucketOwner    *s3type.BucketAlreadyOwnedByYou
		bucketParams   s3.CreateBucketInput
		bucketNotFound *s3type.NotFound
	)

	ctx := context.TODO()
	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(cfg)

	if a.Region == util.RegionException {
		_, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucketName)})
		if err != nil {
			if !errors.As(err, &bucketNotFound) {
				return err
			}
		} else {
			zap.S().Infof("S3 bucket %s already exists", bucketName)
			return nil
		}

		bucketParams = s3.CreateBucketInput{Bucket: aws.String(bucketName)}
	} else {
		bucketParams = s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
			CreateBucketConfiguration: &s3type.CreateBucketConfiguration{
				LocationConstraint: s3type.BucketLocationConstraint(a.Region),
			},
		}
	}

	resp, err := client.CreateBucket(ctx, &bucketParams)
	if err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &bucketExist) {
			zap.S().Infof("S3 bucket %s already exists", bucketName)
			return nil
		} else if errors.As(err, &bucketOwner) {
			zap.S().Infof("S3 bucket %s already exists and owned by you", bucketName)
			return nil
		} else if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusForbidden {
			zap.S().Warnf("S3 bucket %s is not created, you don't have permissions", bucketName)
			return nil
		} else if errors.As(err, &respError) {
			return fmt.Errorf("requestID %s, hostID %s request failure by error: %v", respError.ServiceRequestID(),
				respError.ServiceHostID(), respError.Error())
		}

		return err
	}

	putParams := s3.PutPublicAccessBlockInput{
		Bucket: aws.String(bucketName),
		PublicAccessBlockConfiguration: &s3type.PublicAccessBlockConfiguration{
			BlockPublicAcls:       aws.Bool(true),
			BlockPublicPolicy:     aws.Bool(true),
			IgnorePublicAcls:      aws.Bool(true),
			RestrictPublicBuckets: aws.Bool(true),
		},
	}

	if _, err := client.PutPublicAccessBlock(ctx, &putParams); err != nil {
		if errors.As(err, &respError) {
			return fmt.Errorf("requestID %s, hostID %s request failure by error: %v",
				respError.ServiceRequestID(), respError.ServiceHostID(), respError.Error())
		}

		return err
	}

	zap.S().Infof("created S3 bucket: %s - %s", bucketName, *resp.Location)

	return nil
}

func (a *AwsConfigure) BucketKeyExists(region, bucketName, key string) (bool, error) {
	if len(bucketName) == 0 {
		return false, nil
	}

	ctx := context.TODO()
	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
	if err != nil {
		return false, err
	}

	// needed for specific AWS account where S3 used
	if len(region) > 0 {
		cfg.Region = region
	}

	client := s3.NewFromConfig(cfg)
	_, err = client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// S3ListObjectsAPI defines the interface for the ListObjectsV2 function.
// We use this interface to test the function using a mocked service.
type S3ListObjectsAPI interface {
	ListObjectsV2(ctx context.Context,
		params *s3.ListObjectsV2Input,
		optFns ...func(*s3.Options)) (*s3.ListObjectsV2Output, error)
}

// GetObjects retrieves the objects in an Amazon Simple Storage Service (Amazon S3) bucket
// Inputs:
//
//	c is the context of the method call, which includes the AWS Region
//	api is the interface that defines the method call
//	input defines the input arguments to the service call.
//
// Output:
//
//	If success, a ListObjectsV2Output object containing the result of the service call and nil
//	Otherwise, nil and an error from the call to ListObjectsV2
func GetObjects(c context.Context, api S3ListObjectsAPI, input *s3.ListObjectsV2Input) (*s3.ListObjectsV2Output, error) {
	return api.ListObjectsV2(c, input)
}

func (a *AwsConfigure) DownloadFromBucket(region, bucketName, localDir, filePrefix string) error {
	var noSuchBucket *s3type.NoSuchBucket

	downloadToFile := func(downloader *manager.Downloader, targetDirectory, bucket, key string) error {
		// Create the directories in the path
		file := filepath.Join(targetDirectory, key)
		if err := os.MkdirAll(filepath.Dir(file), 0775); err != nil {
			return err
		}

		fd, err := os.Create(file)
		if err != nil {
			return err
		}
		defer func(fd *os.File) {
			err := fd.Close()
			if err != nil {

			}
		}(fd)

		// Download the file using the AWS SDK for Go
		zap.S().Infof("downloading s3://%s/%s to %s", bucket, key, file)
		_, err = downloader.Download(context.TODO(), fd, &s3.GetObjectInput{Bucket: &bucket, Key: &key})

		return err
	}

	ctx := context.TODO()
	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	// needed for specific AWS account where S3 used
	if len(region) > 0 {
		cfg.Region = region
	}

	client := s3.NewFromConfig(cfg)
	m := manager.NewDownloader(client)
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{Bucket: aws.String(bucketName)})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			if errors.As(err, &noSuchBucket) {
				return fmt.Errorf("specified bucket %s does not exist", bucketName)
			}

			return err
		}

		for _, obj := range page.Contents {
			if strings.HasPrefix(aws.ToString(obj.Key), filePrefix) {
				if err := downloadToFile(m, localDir, bucketName, aws.ToString(obj.Key)); err != nil {
					return err
				}
			}
		}
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	}

	resp, err := GetObjects(context.TODO(), client, input)
	if err != nil {
		return err
	}

	if len(resp.Contents) == 0 {
		zap.S().Warnf("S3 bucket %s is empty, files do not exist", bucketName)
		return nil
	}

	return nil
}

func (a *AwsConfigure) GetFileData(bucketName, key string) ([]byte, error) {
	var client *s3.Client
	ctx := context.TODO()
	if len(a.Profile) > 0 {
		cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
		if err != nil {
			return nil, err
		}

		client = s3.NewFromConfig(cfg)
	} else {
		client = s3.NewFromConfig(aws.Config{Region: a.Region})
	}

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(key),
	}

	resp, err := client.GetObject(ctx, input)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(resp.Body)
}

func (a *AwsConfigure) UploadToBucket(bucketName, localDir, pattern string) error {
	ctx := context.TODO()
	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	uploader := manager.NewUploader(s3.NewFromConfig(cfg))

	match, err := util.WalkMatch(localDir, pattern)
	if err != nil {
		return err
	}

	for _, path := range match {
		if filepath.Base(path) != util.SopsAgeKeyFile {
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			result, err := uploader.Upload(context.TODO(), &s3.PutObjectInput{
				Bucket: aws.String(bucketName),
				Key:    aws.String(filepath.Base(path)),
				Body:   bytes.NewReader(data),
			})
			if err != nil {
				var responseError *awshttp.ResponseError
				if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
					return fmt.Errorf("specified bucket %s does not exist", bucketName)
				}

				if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusForbidden {
					return fmt.Errorf("you are not permitted to upload SOPS age keys for specified bucket %s, "+
						"access denied", bucketName)
				}

				return err
			}

			zap.S().Infof("uploading %s... to %s", path, result.Location)
		}
	}

	return nil
}
