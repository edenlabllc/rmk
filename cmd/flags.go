package cmd

import (
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"

	"rmk/git_handler"
	"rmk/util"
)

func flagsConfig() []cli.Flag {
	return []cli.Flag{
		// TODO: will be deprecated
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "artifact-mode",
				Usage:   "choice of artifact usage model, available: none, online, offline",
				Aliases: []string{"am"},
				EnvVars: []string{"RMK_ARTIFACT_MODE"},
				Value:   util.ArtifactModeDefault,
			},
		),
		// TODO: will be transfer to cluster category for AWS provider
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "aws-ecr-host",
				Usage:   "AWS ECR host",
				Aliases: []string{"aeh"},
				EnvVars: []string{"RMK_AWS_ECR_HOST"},
				Value:   util.AWSECRHost,
			},
		),
		// TODO: will be transfer to cluster category for AWS provider
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "aws-ecr-region",
				Usage:   "AWS region for specific ECR host",
				Aliases: []string{"aer"},
				EnvVars: []string{"RMK_AWS_ECR_REGION"},
				Value:   util.AWSECRRegion,
			},
		),
		// TODO: will be transfer to cluster category for AWS provider
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "aws-ecr-user-name",
				Usage:   "AWS ECR user name",
				Aliases: []string{"aeun"},
				EnvVars: []string{"RMK_AWS_ECR_USER_NAME"},
				Value:   util.AWSECRUserName,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:   "aws-mfa-profile",
				Hidden: true,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:   "aws-mfa-token-expiration",
				Hidden: true,
			},
		),
		&cli.BoolFlag{
			Name:    "aws-reconfigure",
			Usage:   "force AWS profile creation",
			Aliases: []string{"r"},
		},
		// TODO: will be deprecated
		&cli.BoolFlag{
			Name:    "aws-reconfigure-artifact-license",
			Usage:   "force AWS profile creation for artifact license, used only if RMK config option artifact-mode has values: online, offline",
			Aliases: []string{"l"},
		},
		// TODO: will be transfer to cluster category for AWS provider
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Name:    "cluster-provisioner-state-locking",
				Usage:   "disable or enable cluster provisioner state locking",
				Aliases: []string{"c"},
				Value:   true,
			},
		),
		&cli.StringFlag{
			Name:   "config",
			Hidden: true,
		},
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:   "config-from",
				Hidden: true,
			},
		),
		&cli.StringFlag{
			Name: "config-from-environment",
			Usage: "inheritance of RMK config credentials from environments: " +
				git_handler.DefaultDevelop + ", " + git_handler.DefaultStaging + ", " + git_handler.DefaultProduction,
			Aliases: []string{"cfe"},
			EnvVars: []string{"RMK_CONFIG_FROM_ENVIRONMENT"},
		},
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "github-token",
				Usage:   "personal access token for download GitHub artifacts",
				Aliases: []string{"ght"},
				EnvVars: []string{"RMK_GITHUB_TOKEN"},
			},
		),
		// TODO: will be transfer to cluster category
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "cloudflare-token",
				Usage:   "Cloudflare API token for provision NS records",
				Aliases: []string{"cft"},
				EnvVars: []string{"RMK_CLOUDFLARE_TOKEN"},
			},
		),
		// TODO: will be transfer to cluster category for AWS provider
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "root-domain",
				Usage:   "domain name for external access to app services via ingress controller",
				Aliases: []string{"rd"},
				EnvVars: []string{"RMK_ROOT_DOMAIN"},
			},
		),
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Name:    "progress-bar",
				Usage:   "globally disable or enable progress bar for download process",
				Aliases: []string{"p"},
				Value:   true,
			},
		),
		// TODO: will be deprecated
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "s3-charts-repo-region",
				Usage:   "location constraint region of S3 charts repo",
				Aliases: []string{"scrr"},
				EnvVars: []string{"RMK_S3_CHARTS_REPO_REGION"},
				Value:   util.S3ChartsRepoRegion,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "cluster-provider",
				Usage:   "select cluster provider to provision clusters",
				Aliases: []string{"cp"},
				EnvVars: []string{"RMK_CLUSTER_PROVIDER"},
				Value:   util.AWSClusterProvider,
			},
		),
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Name:    "slack-notifications",
				Usage:   "enable Slack notifications",
				Aliases: []string{"n"},
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "slack-webhook",
				Usage:   "URL for Slack webhook",
				Aliases: []string{"sw"},
				EnvVars: []string{"RMK_SLACK_WEBHOOK"},
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "slack-channel",
				Usage:   "channel name for Slack notification",
				Aliases: []string{"sc"},
				EnvVars: []string{"RMK_SLACK_CHANNEL"},
			},
		),
		altsrc.NewStringSliceFlag(
			&cli.StringSliceFlag{
				Name:    "slack-message-details",
				Usage:   "additional information for body of Slack message",
				Aliases: []string{"smd"},
				EnvVars: []string{"RMK_SLACK_MESSAGE_DETAILS"},
			},
		),
	}
}

