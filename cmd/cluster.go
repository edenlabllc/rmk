package cmd

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"regexp"
	"strings"

	yaml2 "github.com/ghodss/yaml"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdlatest "k8s.io/client-go/tools/clientcmd/api/latest"

	"rmk/config"
	"rmk/providers/aws_provider"
	"rmk/providers/azure_provider"
	"rmk/providers/google_provider"
	"rmk/util"
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

func newClusterCommands(conf *config.Config, ctx *cli.Context, workDir string) *ClusterCommands {
	return &ClusterCommands{&ReleaseCommands{Conf: conf, Ctx: ctx, WorkDir: workDir}}
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
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
		Envs:    envs,
	}
}

func (cc *ClusterCommands) kubectl(args ...string) *util.SpecCMD {
	return &util.SpecCMD{
		Args:          args,
		Command:       "kubectl",
		Ctx:           cc.Ctx.Context,
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
	if cc.Ctx.Command.Category == util.CAPI {
		cc.APICluster = true
	}

	_, currentContextName, err := cc.getKubeContext()
	if err != nil {
		return "", "", err
	}

	cc.SpecCMD = cc.prepareHelmfile("--log-level", "error", "-l", "config=clusterctl", "template")
	cc.SpecCMD.DisableStdOut = true
	if err := releaseRunner(cc).runCMD(); err != nil {
		return "", "", fmt.Errorf("Helmfile failed to render template by release label: config=clusterctl\n%s",
			cc.SpecCMD.StderrBuf.String())
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

	re, err := regexp.Compile(`(?i)\b` + cc.Conf.Name + `\b`)
	if err != nil {
		return "", "", err
	}

	for key := range kubeConfig.Contexts {
		if re.MatchString(key) {
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

func (cc *ClusterCommands) provisionDestroyTargetCluster() error {
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

		cc.SpecCMD = cc.prepareHelmfile("--log-level", "error", "-l", "cluster="+cc.Conf.ClusterProvider, "sync")
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
		cc.SpecCMD = cc.prepareHelmfile("--log-level", "error", "-l", "cluster="+cc.Conf.ClusterProvider, "destroy")
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

func CAPIInitAction(conf *config.Config) cli.AfterFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
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
			return cc.applyAWSClusterIdentity()
		case azure_provider.AzureClusterProvider:
			return cc.applyAzureClusterIdentity()
		case google_provider.GoogleClusterProvider:
			return cc.applyGCPClusterIdentitySecret()
		}

		return nil
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
		if err := cc.switchKubeContext(); err != nil {
			return err
		}

		if err := cc.initClusterCTLConfig(); err != nil {
			return err
		}

		switch cc.Conf.ClusterProvider {
		case aws_provider.AWSClusterProvider:
			return cc.applyAWSClusterIdentity()
		case azure_provider.AzureClusterProvider:
			return cc.applyAzureClusterIdentity()
		case google_provider.GoogleClusterProvider:
			return cc.applyGCPClusterIdentitySecret()
		}

		return nil
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
		if err := cc.switchKubeContext(); err != nil {
			return err
		}

		return cc.provisionDestroyTargetCluster()
	}
}
