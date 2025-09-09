package cmd

import (
	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"

	"rmk/util"
)

func flagsConfig() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Category: awsFlagsCategory,
			Name:     "aws-access-key-id",
			Usage:    "AWS access key ID for IAM user",
			Aliases:  []string{"awid"},
			EnvVars:  []string{"RMK_AWS_ACCESS_KEY_ID", "AWS_ACCESS_KEY_ID"},
		},
		&cli.StringFlag{
			Category: awsFlagsCategory,
			Name:     "aws-region",
			Usage:    "AWS region for current AWS account",
			Aliases:  []string{"awr"},
			EnvVars:  []string{"RMK_AWS_REGION", "AWS_REGION", "AWS_DEFAULT_REGION"},
		},
		&cli.StringFlag{
			Category: awsFlagsCategory,
			Name:     "aws-secret-access-key",
			Usage:    "AWS secret access key for IAM user",
			Aliases:  []string{"awsk"},
			EnvVars:  []string{"RMK_AWS_SECRET_ACCESS_KEY", "AWS_SECRET_ACCESS_KEY"},
		},
		&cli.StringFlag{
			Category: awsFlagsCategory,
			Name:     "aws-session-token",
			Usage:    "AWS session token for IAM user",
			Aliases:  []string{"awst"},
			EnvVars:  []string{"RMK_AWS_SESSION_TOKEN", "AWS_SESSION_TOKEN"},
		},
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
		&cli.StringFlag{
			Category: azureFlagsCategory,
			Name:     "azure-client-id",
			Usage:    "Azure client ID for Service Principal",
			Aliases:  []string{"azid"},
			EnvVars:  []string{"RMK_AZURE_CLIENT_ID", "AZURE_CLIENT_ID"},
		},
		&cli.StringFlag{
			Category: azureFlagsCategory,
			Name:     "azure-client-secret",
			Usage:    "Azure client secret for Service Principal",
			Aliases:  []string{"azp"},
			EnvVars:  []string{"RMK_AZURE_CLIENT_SECRET", "AZURE_CLIENT_SECRET"},
		},
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Category: azureFlagsCategory,
				Name:     "azure-key-vault-resource-group-name",
				Usage:    "Azure Key Vault custom resource group name",
				Aliases:  []string{"azkvrg"},
				EnvVars:  []string{"RMK_AZURE_KEY_VAULT_RESOURCE_GROUP_NAME", "AZURE_KEY_VAULT_RESOURCE_GROUP_NAME"},
			},
		),
		&cli.StringFlag{
			Category: azureFlagsCategory,
			Name:     "azure-location",
			Usage:    "Azure location",
			Aliases:  []string{"azl"},
			EnvVars:  []string{"RMK_AZURE_LOCATION", "AZURE_LOCATION"},
		},
		&cli.BoolFlag{
			Category: azureFlagsCategory,
			Name:     "azure-service-principle",
			Usage:    "Azure service principal STDIN content",
			Aliases:  []string{"azsp"},
		},
		&cli.StringFlag{
			Category: azureFlagsCategory,
			Name:     "azure-subscription-id",
			Usage:    "Azure subscription ID for current platform domain",
			Aliases:  []string{"azs"},
			EnvVars:  []string{"RMK_AZURE_SUBSCRIPTION_ID", "AZURE_SUBSCRIPTION_ID"},
		},
		&cli.StringFlag{
			Category: azureFlagsCategory,
			Name:     "azure-tenant-id",
			Usage:    "Azure tenant ID for Service Principal",
			Aliases:  []string{"azt"},
			EnvVars:  []string{"RMK_AZURE_TENANT_ID", "AZURE_TENANT_ID"},
		},
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "cluster-provider",
				Usage:   "cluster provider for provisioning",
				Aliases: []string{"cp"},
				EnvVars: []string{"RMK_CLUSTER_PROVIDER"},
				Value:   util.LocalClusterProvider,
			},
		),
		&cli.StringFlag{
			Name:   "config",
			Hidden: true,
		},
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Name:    "github-token",
				Usage:   "GitHub personal access token, required when using private repositories",
				Aliases: []string{"ght"},
				EnvVars: []string{"RMK_GITHUB_TOKEN"},
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Category: gcpFlagsCategory,
				Name:     "gcp-region",
				Usage:    "GCP region",
				Aliases:  []string{"gr"},
				EnvVars:  []string{"RMK_GCP_REGION", "GCP_REGION"},
			},
		),
		&cli.StringFlag{
			Category: gcpFlagsCategory,
			Name:     "google-application-credentials",
			Usage:    "path to GCP service account credentials JSON file",
			Aliases:  []string{"gac"},
			EnvVars:  []string{"RMK_GOOGLE_APPLICATION_CREDENTIALS", "GOOGLE_APPLICATION_CREDENTIALS"},
		},
		&cli.StringFlag{
			Category: onPremFlagsCategory,
			Name:     "onprem-kube-api-endpoint",
			Usage:    "Kubernetes API server endpoint used to retrieve kubeconfig",
			Aliases:  []string{"opkae"},
			EnvVars:  []string{"RMK_ONPREM_KUBE_API_ENDPOINT"},
		},
		&cli.StringFlag{
			Category: onPremFlagsCategory,
			Name:     "onprem-ssh-private-key",
			Usage:    "path to SSH private key. If not set, RMK will search in default SSH locations (e.g., ~/.ssh/id_[ed25519|rsa|ecdsa|dsa])",
			Aliases:  []string{"opspk"},
			EnvVars:  []string{"RMK_ONPREM_SSH_PRIVATE_KEY"},
		},
		&cli.StringFlag{
			Category: onPremFlagsCategory,
			Name:     "onprem-ssh-user",
			Usage:    "username to use for SSH authentication",
			Aliases:  []string{"opsu"},
			EnvVars:  []string{"RMK_ONPREM_SSH_USER"},
		},
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Name:    "progress-bar",
				Usage:   "globally disable or enable progress bar for download process",
				Aliases: []string{"p"},
				Value:   true,
			},
		),
		altsrc.NewBoolFlag(
			&cli.BoolFlag{
				Category: "Slack notifications",
				Name:     "slack-notifications",
				Usage:    "enable Slack notifications",
				Aliases:  []string{"n"},
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Category: "Slack notifications",
				Name:     "slack-webhook",
				Usage:    "URL for Slack webhook",
				Aliases:  []string{"sw"},
				EnvVars:  []string{"RMK_SLACK_WEBHOOK"},
			},
		),
		altsrc.NewStringFlag(
			&cli.StringFlag{
				Category: "Slack notifications",
				Name:     "slack-channel",
				Usage:    "channel name for Slack notifications",
				Aliases:  []string{"sc"},
				EnvVars:  []string{"RMK_SLACK_CHANNEL"},
			},
		),
		altsrc.NewStringSliceFlag(
			&cli.StringSliceFlag{
				Category: "Slack notifications",
				Name:     "slack-message-details",
				Usage:    "additional details for Slack message body",
				Aliases:  []string{"smd"},
				EnvVars:  []string{"RMK_SLACK_MESSAGE_DETAILS"},
			},
		),
	}
}

