package cmd

import (
	rawjson "encoding/json"
	"fmt"
	"os"
	"strings"

	yaml2 "github.com/ghodss/yaml"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"

	"rmk/config"
	"rmk/git_handler"
	"rmk/providers/aws_provider"
	"rmk/providers/azure_provider"
	"rmk/providers/google_provider"
	"rmk/util"
)

const (
	clusterIdentityNameSuffix   = "identity"
	clusterIdentitySecretSuffix = "identity-secret"

	labelKeyCluster    = "cluster"
	labelKeyConfig     = "config"
	labelValClusterCTL = "clusterctl"
	labelValExtension  = "extension"

	objectIdentityName   = "identityName"
	objectIdentitySecret = "identitySecret"
	objectNamespace      = "namespace"
)

type clusterRunner interface {
	getKubeContext() (string, string, error)
	switchKubeContext() error
}

type ClusterCommands struct {
	*ReleaseCommands
}

type ClusterCTLConfig struct {
	ApiVersion string
	Kind       string
	Metadata   map[string]string
	Spec       interface{}
}

type CAPIManagementCluster struct {
	Name           string `json:"name,omitempty"`
	ServersRunning int    `json:"serversRunning,omitempty"`
	ServersCount   int    `json:"serversCount,omitempty"`
}

