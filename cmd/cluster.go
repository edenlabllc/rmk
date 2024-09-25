package cmd

import (
	"strings"

	"github.com/urfave/cli/v2"

	"rmk/config"
	"rmk/util"
)

type ClusterCommands struct {
	*ReleaseCommands
}

func newClusterCommands(conf *config.Config, ctx *cli.Context, workDir string) *ClusterCommands {
	return &ClusterCommands{&ReleaseCommands{Conf: conf, Ctx: ctx, WorkDir: workDir}}
}

func (cc *ClusterCommands) awsEks() *util.SpecCMD {
	return &util.SpecCMD{
		Envs: []string{
			"AWS_PROFILE=" + cc.Conf.Profile,
			"AWS_CONFIG_FILE=" + strings.Join(cc.Conf.AWSSharedConfigFile(cc.Conf.Profile), ""),
			"AWS_SHARED_CREDENTIALS_FILE=" + strings.Join(cc.Conf.AWSSharedCredentialsFile(cc.Conf.Profile), ""),
		},
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

func (cc *ClusterCommands) awsClusterContext() error {
	cc.SpecCMD = cc.awsEks()
	return runner(cc).runCMD()
}

func clusterSwitchAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateGitHubToken(c, ""); err != nil {
			return err
		}

		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
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
