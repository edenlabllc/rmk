package cmd

import (
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"

	"rmk/providers/aws_provider"
	"rmk/util"
)

const (
	awsFlagsCategory = "AWS authentication"

	awsIAMControllerCredentialsTemplate = `[default]
aws_access_key_id = {{ .AwsCredentialsProfile.AccessKeyID }}
aws_secret_access_key = {{ .AwsCredentialsProfile.SecretAccessKey }}
region = {{ .Region }}
{{- if .AwsCredentialsProfile.SessionToken }}
aws_session_token = {{ .AwsCredentialsProfile.SessionToken }}
{{- end }}
`
	awsIAMControllerSecret            = "aws-iam-controller-secret"
	awsClusterStaticIdentityName      = "aws-cluster-identity"
	awsClusterStaticIdentityNamespace = "capa-system"
	awsClusterStaticIdentitySecret    = "aws-cluster-identity-secret"
)

var awsClusterStaticIdentitySecretType = corev1.SecretTypeOpaque

type AWSClusterStaticIdentityConfig struct {
	*AWSClusterStaticIdentity
	*v1.SecretApplyConfiguration
	AWSIAMControllerSecret *v1.SecretApplyConfiguration
	ManifestFiles          []string
	ManifestFilesDir       string
}

type AWSClusterStaticIdentity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AWSClusterStaticIdentitySpec `json:"spec,omitempty"`
}

type AWSClusterStaticIdentitySpec struct {
	AllowedNamespaces struct {
		NamespaceList []string              `json:"list"`
		Selector      *metav1.LabelSelector `json:"selector,omitempty"`
	} `json:"allowedNamespaces,omitempty"`
	SecretRef string `json:"secretRef"`
}

func NewAWSClusterStaticIdentityConfig(ac *aws_provider.AwsConfigure) *AWSClusterStaticIdentityConfig {
	acic := &AWSClusterStaticIdentityConfig{
		AWSClusterStaticIdentity: &AWSClusterStaticIdentity{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AWSClusterStaticIdentity",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta2",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      awsClusterStaticIdentityName,
				Namespace: awsClusterStaticIdentityNamespace,
				Labels:    map[string]string{"clusterctl.cluster.x-k8s.io/move-hierarchy": "true"},
			},
			Spec: AWSClusterStaticIdentitySpec{
				AllowedNamespaces: struct {
					NamespaceList []string              `json:"list"`
					Selector      *metav1.LabelSelector `json:"selector,omitempty"`
				}(struct {
					NamespaceList []string
					Selector      *metav1.LabelSelector
				}{
					NamespaceList: []string{awsClusterStaticIdentityNamespace},
				}),
				SecretRef: awsClusterStaticIdentitySecret,
			},
		},
		AWSIAMControllerSecret:   v1.Secret(awsIAMControllerSecret, awsClusterStaticIdentityNamespace),
		SecretApplyConfiguration: v1.Secret(awsClusterStaticIdentitySecret, awsClusterStaticIdentityNamespace),
		ManifestFilesDir:         filepath.Join(os.TempDir(), awsClusterStaticIdentityName),
	}

	profile, err := ac.RenderAWSConfigProfile(awsIAMControllerCredentialsTemplate)
	if err != nil {
		return nil
	}

	acic.AWSIAMControllerSecret.Type = &awsClusterStaticIdentitySecretType
	acic.AWSIAMControllerSecret.Data = map[string][]byte{"credentials": profile}

	acic.SecretApplyConfiguration.Type = &awsClusterStaticIdentitySecretType
	acic.SecretApplyConfiguration.Data = map[string][]byte{
		"AccessKeyID":     []byte(ac.AwsCredentialsProfile.AccessKeyID),
		"SecretAccessKey": []byte(ac.AwsCredentialsProfile.SecretAccessKey),
	}

	if len(ac.AwsCredentialsProfile.SessionToken) > 0 {
		acic.SecretApplyConfiguration.Data["SessionToken"] = []byte(ac.AwsCredentialsProfile.SessionToken)
	}

	return acic
}

