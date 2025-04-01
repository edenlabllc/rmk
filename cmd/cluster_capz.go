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
	azureClusterIdentityName      = "azure-cluster-identity"
	azureClusterIdentityNamespace = "capz-system"
	azureClusterIdentitySecret    = "azure-cluster-identity-secret"
	azureFlagsCategory            = "Azure authentication"
)

var azureClusterIdentitySecretType = corev1.SecretTypeOpaque

type AzureClusterIdentityConfig struct {
	*AzureClusterIdentity
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

func NewAzureClusterIdentityConfig(ac *azure_provider.AzureConfigure) *AzureClusterIdentityConfig {
	acic := &AzureClusterIdentityConfig{
		AzureClusterIdentity: &AzureClusterIdentity{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AzureClusterIdentity",
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      azureClusterIdentityName,
				Namespace: azureClusterIdentityNamespace,
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
					NamespaceList: []string{azureClusterIdentityNamespace},
					Selector: &metav1.LabelSelector{
						MatchExpressions: []metav1.LabelSelectorRequirement{
							metav1.LabelSelectorRequirement{
								Key:      "name",
								Operator: metav1.LabelSelectorOpExists,
								Values:   nil,
							},
						},
					},
				}),
				ClientID: ac.ClientID,
				ClientSecret: corev1.SecretReference{
					Name:      azureClusterIdentitySecret,
					Namespace: azureClusterIdentityNamespace,
				},
				TenantID: ac.TenantID,
				Type:     "ServicePrincipal",
			},
		},
		SecretApplyConfiguration: v1.Secret(azureClusterIdentitySecret, azureClusterIdentityNamespace),
		ManifestFilesDir:         filepath.Join(os.TempDir(), azureClusterIdentityName),
	}

	acic.SecretApplyConfiguration.Type = &azureClusterIdentitySecretType
	acic.SecretApplyConfiguration.Data = map[string][]byte{"clientSecret": []byte(ac.ClientSecret)}

	return acic
}

func (acic *AzureClusterIdentityConfig) createAzureClusterIdentityManifestFiles() error {
	if err := os.MkdirAll(acic.ManifestFilesDir, 0775); err != nil {
		return err
	}

	fileCR, err := createManifestFile(acic.AzureClusterIdentity, acic.ManifestFilesDir, azureClusterIdentityName)
	if err != nil {
		return err
	}

	acic.ManifestFiles = append(acic.ManifestFiles, fileCR)

	fileSecret, err := createManifestFile(acic.SecretApplyConfiguration, acic.ManifestFilesDir, azureClusterIdentitySecret)
	if err != nil {
		return err
	}

	acic.ManifestFiles = append(acic.ManifestFiles, fileSecret)

	return nil
}

func (cc *ClusterCommands) applyAzureClusterIdentity() error {
	var kubectlArgs = []string{"apply"}

	ac := azure_provider.NewAzureConfigure()
	if err := ac.ReadSPCredentials(cc.Conf.Name); err != nil {
		return err
	}

	acic := NewAzureClusterIdentityConfig(ac)
	if err := acic.createAzureClusterIdentityManifestFiles(); err != nil {
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