func flagsConfigList() []cli.Flag {
	return append(flagsHidden(),
		&cli.BoolFlag{
			Name:    "all",
			Usage:   "list all tenant configurations",
			Aliases: []string{"a"},
		},
	)
}

func flagsClusterK3DCreate() []cli.Flag {
	return append(flagsHidden(),
		&cli.StringFlag{
			Name:        "k3d-volume-host-path",
			Usage:       "host local directory path for mount into K3D cluster",
			Aliases:     []string{"kv"},
			EnvVars:     []string{"RMK_K3D_VOLUME_HOST_PATH"},
			DefaultText: "present working directory",
		},
	)
}

func flagsClusterK3DImport() []cli.Flag {
	return append(flagsHidden(),
		&cli.StringSliceFlag{
			Name:     "k3d-import-image",
			Usage:    "list images for import into K3D cluster",
			Aliases:  []string{"ki"},
			EnvVars:  []string{"RMK_K3D_IMPORT_IMAGE"},
			Required: true,
		},
	)
}

func flagsClusterCRLogin() []cli.Flag {
	return append(flagsHidden(),
		&cli.BoolFlag{
			Name:    "get-token",
			Usage:   "get ECR token for authentication",
			Aliases: []string{"g"},
		},
	)
}

func flagsClusterPlan() []cli.Flag {
	return append(flagsHidden(),
		&cli.BoolFlag{
			Name:    "plan",
			Usage:   "creates an execution Terraform plan",
			Aliases: []string{"p"},
		},
	)
}

func flagsClusterStateDelete() []cli.Flag {
	return append(flagsHidden(),
		&cli.StringFlag{
			Name:    "resource-address",
			Usage:   "resource address for delete from Terraform state",
			Aliases: []string{"ra"},
			EnvVars: []string{"RMK_CLUSTER_STATE_RESOURCE_ADDRESS"},
		},
	)
}

func flagsClusterSwitch() []cli.Flag {
	return append(flagsHidden(),
		&cli.BoolFlag{
			Name:    "force",
			Usage:   "force update Kubernetes context from remote cluster",
			Aliases: []string{"f"},
		},
	)
}

func flagsProjectGenerate() []cli.Flag {
	return append(flagsHidden(),
		&cli.BoolFlag{
			Name:    "create-sops-age-keys",
			Usage:   "create SOPS age keys for generated project structure",
			Aliases: []string{"c"},
		},
	)
}

func flagsProjectUpdate() []cli.Flag {
	return append(flagsHidden(),
		&cli.StringFlag{
			Name:     "dependency",
			Usage:    "specific dependency name for updating project file",
			Aliases:  []string{"d"},
			Required: true,
			EnvVars:  []string{"RMK_PROJECT_UPDATE_DEPENDENCY"},
		},
		&cli.BoolFlag{
			Name:    "skip-ci",
			Usage:   "add [skip ci] to commit message line to skip triggering other CI builds",
			Aliases: []string{"i"},
		},
		&cli.BoolFlag{
			Name:    "skip-commit",
			Usage:   "only change a version in for project file without committing and pushing it",
			Aliases: []string{"c"},
		},
		&cli.StringFlag{
			Name:     "version",
			Usage:    "specific dependency version for updating project file",
			Aliases:  []string{"v"},
			Required: true,
			EnvVars:  []string{"RMK_PROJECT_UPDATE_VERSION"},
		},
	)
}

func flagsReleaseHelmfile(output bool) []cli.Flag {
	flags := flagsHidden()
	flags = append(flags,
		&cli.StringFlag{
			Name:    "helmfile-args",
			Usage:   "Helmfile additional arguments",
			Aliases: []string{"ha"},
			EnvVars: []string{"RMK_RELEASE_HELMFILE_ARGS"},
			Value:   "",
		},
		&cli.StringFlag{
			Name:    "helmfile-log-level",
			Usage:   "Helmfile log level severity, available: debug, info, warn, error",
			Aliases: []string{"hll"},
			EnvVars: []string{"RMK_RELEASE_HELMFILE_LOG_LEVEL"},
			Value:   "error",
		},
		&cli.StringSliceFlag{
			Name:    "selector",
			Usage:   "only run using releases that match labels. Labels can take form of foo=bar or foo!=bar",
			Aliases: []string{"l"},
			EnvVars: []string{"RMK_RELEASE_SELECTOR"},
		},
		&cli.BoolFlag{
			Name:    "skip-context-switch",
			Usage:   "skip context switch for not provisioned cluster",
			Aliases: []string{"s"},
		},
	)

	if output {
		flags = append(flags,
			&cli.StringFlag{
				Name:    "output",
				Usage:   "output format, available: short, yaml",
				Aliases: []string{"o"},
				EnvVars: []string{"RMK_RELEASE_OUTPUT"},
				Value:   "short",
			},
		)
	}

	return flags
}