type ConfigExtension struct {
	Name      string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

type IdentityNameSpec struct {
	ClusterName string
	Suffix      string
	Prefix      string
}

type IdentityName struct {
	Objects map[string]IdentityNameSpec
}

func newClusterCommands(conf *config.Config, ctx *cli.Context, workDir string) *ClusterCommands {
	return &ClusterCommands{&ReleaseCommands{Conf: conf, Ctx: ctx, WorkDir: workDir}}
}

func newIdentityName(objects map[string]IdentityNameSpec) *IdentityName {
	return &IdentityName{objects}
}

func (in *IdentityName) generateObjectName(objectName string) string {
	if identity, ok := in.Objects[objectName]; ok && len(identity.Suffix) > 0 {
		return fmt.Sprintf("%s-%s", identity.ClusterName, identity.Suffix)
	} else if ok && len(identity.Prefix) > 0 {
		return fmt.Sprintf("%s-%s", identity.Prefix, identity.ClusterName)
	}

	return ""
}

func (cc *ClusterCommands) clusterCTL(args ...string) *util.SpecCMD {
	var envs []string

	switch cc.Conf.ClusterProvider {
	case aws_provider.AWSClusterProvider:
		envs = []string{
			"AWS_B64ENCODED_CREDENTIALS=",
			"CAPA_EKS_IAM=true",
			"CAPA_EKS_ADD_ROLES=true",
			"EXP_MACHINE_POOL=true",
		}
	case azure_provider.AzureClusterProvider:
		envs = []string{
			"EXP_AKS=true",
			"EXP_MACHINE_POOL=true",
			"EXP_CLUSTER_RESOURCE_SET=false",
		}
	case google_provider.GoogleClusterProvider:
		envs = []string{
			"GCP_B64ENCODED_CREDENTIALS=",
			"EXP_CAPG_GKE=true",
			"EXP_MACHINE_POOL=true",
		}
	}

	return &util.SpecCMD{
		Args:    args,
		Command: "clusterctl",
		Ctx:     cc.Ctx,
		Dir:     cc.WorkDir,
		Debug:   true,
		Envs:    envs,
	}
}

func (cc *ClusterCommands) kubectl(args ...string) *util.SpecCMD {
	return &util.SpecCMD{
		Args:          args,
		Command:       "kubectl",
		Ctx:           cc.Ctx,
		Dir:           cc.WorkDir,
		DisableStdOut: false,
		Debug:         true,
	}
}

func createManifestFile(object interface{}, dir, fileName string) (string, error) {
	data, err := json.Marshal(object)
	if err != nil {
		return "", err
	}

	return util.CreateTempYAMLFile(dir, fileName, data)
}

func createClusterCTLConfigFile(output []byte) (string, error) {
	clusterCTL := &ClusterCTLConfig{}
	if err := yaml.Unmarshal(output, &clusterCTL); err != nil {
		return "", err
	}

	data, err := yaml.Marshal(clusterCTL.Spec)
	if err != nil {
		return "", err
	}

	return util.CreateTempYAMLFile(os.TempDir(), clusterCTL.Metadata["name"], data)
}

func (cc *ClusterCommands) getClusterCTLConfig() (string, string, error) {
	var labelSelector = fmt.Sprintf("%s=%s", labelKeyConfig, labelValClusterCTL)

	if cc.Ctx.Command.Category == util.CAPI {
		cc.APICluster = true
	}

	_, currentContextName, err := cc.getKubeContext()
	if err != nil {
		return "", "", err
	}

	cc.SpecCMD = cc.prepareHelmfile("--log-level", "error", "--selector", labelSelector, "template")
	cc.SpecCMD.DisableStdOut = true
	if err := releaseRunner(cc).runCMD(); err != nil {
		return "", "", fmt.Errorf("Helmfile failed to render template by release label: %s\n%s",
			labelSelector, cc.SpecCMD.StderrBuf.String())
	}

	fileName, err := createClusterCTLConfigFile(cc.SpecCMD.StdoutBuf.Bytes())
	if err != nil {
		return "", "", err
	}

	return currentContextName, fileName, nil
}

func (cc *ClusterCommands) initClusterCTLConfig() error {
	contextName, clusterCTLConfig, err := cc.getClusterCTLConfig()
	if err != nil {
		return err
	}

	cc.SpecCMD = cc.clusterCTL("init", "--infrastructure", cc.Conf.ClusterProvider,
		"--wait-providers", "--kubeconfig-context", contextName, "--config", clusterCTLConfig)

	if err := releaseRunner(cc).runCMD(); err != nil {
		if err := os.RemoveAll(clusterCTLConfig); err != nil {
			return err
		}

		return err
	}

	return os.RemoveAll(clusterCTLConfig)
}

func (cc *ClusterCommands) initConfigExtensions() error {
	var (
		configExtensions []ConfigExtension
		labelSelector    = fmt.Sprintf("%s=%s,%s=%s",
			labelKeyConfig, labelValExtension, labelKeyCluster, cc.Conf.ClusterProvider)
	)

	cc.SpecCMD = cc.prepareHelmfile("--allow-no-matching-release", "--log-level", "error",
		"--selector", labelSelector, "list", "--output", "json")
	cc.SpecCMD.DisableStdOut = true
	if err := releaseRunner(cc).runCMD(); err != nil {
		return fmt.Errorf("failed to list config extensions: %s", cc.SpecCMD.StderrBuf.String())
	}

	if len(cc.SpecCMD.StdoutBuf.String()) == 0 {
		return nil
	}

	data := cc.SpecCMD.StdoutBuf.Bytes()
	src := (*rawjson.RawMessage)(&data)
	if err := json.Unmarshal(*src, &configExtensions); err != nil {
		return fmt.Errorf("failed to parse config extensions JSON: %w", err)
	}

	cc.SpecCMD = cc.prepareHelmfile("--log-level", "error", "--selector", labelSelector, "sync")
	cc.SpecCMD.DisableStdOut = true
	if err := releaseRunner(cc).runCMD(); err != nil {
		return fmt.Errorf("failed to sync config extensions: %s", cc.SpecCMD.StderrBuf.String())
	}

	for _, val := range configExtensions {
		zap.S().Info("installed config extension: name: " + val.Name + ", namespace: " + val.Namespace)
	}

	return nil
}

func (cc *ClusterCommands) manageKubeConfigItem(itemType, itemName string) error {
	cc.SpecCMD = cc.kubectl("config", itemType, itemName)
	cc.SpecCMD.DisableStdOut = true
	if err := releaseRunner(cc).runCMD(); err != nil {
		return fmt.Errorf("%s", strings.ReplaceAll(cc.SpecCMD.StderrBuf.String(), "\n", ""))
	}

	zap.S().Infof("%s", strings.ReplaceAll(cc.SpecCMD.StdoutBuf.String(), "\n", ""))

	return nil
}

func (cc *ClusterCommands) mergeKubeConfigs(clusterContext []byte) error {
	var object interface{}

	if err := yaml.Unmarshal(clusterContext, &object); err != nil {
		return err
	}

	file, err := createManifestFile(object, os.TempDir(), cc.Conf.Name+"-"+cc.Conf.ClusterProvider+"-kubeconfig")
	if err != nil {
		return err
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.Precedence = append(loadingRules.Precedence, file)

	mergeConfig, err := loadingRules.Load()
	if err != nil {
		if err := os.RemoveAll(file); err != nil {
			return err
		}

		return err
	}

	if err := os.RemoveAll(file); err != nil {
		return err
	}

	data, err := runtime.Encode(clientcmdlatest.Codec, mergeConfig)
	if err != nil {
		return err
	}

	kubeConfig, err := yaml2.JSONToYAML(data)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(util.GetHomePath(clientcmd.RecommendedHomeDir), 0755); err != nil {
		return err
	}

	if err := os.WriteFile(clientcmd.NewDefaultPathOptions().GlobalFile, kubeConfig, 0644); err != nil {
		return err
	}

	return cc.manageKubeConfigItem("use-context", cc.Conf.Name)
}

func (cc *ClusterCommands) getKubeContext() (string, string, error) {
	var (
		contextNames []string
		contextName  string
	)

	kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).RawConfig()
	if err != nil {
		return "", "", err
	}

	for key := range kubeConfig.Contexts {
		if cc.Conf.Name == key {
			contextNames = append(contextNames, key)
		}
	}

	switch {
	case len(contextNames) > 1:
		return "", "",
			fmt.Errorf("detected more than one Kubernetes context with names %s leading to conflict, "+
				"please delete or rename all contexts except one", strings.Join(contextNames, ", "))
	case len(contextNames) > 0:
		contextName = contextNames[0]
	default:
		contextName = ""
	}

	if cc.K3DCluster && len(contextName) > 0 && !strings.Contains(contextName, util.K3DPrefix) {
		return "", "", fmt.Errorf("remote Kubernetes context already exists %s for this branch", contextName)
	}

	return contextName, kubeConfig.CurrentContext, nil
}

func (cc *ClusterCommands) switchKubeContext() error {
	contextName, currentContextName, err := cc.getKubeContext()
	if err != nil {
		return err
	}

	if cc.Ctx.Command.Category == util.CAPI {
		kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).RawConfig()
		if err != nil {
			return err
		}

		if _, ok := kubeConfig.Contexts[util.CAPIContextName]; ok && currentContextName != util.CAPIContextName {
			return cc.manageKubeConfigItem("use-context", util.CAPIContextName)
		} else if ok && currentContextName == util.CAPIContextName {
			return nil
		} else if !ok {
			return fmt.Errorf("kubernetes context %s for CAPI management cluster not found", util.CAPIContextName)
		}
	}

	if len(contextName) > 0 && !cc.UpdateContext {
		if contextName != currentContextName {
			return cc.manageKubeConfigItem("use-context", contextName)
		}

		return nil
	}

	if strings.Contains(contextName, util.K3DPrefix) && cc.UpdateContext {
		return fmt.Errorf("current context %s already used for K3D cluster, --force flag cannot be used", contextName)
	}

	switch cc.Conf.ClusterProvider {
	case aws_provider.AWSClusterProvider:
		clusterContext, err := cc.getAWSClusterContext()
		if err != nil {
			return err
		}

		if err := cc.mergeKubeConfigs(clusterContext); err != nil {
			return err
		}
	case azure_provider.AzureClusterProvider:
		clusterContext, err := cc.getAzureClusterContext()
		if err != nil {
			return err
		}

		if err := cc.mergeKubeConfigs(clusterContext); err != nil {
			return err
		}
	case google_provider.GoogleClusterProvider:
		clusterContext, err := cc.getGCPClusterContext()
		if err != nil {
			return err
		}

		if err := cc.mergeKubeConfigs(clusterContext); err != nil {
			return err
		}
	}

	return nil
}

