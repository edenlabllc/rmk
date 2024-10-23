package cmd

import (
	"strings"

	"rmk/util"
)

const (
	awsFlagsCategory = "AWS authentication"
)

func (cc *ClusterCommands) getAWSEksKubeConfig() *util.SpecCMD {
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
	cc.SpecCMD = cc.getAWSEksKubeConfig()
	return releaseRunner(cc).runCMD()
}