func (acic *AWSClusterStaticIdentityConfig) createAWSClusterIdentityManifestFiles() error {
	if err := os.MkdirAll(acic.ManifestFilesDir, 0775); err != nil {
		return err
	}

	fileCR, err := createManifestFile(acic.AWSClusterStaticIdentity, acic.ManifestFilesDir, awsClusterStaticIdentityName)
	if err != nil {
		return err
	}

	acic.ManifestFiles = append(acic.ManifestFiles, fileCR)

	fileCRSecret, err := createManifestFile(acic.SecretApplyConfiguration, acic.ManifestFilesDir, awsClusterStaticIdentitySecret)
	if err != nil {
		return err
	}

	acic.ManifestFiles = append(acic.ManifestFiles, fileCRSecret)

	fileIAMControllerSecret, err := createManifestFile(acic.AWSIAMControllerSecret, acic.ManifestFilesDir, awsIAMControllerSecret)
	if err != nil {
		return err
	}

	acic.ManifestFiles = append(acic.ManifestFiles, fileIAMControllerSecret)

	return nil
}

func (cc *ClusterCommands) applyAWSClusterIdentity() error {
	var kubectlArgs = []string{"apply"}

	ac := aws_provider.NewAwsConfigure(cc.Ctx.Context, cc.Conf.Profile)
	if err := ac.ReadAWSConfigProfile(); err != nil {
		return err
	}

	acic := NewAWSClusterStaticIdentityConfig(ac)
	if err := acic.createAWSClusterIdentityManifestFiles(); err != nil {
		return err
	}

	for _, val := range acic.ManifestFiles {
		kubectlArgs = append(kubectlArgs, "-f", val)
	}

	cc.SpecCMD = cc.kubectl(kubectlArgs...)
	if err := releaseRunner(cc).runCMD(); err != nil {
		if err := os.RemoveAll(acic.ManifestFilesDir); err != nil {
			return err
		}

		return err
	}

	return os.RemoveAll(acic.ManifestFilesDir)
}

func (cc *ClusterCommands) getAWSClusterContext() ([]byte, error) {
	return aws_provider.NewAwsConfigure(cc.Ctx.Context, cc.Conf.Profile).GetAWSClusterContext(cc.Conf.Name)
}

func (cc *ClusterCommands) createAWSClusterSSHKey() error {
	return aws_provider.NewAwsConfigure(cc.Ctx.Context, cc.Conf.Profile).CreateAWSEC2SSHKey(cc.Conf.Name)
}

func (cc *ClusterCommands) deleteAWSClusterSSHKey() error {
	return aws_provider.NewAwsConfigure(cc.Ctx.Context, cc.Conf.Profile).DeleteAWSEC2SSHKey(cc.Conf.Name)
}

func (cc *ClusterCommands) createAWSSecrets() error {
	a := aws_provider.NewAwsConfigure(cc.Ctx.Context, cc.Conf.Profile)

	secrets, err := a.GetAWSSecrets(cc.Conf.Tenant)
	if err != nil {
		return err
	}

	if len(secrets) > 0 || !util.IsExists(cc.Conf.SopsAgeKeys, false) {
		return nil
	}

	walkMatch, err := util.WalkMatch(cc.Conf.SopsAgeKeys, cc.Conf.Tenant+"*"+util.SopsAgeKeyExt)
	if err != nil {
		return err
	}

	for _, val := range walkMatch {
		file, err := os.ReadFile(val)
		if err != nil {
			return err
		}

		keyName := strings.TrimSuffix(filepath.Base(val), util.SopsAgeKeyExt)

		if err := a.SetAWSSecret(cc.Conf.Tenant, keyName, file); err != nil {
			return err
		}
	}

	return nil
}