func flagsReleaseRollback() []cli.Flag {
	return append(flagsHidden(),
		&cli.BoolFlag{
			Name:    "skip-context-switch",
			Usage:   "skip context switch for not provisioned cluster",
			Aliases: []string{"s"},
		},
		&cli.StringSliceFlag{
			Name:     "release-name",
			Usage:    "list release names for rollback status in Kubernetes",
			Aliases:  []string{"rn"},
			Required: true,
		},
	)
}

func flagsReleaseUpdate() []cli.Flag {
	return append(flagsHidden(),
		&cli.BoolFlag{
			Name:    "commit",
			Usage:   "only commit and push changes for releases file",
			Aliases: []string{"c"},
		},
		&cli.BoolFlag{
			Name:    "deploy",
			Usage:   "deploy updated releases after committed and pushed changes",
			Aliases: []string{"d"},
		},
		&cli.StringFlag{
			Name:     "repository",
			Usage:    "specific repository for updating releases file",
			Aliases:  []string{"r"},
			Required: true,
			EnvVars:  []string{"RMK_RELEASE_UPDATE_REPOSITORY"},
		},
		&cli.BoolFlag{
			Name:    "skip-ci",
			Usage:   "add [skip ci] to commit message line to skip triggering other CI builds",
			Aliases: []string{"i"},
		},
		&cli.BoolFlag{
			Name:    "skip-context-switch",
			Usage:   "skip context switch for not provisioned cluster",
			Aliases: []string{"s"},
		},
		&cli.StringFlag{
			Name:     "tag",
			Usage:    "specific tag for updating releases file",
			Aliases:  []string{"t"},
			Required: true,
			EnvVars:  []string{"RMK_RELEASE_UPDATE_TAG"},
		},
	)
}

func flagsSecretGenerate() []cli.Flag {
	return append(flagsSecretManager(),
		&cli.BoolFlag{
			Name:    "force",
			Usage:   "force overwriting current secrets after generating new",
			Aliases: []string{"f"},
		},
	)
}

func flagsSecretManager() []cli.Flag {
	return append(flagsHidden(),
		&cli.StringSliceFlag{
			Name:    "scope",
			Usage:   "specific scopes for selected secrets",
			Aliases: []string{"s"},
			EnvVars: []string{"RMK_SECRET_MANAGER_SCOPE"},
		},
		&cli.StringSliceFlag{
			Name:    "environment",
			Usage:   "specific environments for selected secrets",
			Aliases: []string{"e"},
			EnvVars: []string{"RMK_SECRET_MANAGER_ENVIRONMENT"},
		},
	)
}

func flagsUpdate() []cli.Flag {
	return append(flagsHidden(),
		&cli.BoolFlag{
			Name:    "release-candidate",
			Usage:   "force update RMK to latest release candidate version",
			Aliases: []string{"r"},
		},
		&cli.StringFlag{
			Name:        "version",
			Usage:       "RMK special version.",
			Aliases:     []string{"v"},
			DefaultText: "empty value corresponds latest version",
			EnvVars:     []string{"RMK_UPDATE_VERSION"},
		},
	)
}

func flagsHidden() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:   "config",
			Hidden: true,
		},
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:   "artifact-mode",
				Hidden: true,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:   "github-token",
				Hidden: true,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:   "cloudflare-token",
				Hidden: true,
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:   "cluster-provider",
				Hidden: true,
			},
		),
	}
}

func FlagsGlobal() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:    "log-format",
			Usage:   "log output format, available: console, json",
			Aliases: []string{"lf"},
			EnvVars: []string{"RMK_LOG_FORMAT"},
			Value:   "console",
		},
		&cli.StringFlag{
			Name:    "log-level",
			Usage:   "log level severity, available: debug, info, error",
			Aliases: []string{"ll"},
			EnvVars: []string{"RMK_LOG_LEVEL"},
			Value:   "info",
		},
	}
}
