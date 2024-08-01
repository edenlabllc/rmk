package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"rmk/config"
	"rmk/util"
)

type StateRunner interface {
	clusterStateDelete() error
	clusterStateList() error
	clusterStateRefresh() error
}

type ClusterCommands struct {
	Conf     *config.Config
	Ctx      *cli.Context
	SpecCMDs []*util.SpecCMD
	PlanFile string
	WorkDir  string
}

func (cc *ClusterCommands) clusterRootDir() (string, error) {
	for _, provider := range cc.Conf.Clusters {
		if strings.HasPrefix(provider.Name, cc.Conf.ClusterProvider) {
			return provider.DstPath, nil
		}
	}

	return "", fmt.Errorf("destination path for cluster provider %s not found", cc.Conf.ClusterProvider)
}

func (cc *ClusterCommands) awsEks() *util.SpecCMD {
	return &util.SpecCMD{
		Args: []string{"eks", "--region",
			cc.Conf.Region,
			"update-kubeconfig",
			"--name",
			cc.Conf.Name + "-eks",
			"--profile",
			cc.Conf.Profile,
		},
		Command: "aws",
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
	}
}

func (cc *ClusterCommands) runBatchCMD() error {
	if err := os.Unsetenv("AWS_PROFILE"); err != nil {
		return err
	}

	for _, val := range cc.SpecCMDs {
		val.Envs = []string{
			"AWS_PROFILE=" + cc.Conf.Profile,
			"AWS_CONFIG_FILE=" + strings.Join(cc.Conf.AWSSharedConfigFile(cc.Conf.Profile), ""),
			"AWS_SHARED_CREDENTIALS_FILE=" + strings.Join(cc.Conf.AWSSharedCredentialsFile(cc.Conf.Profile), ""),
		}
		if err := val.AddEnv(); err != nil {
			return err
		}

		if err := val.ExecCMD(); err != nil {
			if val.Debug {
				zap.S().Debugf("command: %s", val.CommandStr)
				zap.S().Debugf("path: %s", val.Dir)
				for _, v := range val.Envs {
					zap.S().Debugf("env: %s", v)
				}
			}

			return err
		}

		if val.Debug {
			zap.S().Debugf("command: %s", val.CommandStr)
			zap.S().Debugf("path: %s", val.Dir)
			for _, v := range val.Envs {
				zap.S().Debugf("env: %s", v)
			}
		}
	}

	return nil
}

func (cc *ClusterCommands) initialize() *util.SpecCMD {
	args := []string{
		"init",
		"-backend=true",
		"-backend-config=region=" + cc.Conf.Region,
		"-backend-config=bucket=" + cc.Conf.Terraform.BucketName,
		"-backend-config=key=" + cc.Conf.Terraform.BucketKey,
	}
	if cc.Conf.ClusterProvisionerSL {
		args = append(args, "-backend-config=dynamodb_table="+cc.Conf.Terraform.DDBTableName)
	}

	args = append(args, "-reconfigure")
	return &util.SpecCMD{
		Args:    args,
		Command: "terraform",
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
	}
}

func (cc *ClusterCommands) validate() *util.SpecCMD {
	return &util.SpecCMD{
		Args:    []string{"validate"},
		Command: "terraform",
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
	}
}

func (cc *ClusterCommands) workspace(args ...string) *util.SpecCMD {
	return &util.SpecCMD{
		Args:    append([]string{"workspace"}, args...),
		Command: "terraform",
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
	}
}

func (cc *ClusterCommands) output(args ...string) *util.SpecCMD {
	return &util.SpecCMD{
		Args:          append([]string{"output"}, args...),
		Command:       "terraform",
		Ctx:           cc.Ctx.Context,
		Dir:           cc.WorkDir,
		DisableStdOut: true,
		Debug:         false,
	}
}

func (cc *ClusterCommands) destroy() *util.SpecCMD {
	return &util.SpecCMD{
		Args: []string{"destroy", "-auto-approve",
			"-var=aws_account_id=" + cc.Conf.AccountID,
			"-var=cloudflare_api_token=" + cc.Conf.CloudflareToken,
			"-var=name=" + cc.Conf.Name,
			"-var=region=" + cc.Conf.Region,
			"-var=root_domain=" + cc.Conf.RootDomain,
			"-var=terraform_bucket_key=" + cc.Conf.Terraform.BucketKey,
			"-var=terraform_bucket_name=" + cc.Conf.Terraform.BucketName,
		},
		Command: "terraform",
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
	}
}

