package cmd

import (
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"

	"rmk/providers/azure_provider"
	"rmk/util"
)

const (
	azureClusterIdentityNamespacePrefix = "capz"
	azureFlagsCategory                  = "Azure authentication"
)

var azureClusterIdentitySecretType = corev1.SecretTypeOpaque

type AzureClusterIdentityConfig struct {
	*AzureClusterIdentity
	*v1.NamespaceApplyConfiguration
	*v1.SecretApplyConfiguration
	ManifestFilesDir string
	ManifestFiles    []string
}

type AzureClusterIdentity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec AzureClusterIdentitySpec `json:"spec,omitempty"`
}

type AzureClusterIdentitySpec struct {
	AllowedNamespaces struct {
		NamespaceList []string              `json:"list"`
		Selector      *metav1.LabelSelector `json:"selector,omitempty"`
	} `json:"allowedNamespaces,omitempty"`
	ClientID     string                 `json:"clientID"`
	ClientSecret corev1.SecretReference `json:"clientSecret,omitempty"`
	TenantID     string                 `json:"tenantID"`
	Type         string                 `json:"type"`
}

func NewAzureClusterIdentityConfig(ac *azure_provider.AzureConfigure, identity *IdentityName) *AzureClusterIdentityConfig {
	acic := &AzureClusterIdentityConfig{
		AzureClusterIdentity: &AzureClusterIdentity{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AzureClusterIdentity",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      identity.generateObjectName(objectIdentityName),
				Namespace: identity.generateObjectName(objectNamespace),
				Labels:    map[string]string{"clusterctl.cluster.x-k8s.io/move-hierarchy": "true"},
			},
			Spec: AzureClusterIdentitySpec{
				AllowedNamespaces: struct {
					NamespaceList []string              `json:"list"`
					Selector      *metav1.LabelSelector `json:"selector,omitempty"`
				}(struct {
					NamespaceList []string
					Selector      *metav1.LabelSelector
				}{
					NamespaceList: []string{identity.generateObjectName(objectNamespace)},
				}),
				ClientID: ac.ClientID,
				ClientSecret: corev1.SecretReference{
					Name:      identity.generateObjectName(objectIdentitySecret),
					Namespace: identity.generateObjectName(objectNamespace),
				},
				TenantID: ac.TenantID,
				Type:     "ServicePrincipal",
			},
		},
		NamespaceApplyConfiguration: v1.Namespace(identity.generateObjectName(objectNamespace)),
		SecretApplyConfiguration: v1.Secret(identity.generateObjectName(objectIdentitySecret),
			identity.generateObjectName(objectNamespace)),
		ManifestFilesDir: filepath.Join(os.TempDir(), identity.generateObjectName(objectIdentityName)),
	}

	acic.SecretApplyConfiguration.Type = &azureClusterIdentitySecretType
	acic.SecretApplyConfiguration.Data = map[string][]byte{"clientSecret": []byte(ac.ClientSecret)}

	return acic
}

func (acic *AzureClusterIdentityConfig) createAzureClusterIdentityManifestFiles(identity *IdentityName) error {
	if err := os.MkdirAll(acic.ManifestFilesDir, 0775); err != nil {
		return err
	}

	fileCR, err := createManifestFile(acic.AzureClusterIdentity, acic.ManifestFilesDir,
		identity.generateObjectName(objectIdentityName))
	if err != nil {
		return err
	}

	fileNamespace, err := createManifestFile(acic.NamespaceApplyConfiguration, acic.ManifestFilesDir,
		identity.generateObjectName(objectNamespace))
	if err != nil {
		return err
	}

	fileSecret, err := createManifestFile(acic.SecretApplyConfiguration, acic.ManifestFilesDir,
		identity.generateObjectName(objectIdentitySecret))
	if err != nil {
		return err
	}

	acic.ManifestFiles = append(acic.ManifestFiles, fileNamespace, fileCR, fileSecret)

	return nil
}

func (cc *ClusterCommands) applyAzureClusterIdentity() error {
	var kubectlArgs = []string{"apply"}

	ac := azure_provider.NewAzureConfigure()
	if err := ac.ReadSPCredentials(cc.Conf.Name); err != nil {
		return err
	}

	identityName := newIdentityName(map[string]IdentityNameSpec{
		objectIdentityName: {
			ClusterName: cc.Conf.Name,
			Suffix:      clusterIdentityNameSuffix,
		},
		objectIdentitySecret: {
			ClusterName: cc.Conf.Name,
			Suffix:      clusterIdentitySecretSuffix,
		},
		objectNamespace: {
			ClusterName: cc.Conf.Name,
			Prefix:      azureClusterIdentityNamespacePrefix,
		},
	})

	acic := NewAzureClusterIdentityConfig(ac, identityName)
	if err := acic.createAzureClusterIdentityManifestFiles(identityName); err != nil {
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

func (cc *ClusterCommands) getAzureClusterContext() ([]byte, error) {
	ac := azure_provider.NewAzureConfigure()

	if err := ac.NewAzureClient(cc.Ctx.Context, cc.Conf.Name); err != nil {
		return nil, err
	}

	return ac.GetAzureClusterContext(cc.Conf.Tenant, cc.Conf.Name)
}

func (cc *ClusterCommands) createAzureSecrets(ac *azure_provider.AzureConfigure) error {
	if err := ac.NewAzureClient(cc.Ctx.Context, cc.Conf.Name); err != nil {
		return err
	}

	secrets, err := ac.GetAzureSecrets()
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
		value := string(file)

		if err := ac.SetAzureSecret(keyName, value); err != nil {
			return err
		}
	}

	return nil
}
