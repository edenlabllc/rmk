package cmd

import (
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"

	"rmk/providers/onprem_provider"
)

const (
	onPremClusterIdentityNamespacePrefix = "capop"
	onPremFlagsCategory                  = "On-Premise SSH authentication"
)

var onPremClusterIdentitySecretType = corev1.SecretTypeOpaque

type OnPremClusterIdentityConfig struct {
	*v1.NamespaceApplyConfiguration
	*v1.SecretApplyConfiguration
	ManifestFiles    []string
	ManifestFilesDir string
}

func NewOnPremClusterIdentityConfig(onPrem *onprem_provider.OnPremConfigure, identity *IdentityName) *OnPremClusterIdentityConfig {
	opcc := &OnPremClusterIdentityConfig{
		NamespaceApplyConfiguration: v1.Namespace(identity.generateObjectName(objectNamespace)),
		SecretApplyConfiguration: v1.Secret(identity.generateObjectName(objectIdentitySecret),
			identity.generateObjectName(objectNamespace)),
		ManifestFilesDir: filepath.Join(os.TempDir(), identity.generateObjectName(objectIdentityName)),
	}

	opcc.SecretApplyConfiguration.Type = &onPremClusterIdentitySecretType
	opcc.SecretApplyConfiguration.Data = map[string][]byte{"private_key": onPrem.GetSSHPrivateKeyContent()}

	return opcc
}

func (opcic *OnPremClusterIdentityConfig) createOnPremClusterIdentitySecretManifestFiles(identity *IdentityName) error {
	if err := os.MkdirAll(opcic.ManifestFilesDir, 0775); err != nil {
		return err
	}

	fileNamespace, err := createManifestFile(opcic.NamespaceApplyConfiguration, opcic.ManifestFilesDir,
		identity.generateObjectName(objectNamespace))
	if err != nil {
		return err
	}

	fileSecret, err := createManifestFile(opcic.SecretApplyConfiguration, opcic.ManifestFilesDir,
		identity.generateObjectName(objectIdentitySecret))
	if err != nil {
		return err
	}

	opcic.ManifestFiles = append(opcic.ManifestFiles, fileNamespace, fileSecret)

	return nil
}

func (cc *ClusterCommands) applyOnPremClusterIdentitySecret() error {
	var kubectlArgs = []string{"apply"}

	identityName := newIdentityName(map[string]IdentityNameSpec{
		objectIdentityName: {
			ClusterName: cc.Conf.Name,
			Suffix:      clusterIdentityNameSuffix,
		},
		objectIdentitySecret: {
			ClusterName: onPremClusterIdentityNamespacePrefix + "-ssh",
			Suffix:      clusterIdentitySecretSuffix,
		},
		objectNamespace: {
			ClusterName: "system",
			Prefix:      onPremClusterIdentityNamespacePrefix,
		},
	})

	opcic := NewOnPremClusterIdentityConfig(cc.Conf.OnPremConfigure, identityName)
	if err := opcic.createOnPremClusterIdentitySecretManifestFiles(identityName); err != nil {
		return err
	}

	for _, val := range opcic.ManifestFiles {
		kubectlArgs = append(kubectlArgs, "-f", val)
	}

	cc.SpecCMD = cc.kubectl(kubectlArgs...)
	if err := releaseRunner(cc).runCMD(); err != nil {
		if err := os.RemoveAll(opcic.ManifestFilesDir); err != nil {
			return err
		}

		return err
	}

	return os.RemoveAll(opcic.ManifestFilesDir)
}

func (cc *ClusterCommands) getOnPremClusterContext() ([]byte, error) {
	auth, err := cc.Conf.OnPremConfigure.SSHAuth(false, false)
	if err != nil {
		return nil, err
	}

	return cc.Conf.OnPremConfigure.GetOnPremClusterContext(cc.Conf.Name, auth)
}