func (cc *ClusterCommands) checkCAPIManagementCluster() error {
	var cMC []CAPIManagementCluster

	k := &K3DCommands{cc.ReleaseCommands}
	if err := k.prepareK3D("cluster", "list", util.CAPI, "--output", "json"); err != nil {
		return err
	}

	k.SpecCMD.DisableStdOut = true
	if err := releaseRunner(k).runCMD(); err != nil {
		if strings.Contains(k.SpecCMD.StderrBuf.String(), "No nodes found for given cluster") {
			return fmt.Errorf("%s management cluster not found", strings.ToUpper(util.CAPI))
		} else {
			return fmt.Errorf("%s", k.SpecCMD.StderrBuf.String())
		}
	}

	data := k.SpecCMD.StdoutBuf.Bytes()
	src := (*rawjson.RawMessage)(&data)
	if err := json.Unmarshal(*src, &cMC); err != nil {
		return err
	}

	if len(cMC) > 0 {
		if cMC[0].Name == util.CAPI && (cMC[0].ServersCount != 1 || cMC[0].ServersRunning != 1) {
			return fmt.Errorf("%s management cluster not running", strings.ToUpper(util.CAPI))
		}
	}

	return nil
}

func (cc *ClusterCommands) provisionDestroyTargetCluster() error {
	var labelSelector = fmt.Sprintf("%s=%s,%s!=%s",
		labelKeyCluster, cc.Conf.ClusterProvider, labelKeyConfig, labelValExtension)

	if cc.Ctx.Command.Category == util.CAPI {
		cc.APICluster = true
	}

	_, _, err := cc.getKubeContext()
	if err != nil {
		return err
	}

	switch cc.Ctx.Command.Name {
	case "provision":
		switch cc.Conf.ClusterProvider {
		case aws_provider.AWSClusterProvider:
			if err := cc.createAWSClusterSSHKey(); err != nil {
				return err
			}

			if err := cc.createAWSSecrets(); err != nil {
				return err
			}
		case azure_provider.AzureClusterProvider:
			if err := cc.createAzureSecrets(cc.Conf.AzureConfigure); err != nil {
				return err
			}
		case google_provider.GoogleClusterProvider:
			if err := cc.createGCPNATGateway(); err != nil {
				return err
			}

			if err := cc.createGCPSecrets(); err != nil {
				return err
			}
		}

		cc.SpecCMD = cc.prepareHelmfile("--log-level", "error", "--selector", labelSelector, "sync")
		if err := releaseRunner(cc).runCMD(); err != nil {
			return err
		}

		switch cc.Conf.ClusterProvider {
		case aws_provider.AWSClusterProvider:
			clusterContext, err := cc.getAWSClusterContext()
			if err != nil {
				return err
			}

			if err := cc.mergeKubeConfigs(clusterContext); err != nil {
				return err
			}
		case azure_provider.AzureClusterProvider:
			clusterContext, err := cc.getAzureClusterContext()
			if err != nil {
				return err
			}

			if err := cc.mergeKubeConfigs(clusterContext); err != nil {
				return err
			}
		case google_provider.GoogleClusterProvider:
			clusterContext, err := cc.getGCPClusterContext()
			if err != nil {
				return err
			}

			if err := cc.mergeKubeConfigs(clusterContext); err != nil {
				return err
			}
		}
	case "destroy":
		cc.SpecCMD = cc.prepareHelmfile("--log-level", "error", "--selector", labelSelector, "destroy")
		if err := releaseRunner(cc).runCMD(); err != nil {
			return err
		}

		switch cc.Conf.ClusterProvider {
		case aws_provider.AWSClusterProvider:
			if err := cc.deleteAWSClusterSSHKey(); err != nil {
				return err
			}
		case azure_provider.AzureClusterProvider:
			// Pre-destroy hook for Azure
		case google_provider.GoogleClusterProvider:
			if err := cc.deleteGCPNATGateway(); err != nil {
				return err
			}
		}

		kubeConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			clientcmd.NewDefaultClientConfigLoadingRules(), &clientcmd.ConfigOverrides{}).RawConfig()
		if err != nil {
			return err
		}

		if context, ok := kubeConfig.Contexts[cc.Conf.Name]; ok {
			if err := cc.manageKubeConfigItem("delete-context", cc.Conf.Name); err != nil {
				return err
			}

			if err := cc.manageKubeConfigItem("delete-cluster", context.Cluster); err != nil {
				return err
			}

			if err := cc.manageKubeConfigItem("delete-user", context.AuthInfo); err != nil {
				return err
			}
		}
	}

	return nil
}

func clusterSwitchAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		cc := newClusterCommands(conf, c, util.GetPwdPath(""))
		cc.UpdateContext = c.Bool("force")

		return cc.switchKubeContext()
	}
}

func CAPIInitAction(conf *config.Config, gitSpec *git_handler.GitSpec) cli.AfterFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		// Additional checking is needed because After() is run even if Action() panics
		configPath := util.GetHomePath(util.RMKDir, util.RMKConfig, gitSpec.ID+".yaml")
		if !util.IsExists(configPath, true) {
			return nil
		}

		cc := newClusterCommands(conf, c, util.GetPwdPath())
		if err := cc.switchKubeContext(); err != nil {
			return err
		}

		if err := cc.initClusterCTLConfig(); err != nil {
			return err
		}

		switch cc.Conf.ClusterProvider {
		case aws_provider.AWSClusterProvider:
			if err := cc.applyAWSClusterIdentity(); err != nil {
				return err
			}
		case azure_provider.AzureClusterProvider:
			if err := cc.applyAzureClusterIdentity(); err != nil {
				return err
			}
		case google_provider.GoogleClusterProvider:
			if err := cc.applyGCPClusterIdentitySecret(); err != nil {
				return err
			}
		}

		return cc.initConfigExtensions()
	}
}

func CAPIUpdateAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		cc := newClusterCommands(conf, c, util.GetPwdPath())
		if err := cc.checkCAPIManagementCluster(); err != nil {
			return err
		}

		if err := cc.switchKubeContext(); err != nil {
			return err
		}

		if err := cc.initClusterCTLConfig(); err != nil {
			return err
		}

		switch cc.Conf.ClusterProvider {
		case aws_provider.AWSClusterProvider:
			if err := cc.applyAWSClusterIdentity(); err != nil {
				return err
			}
		case azure_provider.AzureClusterProvider:
			if err := cc.applyAzureClusterIdentity(); err != nil {
				return err
			}
		case google_provider.GoogleClusterProvider:
			if err := cc.applyGCPClusterIdentitySecret(); err != nil {
				return err
			}
		}

		return cc.initConfigExtensions()
	}
}

func CAPIProvisionDestroyAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		cc := newClusterCommands(conf, c, util.GetPwdPath())
		if err := cc.checkCAPIManagementCluster(); err != nil {
			return err
		}

		if err := cc.switchKubeContext(); err != nil {
			return err
		}

		return cc.provisionDestroyTargetCluster()
	}
}
