package cmd

import (
	"os"
	"path/filepath"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/client-go/applyconfigurations/core/v1"

	"rmk/providers/google_provider"
	"rmk/util"
)

const (
	gcpClusterIdentityNamespacePrefix = "capg"
	gcpFlagsCategory                  = "GCP authentication"
)

var gcpClusterIdentitySecretType = corev1.SecretTypeOpaque

type GCPClusterIdentityConfig struct {
	*v1.NamespaceApplyConfiguration
	*v1.SecretApplyConfiguration
	ManifestFiles    []string
	ManifestFilesDir string
}

func NewGCPClusterIdentityConfig(gcp *google_provider.GCPConfigure, identity *IdentityName) *GCPClusterIdentityConfig {
	gcpcc := &GCPClusterIdentityConfig{
		NamespaceApplyConfiguration: v1.Namespace(identity.generateObjectName(objectNamespace)),
		SecretApplyConfiguration: v1.Secret(identity.generateObjectName(objectIdentitySecret),
			identity.generateObjectName(objectNamespace)),
		ManifestFilesDir: filepath.Join(os.TempDir(), identity.generateObjectName(objectIdentityName)),
	}

	gcpcc.SecretApplyConfiguration.Type = &gcpClusterIdentitySecretType
	gcpcc.SecretApplyConfiguration.Data = map[string][]byte{"credentials": gcp.AppCredentials.JSON()}

	return gcpcc
}

func (gic *GCPClusterIdentityConfig) createGCPClusterIdentitySecretManifestFiles(identity *IdentityName) error {
	if err := os.MkdirAll(gic.ManifestFilesDir, 0775); err != nil {
		return err
	}

	fileNamespace, err := createManifestFile(gic.NamespaceApplyConfiguration, gic.ManifestFilesDir,
		identity.generateObjectName(objectNamespace))
	if err != nil {
		return err
	}

	fileSecret, err := createManifestFile(gic.SecretApplyConfiguration, gic.ManifestFilesDir,
		identity.generateObjectName(objectIdentitySecret))
	if err != nil {
		return err
	}

	gic.ManifestFiles = append(gic.ManifestFiles, fileNamespace, fileSecret)

	return nil
}

func (cc *ClusterCommands) applyGCPClusterIdentitySecret() error {
	var kubectlArgs = []string{"apply"}

	gcp := google_provider.NewGCPConfigure(cc.Ctx.Context, cc.Conf.GCPConfigure.AppCredentialsPath)
	if err := gcp.ReadSACredentials(); err != nil {
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
			Prefix:      gcpClusterIdentityNamespacePrefix,
		},
	})

	gic := NewGCPClusterIdentityConfig(gcp, identityName)
	if err := gic.createGCPClusterIdentitySecretManifestFiles(identityName); err != nil {
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

func (cc *ClusterCommands) createGCPNATGateway() error {
	return google_provider.NewGCPConfigure(cc.Ctx.Context, cc.Conf.GCPConfigure.AppCredentialsPath).
		CreateGCPCloudNATGateway(cc.Conf.GCPRegion)
}

func (cc *ClusterCommands) deleteGCPNATGateway() error {
	return google_provider.NewGCPConfigure(cc.Ctx.Context, cc.Conf.GCPConfigure.AppCredentialsPath).
		DeleteGCPCloudNATGateway(cc.Conf.GCPRegion)
}

func (cc *ClusterCommands) createGCPSecrets() error {
	gcp := google_provider.NewGCPConfigure(cc.Ctx.Context, cc.Conf.GCPConfigure.AppCredentialsPath)

	secrets, err := gcp.GetGCPSecrets(cc.Conf.Tenant)
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

		if err := gcp.SetGCPSecret(cc.Conf.Tenant, cc.Conf.GCPRegion, keyName, file); err != nil {
			return err
		}
	}

	return nil
}