func (cc *ClusterCommands) plan() *util.SpecCMD {
	return &util.SpecCMD{
		Args: []string{
			"plan",
			"-out=" + cc.PlanFile,
			"-var=aws_account_id=" + cc.Conf.AccountID,
			"-var=cloudflare_api_token=" + cc.Conf.CloudflareToken,
			"-var=name=" + cc.Conf.Name,
			"-var=region=" + cc.Conf.Region,
			"-var=root_domain=" + cc.Conf.RootDomain,
			"-var=terraform_bucket_key=" + cc.Conf.Terraform.BucketKey,
			"-var=terraform_bucket_name=" + cc.Conf.Terraform.BucketName,
		},
		Command: "terraform",
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
	}
}

func (cc *ClusterCommands) apply() *util.SpecCMD {
	return &util.SpecCMD{
		Args:    []string{"apply", cc.PlanFile},
		Command: "terraform",
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
	}
}

func (cc *ClusterCommands) listResources() *util.SpecCMD {
	return &util.SpecCMD{
		Args:    []string{"state", "list"},
		Command: "terraform",
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
	}
}

func (cc *ClusterCommands) refresh() *util.SpecCMD {
	return &util.SpecCMD{
		Args: []string{"refresh",
			"-var=aws_account_id=" + cc.Conf.AccountID,
			"-var=cloudflare_api_token=" + cc.Conf.CloudflareToken,
			"-var=name=" + cc.Conf.Name,
			"-var=region=" + cc.Conf.Region,
			"-var=root_domain=" + cc.Conf.RootDomain,
			"-var=terraform_bucket_key=" + cc.Conf.Terraform.BucketKey,
			"-var=terraform_bucket_name=" + cc.Conf.Terraform.BucketName,
		},
		Command: "terraform",
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
	}
}

func (cc *ClusterCommands) state(args ...string) *util.SpecCMD {
	return &util.SpecCMD{
		Args:    append([]string{"state"}, args...),
		Command: "terraform",
		Ctx:     cc.Ctx.Context,
		Dir:     cc.WorkDir,
		Debug:   true,
	}
}

func (cc *ClusterCommands) clusterContext() error {
	cc.SpecCMDs = append(cc.SpecCMDs, cc.awsEks())
	return cc.runBatchCMD()
}

func (cc *ClusterCommands) clusterDestroy() error {
	checkWorkspace, err := cc.Conf.BucketKeyExists("", cc.Conf.Terraform.BucketName, "env:/"+cc.Conf.Name+"/tf.tfstate")
	if err != nil {
		return err
	}

	if checkWorkspace {
		cc.SpecCMDs = append(cc.SpecCMDs, cc.initialize(), cc.validate(), cc.workspace("select", cc.Conf.Name))
		if err := cc.runBatchCMD(); err != nil {
			return err
		}

		destroy := cc.destroy()

		match, err := util.WalkMatch(
			util.GetPwdPath(util.TenantValuesDIR, "clusters", util.AWSClusterProvider, cc.Conf.Environment),
			"*."+util.TerraformVarsExt,
		)
		if err != nil {
			return err
		}

		for _, val := range match {
			destroy.Args = append(destroy.Args, "-var-file="+val)
		}

		if err := os.RemoveAll(cc.PlanFile); err != nil {
			return err
		}

		cc.SpecCMDs = append([]*util.SpecCMD{}, destroy, cc.workspace("select", "default"),
			cc.workspace("delete", cc.Conf.Name))

		return cc.runBatchCMD()
	} else {
		zap.S().Infof("Terraform cluster in workspace %s already deleted or not created", cc.Conf.Name)
		return nil
	}
}

func (cc *ClusterCommands) clusterList() error {
	cc.SpecCMDs = append(cc.SpecCMDs, cc.initialize(), cc.workspace("list"))
	return cc.runBatchCMD()
}

func (cc *ClusterCommands) clusterProvision() error {
	var workspace *util.SpecCMD

	if err := os.MkdirAll(filepath.Join(cc.WorkDir, "plans"), 0755); err != nil {
		zap.S().Fatal(err)
	}

	checkWorkspace, err := cc.Conf.BucketKeyExists("", cc.Conf.Terraform.BucketName, "env:/"+cc.Conf.Name+"/tf.tfstate")
	if err != nil {
		zap.S().Fatal(err)
	}

	if checkWorkspace {
		workspace = cc.workspace("select", cc.Conf.Name)
	} else {
		workspace = cc.workspace("new", cc.Conf.Name)
	}

	plan := cc.plan()

	match, err := util.WalkMatch(
		util.GetPwdPath(util.TenantValuesDIR, "clusters", util.AWSClusterProvider, cc.Conf.Environment),
		"*."+util.TerraformVarsExt,
	)

	for _, val := range match {
		plan.Args = append(plan.Args, "-var-file="+val)
	}

	if cc.Ctx.Bool("plan") {
		cc.SpecCMDs = append(cc.SpecCMDs, cc.initialize(), cc.validate(), workspace, plan)
		return cc.runBatchCMD()
	}

	cc.SpecCMDs = append(cc.SpecCMDs, cc.initialize(), cc.validate(), workspace, plan, cc.apply())
	if err := cc.runBatchCMD(); err != nil {
		return err
	}

	rc := &ReleaseCommands{
		Conf:          cc.Conf,
		Ctx:           cc.Ctx,
		WorkDir:       util.GetPwdPath(""),
		UpdateContext: true,
	}

	return rc.releaseKubeContext()
}