func flagsConfigList() []cli.Flag {
	return append(flagsHidden(),
		&cli.BoolFlag{
			Name:    "all",
			Usage:   "list all project configurations",
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
			Usage:    "list of images to import into running K3D cluster",
			Aliases:  []string{"ki"},
			EnvVars:  []string{"RMK_K3D_IMPORT_IMAGE"},
			Required: true,
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
		&cli.StringSliceFlag{
			Name:    "environment",
			Usage:   "list of project environments, root-domain config option must be provided: <environment>.root-domain=<domain-name>",
			Aliases: []string{"e"},
		},
		&cli.StringSliceFlag{
			Name:    "owner",
			Usage:   "list of project owners",
			Aliases: []string{"o"},
		},
		&cli.StringSliceFlag{
			Name:    "scope",
			Usage:   "list of project scopes",
			Aliases: []string{"s"},
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
			Usage:   "list of release labels, used as selector, selector can take form of foo=bar or foo!=bar",
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
			Name:    "environment",
			Usage:   "list of secret environments, used as selector",
			Aliases: []string{"e"},
			EnvVars: []string{"RMK_SECRET_MANAGER_ENVIRONMENT"},
		},
		&cli.StringSliceFlag{
			Name:    "scope",
			Usage:   "list of secret scopes, used as selector",
			Aliases: []string{"s"},
			EnvVars: []string{"RMK_SECRET_MANAGER_SCOPE"},
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
			Usage:       "RMK special version",
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
				Name:   "github-token",
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
