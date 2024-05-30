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
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	ddbtype "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3type "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"go.uber.org/zap"

	"rmk/system"
)

type AwsConfigure struct {
	Profile               string `yaml:"profile,omitempty"`
	Region                string `yaml:"region,omitempty"`
	AccountID             string `yaml:"account_id,omitempty"`
	UserName              string `yaml:"user_name,omitempty"`
	MFADeviceSerialNumber string `yaml:"mfa_device,omitempty"`
	*MFAToken             `yaml:"-"`
	MFAProfileCredentials aws.Credentials `yaml:"-"`
}

type MFAToken struct {
	AccessKeyId     string
	Expiration      time.Time
	SecretAccessKey string
	SessionToken    string
}

func (a *AwsConfigure) AWSSharedConfigFile(profile string) []string {
	return []string{system.GetHomePath(".aws", "config_"+profile)}
}

func (a *AwsConfigure) AWSSharedCredentialsFile(profile string) []string {
	return []string{system.GetHomePath(".aws", "credentials_"+profile)}
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

func (a *AwsConfigure) GetUserName() error {
	ctx := context.TODO()
	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	user, err := iam.NewFromConfig(cfg).GetUser(ctx, &iam.GetUserInput{})
	if err != nil {
		return err
	}

	a.UserName = aws.ToString(user.User.UserName)

	return nil
}

func (a *AwsConfigure) GetAWSCredentials() error {
	ctx := context.TODO()
	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	if a.MFAProfileCredentials, err = cfg.Credentials.Retrieve(ctx); err != nil {
		return err
	}

	return nil
}

func (a *AwsConfigure) GetMFADevicesSerialNumbers() error {
	var serialNumbers = make(map[string]string)

	if err := a.GetUserName(); err != nil {
		return err
	}

	if err := a.GetAWSCredentials(); err != nil {
		return err
	}

	ctx := context.TODO()
	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	mfaDevices, err := iam.NewFromConfig(cfg).ListMFADevices(ctx,
		&iam.ListMFADevicesInput{UserName: aws.String(a.UserName)})
	if err != nil {
		return err
	}

	if len(mfaDevices.MFADevices) > 1 {
		for key, val := range mfaDevices.MFADevices {
			fmt.Printf("%d. - MFA Device SerialNumber: %s\n", key+1, aws.ToString(val.SerialNumber))
			serialNumbers[strconv.Itoa(key+1)] = aws.ToString(val.SerialNumber)
		}

		if _, ok := serialNumbers[system.ReadStdin("number SerialNumber")]; ok {
			a.MFADeviceSerialNumber = serialNumbers[system.ReadStdin("number SerialNumber")]
		} else {
			return fmt.Errorf("incorrectly specified number SerialNumber")
		}
	} else if len(mfaDevices.MFADevices) == 1 {
		a.MFADeviceSerialNumber = aws.ToString(mfaDevices.MFADevices[0].SerialNumber)
	}

	return nil
}

func (a *AwsConfigure) GetMFASessionToken() error {
	ctx := context.TODO()
	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	if err := a.GetMFADevicesSerialNumbers(); err != nil {
		return err
	}

	token, err := sts.NewFromConfig(cfg).GetSessionToken(ctx, &sts.GetSessionTokenInput{
		DurationSeconds: aws.Int32(43200),
		SerialNumber:    aws.String(a.MFADeviceSerialNumber),
		TokenCode:       aws.String(system.ReadStdin("TOTP")),
	})
	if err != nil {
		return err
	}

	a.MFAToken = &MFAToken{
		AccessKeyId:     aws.ToString(token.Credentials.AccessKeyId),
		Expiration:      aws.ToTime(token.Credentials.Expiration),
		SecretAccessKey: aws.ToString(token.Credentials.SecretAccessKey),
		SessionToken:    aws.ToString(token.Credentials.SessionToken),
	}

	return nil
}

func (a *AwsConfigure) GetAwsConfigure(profile string) (bool, error) {
	ctx := context.TODO()
	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx,
		config.WithSharedConfigFiles(a.AWSSharedConfigFile(profile)),
		config.WithSharedCredentialsFiles(a.AWSSharedCredentialsFile(profile)),
		config.WithSharedConfigProfile(profile),
	))
	if err != nil {
		return true, err
	}

	client := sts.NewFromConfig(cfg)
	identity, err := client.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return false, err
	}

	a.Region = cfg.Region
	a.AccountID = aws.ToString(identity.Account)

	return true, nil
}

func (a *AwsConfigure) GetECRCredentials(region string) (map[string]string, error) {
	ctx := context.TODO()
	ecrCredentials := make(map[string]string)

	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
	if err != nil {
		return nil, err
	}

	// needed for specific AWS account where ECR used
	cfg.Region = region

	svc := ecr.NewFromConfig(cfg)
	token, err := svc.GetAuthorizationToken(ctx, &ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return nil, err
	}

	authData := token.AuthorizationData[0].AuthorizationToken
	data, err := base64.StdEncoding.DecodeString(*authData)
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(string(data), ":", 2)
	if len(parts) < 2 {
		return nil, fmt.Errorf("it is impossible to get ECR user and password "+
			"for current AWS profile: %s", a.Profile)
	}

	ecrCredentials[parts[0]] = parts[1]

	return ecrCredentials, nil
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

	if a.Region == system.RegionException {
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

func (a *AwsConfigure) CreateDynamoDBTable(tableName string) error {
	ctx := context.TODO()
	cfg, err := a.errorProxy(config.LoadDefaultConfig(ctx, a.configOptions()...))
	if err != nil {
		return err
	}

	client := dynamodb.NewFromConfig(cfg)
	dTableParams := &dynamodb.CreateTableInput{
		AttributeDefinitions: []ddbtype.AttributeDefinition{
			{
				AttributeName: aws.String("LockID"),
				AttributeType: ddbtype.ScalarAttributeTypeS,
			},
		},
		KeySchema: []ddbtype.KeySchemaElement{
			{
				AttributeName: aws.String("LockID"),
				KeyType:       ddbtype.KeyTypeHash,
			},
		},
		TableName:   aws.String(tableName),
		BillingMode: ddbtype.BillingModePayPerRequest,
	}

	resp, err := client.CreateTable(ctx, dTableParams)
	if err != nil {
		var (
			tableExist   *ddbtype.ResourceInUseException
			accessDenied smithy.APIError
			operation    *smithy.OperationError
		)

		if errors.As(err, &tableExist) {
			zap.S().Infof("DynamoDB table %s already exists", tableName)
			return nil
		} else if errors.As(err, &accessDenied) && errors.As(err, &operation) &&
			operation.Operation() == "CreateTable" && accessDenied.ErrorCode() == "AccessDeniedException" {
			zap.S().Warnf("DynamoDB table %s is not created, you don't have permissions", tableName)
			return nil
		}

		return err
	}

	zap.S().Infof("created DynamoDB table: %s - %s", tableName, *resp.TableDescription.TableArn)

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

	match, err := system.WalkMatch(localDir, pattern)
	if err != nil {
		return err
	}

	for _, path := range match {
		if filepath.Base(path) != system.SopsAgeKeyFile {
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