func (cc *ClusterCommands) clusterStateDelete() error {
	cc.SpecCMDs = append(cc.SpecCMDs, cc.state("rm", cc.Ctx.String("resource-address")))

	return cc.runBatchCMD()
}

func (cc *ClusterCommands) clusterStateList() error {
	cc.SpecCMDs = append(cc.SpecCMDs, cc.state(cc.Ctx.Command.Name))

	return cc.runBatchCMD()
}

func (cc *ClusterCommands) clusterStateRefresh() error {
	var workspace *util.SpecCMD

	checkWorkspace, err := cc.Conf.BucketKeyExists("", cc.Conf.Terraform.BucketName, "env:/"+cc.Conf.Name+"/tf.tfstate")
	if err != nil {
		zap.S().Fatal(err)
	}

	if checkWorkspace {
		workspace = cc.workspace("select", cc.Conf.Name)
	} else {
		workspace = cc.workspace("new", cc.Conf.Name)
	}

	refresh := cc.refresh()
	match, err := util.WalkMatch(
		util.GetPwdPath(util.TenantValuesDIR, "clusters", util.AWSClusterProvider, cc.Conf.Environment),
		"*."+util.TerraformVarsExt,
	)
	if err != nil {
		return err
	}

	for _, val := range match {
		refresh.Args = append(refresh.Args, "-var-file="+val)
	}

	cc.SpecCMDs = append(cc.SpecCMDs, cc.initialize(), cc.validate(), workspace, refresh)

	return cc.runBatchCMD()
}

func clusterDestroyAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(false), c, false); err != nil {
			return err
		}

		cc := &ClusterCommands{
			Conf: conf,
			Ctx:  c,
		}

		if pkgDst, err := cc.clusterRootDir(); err != nil {
			return err
		} else {
			cc.WorkDir = filepath.Join(pkgDst, "terraform")
		}

		cc.PlanFile = filepath.Join(cc.WorkDir, "plans", conf.Name+"__"+conf.Environment+".tfplan")

		return cc.clusterDestroy()
	}
}

func clusterListAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(false), c, false); err != nil {
			return err
		}

		cc := &ClusterCommands{
			Conf: conf,
			Ctx:  c,
		}

		if pkgDst, err := cc.clusterRootDir(); err != nil {
			return err
		} else {
			cc.WorkDir = filepath.Join(pkgDst, "terraform")
		}

		return cc.clusterList()
	}
}

func clusterProvisionAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(true), c, false); err != nil {
			return err
		}

		cc := &ClusterCommands{
			Conf: conf,
			Ctx:  c,
		}

		if pkgDst, err := cc.clusterRootDir(); err != nil {
			return err
		} else {
			cc.WorkDir = filepath.Join(pkgDst, "terraform")
		}

		cc.PlanFile = filepath.Join(cc.WorkDir, "plans", conf.Name+"__"+conf.Environment+".tfplan")

		if err := cc.clusterProvision(); err != nil {
			return err
		}

		if err := conf.GetTerraformOutputs(); err != nil {
			return err
		}

		return conf.CreateConfigFile()
	}
}

func clusterStateAction(conf *config.Config, action func(stateRunner StateRunner) error) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(false), c, false); err != nil {
			return err
		}

		cc := &ClusterCommands{
			Conf: conf,
			Ctx:  c,
		}

		if pkgDst, err := cc.clusterRootDir(); err != nil {
			return err
		} else {
			cc.WorkDir = filepath.Join(pkgDst, "terraform")
		}

		return action(cc)
	}
}

func clusterSwitchAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(false), c, false); err != nil {
			return err
		}

		rc := &ReleaseCommands{
			Conf:          conf,
			Ctx:           c,
			WorkDir:       util.GetPwdPath(""),
			UpdateContext: c.Bool("force"),
		}

		return rc.releaseKubeContext()
	}
}
