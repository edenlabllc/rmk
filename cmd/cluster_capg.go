package cmd

import (
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"

	"rmk/providers/google_provider"
)

const (
	gcpClusterIdentityName      = "gcp-cluster-identity"
	gcpClusterIdentityNamespace = "capg-system"
	gcpClusterIdentitySecret    = "gcp-cluster-identity-secret"
	gcpFlagsCategory            = "GCP authentication"
)

var gcpClusterIdentitySecretType = corev1.SecretTypeOpaque

type GCPClusterIdentityConfig struct {
	*v1.SecretApplyConfiguration
	ManifestFiles    []string
	ManifestFilesDir string
}

func NewGCPClusterIdentityConfig(gcp *google_provider.GCPConfigure) *GCPClusterIdentityConfig {
	gcpcc := &GCPClusterIdentityConfig{
		SecretApplyConfiguration: v1.Secret(gcpClusterIdentitySecret, gcpClusterIdentityNamespace),
		ManifestFilesDir:         filepath.Join(os.TempDir(), gcpClusterIdentityName),
	}

	gcpcc.SecretApplyConfiguration.Type = &gcpClusterIdentitySecretType
	gcpcc.SecretApplyConfiguration.Data = map[string][]byte{"credentials": gcp.AppCredentials.JSON()}

	return gcpcc
}

func (gic *GCPClusterIdentityConfig) createGCPClusterIdentitySecretManifestFiles() error {
	if err := os.MkdirAll(gic.ManifestFilesDir, 0775); err != nil {
		return err
	}

	fileSecret, err := createManifestFile(gic.SecretApplyConfiguration, gic.ManifestFilesDir, gcpClusterIdentitySecret)
	if err != nil {
		return err
	}

	gic.ManifestFiles = append(gic.ManifestFiles, fileSecret)

	return nil
}

func (cc *ClusterCommands) applyGCPClusterIdentitySecret() error {
	var kubectlArgs = []string{"apply"}

	gcp := google_provider.NewGCPConfigure(cc.Ctx.Context, cc.Conf.GCPConfigure.AppCredentialsPath)
	if err := gcp.ReadSACredentials(); err != nil {
		return err
	}

	gic := NewGCPClusterIdentityConfig(gcp)
	if err := gic.createGCPClusterIdentitySecretManifestFiles(); err != nil {
		return err
	}

	for _, val := range gic.ManifestFiles {
		kubectlArgs = append(kubectlArgs, "-f", val)
	}

	cc.SpecCMD = cc.kubectl(kubectlArgs...)
	if err := releaseRunner(cc).runCMD(); err != nil {
		if err := os.RemoveAll(gic.ManifestFilesDir); err != nil {
			return err
		}

		return err
	}

	return os.RemoveAll(gic.ManifestFilesDir)
}

func (cc *ClusterCommands) getGCPClusterContext() ([]byte, error) {
	return google_provider.NewGCPConfigure(cc.Ctx.Context, cc.Conf.GCPConfigure.AppCredentialsPath).
		GetGCPClusterContext(cc.Conf.Name)
}
