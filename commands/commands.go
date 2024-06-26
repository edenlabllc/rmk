package commands

import (
	"sort"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"go.uber.org/zap"

	"rmk/aws_provider"
	"rmk/config"
	"rmk/git_handler"
	"rmk/system"
)

type Flags map[string][]cli.Flag

func Commands() []*cli.Command {
	conf := &config.Config{AwsConfigure: &aws_provider.AwsConfigure{}}
	gitSpec := &git_handler.GitSpec{
		DefaultBranches: []string{
			git_handler.DefaultDevelop,
			git_handler.DefaultStaging,
			git_handler.DefaultProduction,
		},
	}
	flags := Flags{
		"clusterCRLogin":            flagsClusterCRLogin(),
		"clusterK3DCreate":          flagsClusterK3DCreate(),
		"clusterK3DImport":          flagsClusterK3DImport(),
		"clusterPlan":               flagsClusterPlan(),
		"clusterStateDelete":        flagsClusterStateDelete(),
		"clusterSwitch":             flagsClusterSwitch(),
		"config":                    flagsConfig(),
		"configList":                flagsConfigList(),
		"hidden":                    flagsHidden(),
		"projectGenerate":           flagsProjectGenerate(),
		"projectUpdate":             flagsProjectUpdate(),
		"releaseHelmfile":           flagsReleaseHelmfile(false),
		"releaseHelmfileWithOutput": flagsReleaseHelmfile(true),
		"releaseRollback":           flagsReleaseRollback(),
		"releaseUpdate":             flagsReleaseUpdate(),
		"secretGenerate":            flagsSecretGenerate(),
		"secretManager":             flagsSecretManager(),
		"update":                    flagsUpdate(),
	}

	for key := range flags {
		sort.Sort(cli.FlagsByName(flags[key]))
	}

	return []*cli.Command{
		{
			Name:  "completion",
			Usage: "Completion management",
			Subcommands: []*cli.Command{
				{
					Name:        "zsh",
					Usage:       "View Zsh completion scripts",
					Description: system.CompletionZshDescription,
					Aliases:     []string{"z"},
					Category:    "completion",
					Action:      completionAction(),
				},
			},
		},
		{
			Name:  "config",
			Usage: "Configuration management",
			Subcommands: []*cli.Command{
				{
					Name:     "init",
					Usage:    "Initialize configuration for current tenant and selected environment",
					Aliases:  []string{"i"},
					Before:   initInputSourceWithContext(gitSpec, flags["config"]),
					Flags:    flags["config"],
					Category: "config",
					Action:   configInitAction(conf, gitSpec),
				},
				{
					Name:         "delete",
					Usage:        "Delete configuration for selected environment",
					Aliases:      []string{"d"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "config",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       configDeleteAction(conf),
				},
				{
					Name:         "list",
					Usage:        "List available configurations for current tenant",
					Aliases:      []string{"l"},
					Flags:        flags["configList"],
					Category:     "config",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       configListAction(conf, gitSpec),
				},
				{
					Name:         "view",
					Usage:        "View configuration for selected environment",
					Aliases:      []string{"v"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "config",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       configViewAction(conf),
				},
			},
		},
		{
			Name:  "cluster",
			Usage: "Cluster management",
			Subcommands: []*cli.Command{
				{
					Name:     "container-registry",
					Usage:    "Container registry management",
					Aliases:  []string{"c"},
					Category: "cluster",
					Subcommands: []*cli.Command{
						{
							Name:         "login",
							Usage:        "Log in to container registry",
							Before:       readInputSourceWithContext(gitSpec, conf, flags["clusterCRLogin"]),
							Flags:        flags["clusterCRLogin"],
							Category:     "container-registry",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       containerRegistryAction(conf, DockerRunner.dockerLogin),
						},
						{
							Name:         "logout",
							Usage:        "Log out from container registry",
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "container-registry",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       containerRegistryAction(conf, DockerRunner.dockerLogout),
						},
					},
				},
				{
					Name:         "destroy",
					Usage:        "Destroy AWS cluster using Terraform",
					Aliases:      []string{"d"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "cluster",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       clusterDestroyAction(conf),
				},
				{
					Name:         "list",
					Usage:        "List all Terraform available workspaces",
					Aliases:      []string{"l"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "cluster",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       clusterListAction(conf),
				},
				{
					Name:     "k3d",
					Usage:    "K3D cluster management",
					Aliases:  []string{"k"},
					Category: "cluster",
					Subcommands: []*cli.Command{
						{
							Name:         "create",
							Usage:        "Create K3D cluster",
							Aliases:      []string{"c"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["clusterK3DCreate"]),
							Flags:        flags["clusterK3DCreate"],
							Category:     "k3d",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       K3DCreateAction(conf),
						},
						{
							Name:         "delete",
							Usage:        "Delete K3D cluster",
							Aliases:      []string{"d"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "k3d",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       K3DAction(conf, K3DRunner.createDeleteK3DCluster),
						},
						{
							Name:         "import",
							Usage:        "Import images from docker to K3D cluster",
							Aliases:      []string{"i"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["clusterK3DImport"]),
							Flags:        flags["clusterK3DImport"],
							Category:     "k3d",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       K3DAction(conf, K3DRunner.importImageToK3DCluster),
						},
						{
							Name:         "list",
							Usage:        "List K3D clusters",
							Aliases:      []string{"l"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "k3d",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       K3DAction(conf, K3DRunner.listK3DClusters),
						},
						{
							Name:         "start",
							Usage:        "Start K3D cluster",
							Aliases:      []string{"s"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							BashComplete: system.ShellCompleteCustomOutput,
							Category:     "k3d",
							Action:       K3DAction(conf, K3DRunner.startStopK3DCluster),
						},
						{
							Name:         "stop",
							Usage:        "Stop K3D cluster",
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "k3d",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       K3DAction(conf, K3DRunner.startStopK3DCluster),
						},
					},
				},
				{
					Name:         "provision",
					Usage:        "Provision AWS cluster using Terraform",
					Aliases:      []string{"p"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["clusterPlan"]),
					Flags:        flags["clusterPlan"],
					Category:     "cluster",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       clusterProvisionAction(conf),
				},
				{
					Name:     "state",
					Usage:    "State cluster management using Terraform",
					Aliases:  []string{"t"},
					Category: "cluster",
					Subcommands: []*cli.Command{
						{
							Name:         "delete",
							Usage:        "Delete resource from Terraform state",
							Aliases:      []string{"d"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["clusterStateDelete"]),
							Flags:        flags["clusterStateDelete"],
							Category:     "state",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       clusterStateAction(conf, StateRunner.clusterStateDelete),
						},
						{
							Name:         "list",
							Usage:        "List resources from Terraform state",
							Aliases:      []string{"l"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "state",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       clusterStateAction(conf, StateRunner.clusterStateList),
						},
						{
							Name:         "refresh",
							Usage:        "Update state file for AWS cluster using Terraform",
							Aliases:      []string{"r"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "state",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       clusterStateAction(conf, StateRunner.clusterStateRefresh),
						},
					},
				},
				{
					Name:         "switch",
					Usage:        "Switch Kubernetes context for tenant cluster",
					Aliases:      []string{"s"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["clusterSwitch"]),
					Flags:        flags["clusterSwitch"],
					Category:     "cluster",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       clusterSwitchAction(conf),
				},
			},
		},
		{
			Name:  "doc",
			Usage: "Documentation management",
			Subcommands: []*cli.Command{
				{
					Name:     "generate",
					Usage:    "Generate documentation by commands and flags in Markdown format",
					Aliases:  []string{"g"},
					Category: "doc",
					Action:   docGenerateAction(),
				},
			},
		},
		{
			Name:  "project",
			Usage: "Project management",
			Subcommands: []*cli.Command{
				{
					Name:         "generate",
					Usage:        "Generate project directories and files structure",
					Aliases:      []string{"g"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["projectGenerate"]),
					Flags:        flags["projectGenerate"],
					Category:     "project",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       projectGenerateAction(conf, gitSpec),
				},
				{
					Name:         "update",
					Usage:        "Update project file with specific dependencies version",
					Aliases:      []string{"u"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["projectUpdate"]),
					Flags:        flags["projectUpdate"],
					Category:     "project",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       projectUpdateAction(conf, gitSpec),
				},
			},
		},
		{
			Name:  "release",
			Usage: "Release components list from state file (Helmfile)",
			Subcommands: []*cli.Command{
				{
					Name:         "build",
					Usage:        "Build releases",
					Aliases:      []string{"b"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseHelmfile"]),
					Flags:        flags["releaseHelmfile"],
					Category:     "release",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       releaseHelmfileAction(conf),
				},
				{
					Name:         "destroy",
					Usage:        "Destroy releases",
					Aliases:      []string{"d"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseHelmfile"]),
					Flags:        flags["releaseHelmfileWithOutput"],
					Category:     "release",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       releaseHelmfileAction(conf),
				},
				{
					Name:         "list",
					Usage:        "List releases",
					Aliases:      []string{"l"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseHelmfileWithOutput"]),
					Flags:        flags["releaseHelmfileWithOutput"],
					Category:     "release",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       releaseHelmfileAction(conf),
				},
				{
					Name:         "rollback",
					Usage:        "Rollback specific releases to latest stable state",
					Aliases:      []string{"r"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseRollback"]),
					Flags:        flags["releaseRollback"],
					Category:     "release",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       releaseRollbackAction(conf),
				},
				{
					Name:         "sync",
					Usage:        "Sync releases",
					Aliases:      []string{"s"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseHelmfile"]),
					Flags:        flags["releaseHelmfile"],
					Category:     "release",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       releaseHelmfileAction(conf),
				},
				{
					Name:         "template",
					Usage:        "Template releases",
					Aliases:      []string{"t"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseHelmfile"]),
					Flags:        flags["releaseHelmfile"],
					Category:     "release",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       releaseHelmfileAction(conf),
				},
				{
					Name:         "update",
					Usage:        "Update releases file with specific environment values",
					Aliases:      []string{"u"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseUpdate"]),
					Flags:        flags["releaseUpdate"],
					Category:     "release",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       releaseUpdateAction(conf, gitSpec),
				},
			},
		},
		{
			Name:  "secret",
			Usage: "secrets management",
			Subcommands: []*cli.Command{
				{
					Name:     "manager",
					Usage:    "batch secrets management",
					Aliases:  []string{"m"},
					Category: "secret",
					Subcommands: []*cli.Command{
						{
							Name:         "decrypt",
							Usage:        "Decrypt secrets batch for selected scope and environment",
							Aliases:      []string{"d"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["secretManager"]),
							Flags:        flags["secretManager"],
							Category:     "manager",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       secretMgrEncryptDecryptAction(conf),
						},
						{
							Name:         "encrypt",
							Usage:        "Encrypt secrets batch for selected scope and environment",
							Aliases:      []string{"e"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["secretManager"]),
							Flags:        flags["secretManager"],
							Category:     "manager",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       secretMgrEncryptDecryptAction(conf),
						},
						{
							Name:         "generate",
							Usage:        "Generate secrets batch for selected scope and environment",
							Aliases:      []string{"g"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["secretGenerate"]),
							Flags:        flags["secretGenerate"],
							Category:     "manager",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       secretMgrGenerateAction(conf),
						},
					},
				},
				{
					Name:     "keys",
					Usage:    "SOPS age keys management",
					Aliases:  []string{"k"},
					Category: "secret",
					Subcommands: []*cli.Command{
						{
							Name:         "create",
							Usage:        "Create SOPS age keys",
							Aliases:      []string{"c"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "keys",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       secretKeysCreateAction(conf),
						},
						{
							Name:         "download",
							Usage:        "Download SOPS age keys from S3 bucket",
							Aliases:      []string{"d"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "keys",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       secretKeysDownloadAction(conf),
						},
						{
							Name:         "upload",
							Usage:        "Upload SOPS age keys to S3 bucket",
							Aliases:      []string{"u"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "keys",
							BashComplete: system.ShellCompleteCustomOutput,
							Action:       secretKeysUploadAction(conf),
						},
					},
				},
				{
					Name:         "encrypt",
					Usage:        "Encrypt secret file",
					Aliases:      []string{"e"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "secret",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       secretAction(conf, SecretRunner.helmSecretsEncrypt),
				},
				{
					Name:         "decrypt",
					Usage:        "Decrypt secret file",
					Aliases:      []string{"d"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "secret",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       secretAction(conf, SecretRunner.helmSecretsDecrypt),
				},
				{
					Name:         "view",
					Usage:        "View secret file",
					Aliases:      []string{"v"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "secret",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       secretAction(conf, SecretRunner.helmSecretsView),
				},
				{
					Name:         "edit",
					Usage:        "Edit secret file",
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "secret",
					BashComplete: system.ShellCompleteCustomOutput,
					Action:       secretAction(conf, SecretRunner.helmSecretsEdit),
				},
			},
		},
		{
			Name:         "update",
			Usage:        "Update RMK CLI to a new version",
			Flags:        flags["update"],
			BashComplete: system.ShellCompleteCustomOutput,
			Action:       updateAction(),
		},
	}
}

func initInputSourceWithContext(gitSpec *git_handler.GitSpec, flags []cli.Flag) cli.BeforeFunc {
	return func(ctx *cli.Context) error {
		if err := gitSpec.GenerateID(); err != nil {
			return err
		}

		return inputSourceContext(ctx, flags, system.GetHomePath(system.RMKDir, system.RMKConfig, gitSpec.ID+".yaml"))
	}
}

func readInputSourceWithContext(gitSpec *git_handler.GitSpec, conf *config.Config, flags []cli.Flag) cli.BeforeFunc {
	return func(ctx *cli.Context) error {
		if err := gitSpec.GenerateID(); err != nil {
			return err
		}

		configPath := system.GetHomePath(system.RMKDir, system.RMKConfig, gitSpec.ID+".yaml")
		if err := conf.ReadConfigFile(configPath); err != nil {
			zap.S().Errorf(system.ConfigNotInitializedErrorText)
			return err
		}

		return inputSourceContext(ctx, flags, configPath)
	}
}

func inputSourceContext(ctx *cli.Context, flags []cli.Flag, configPath string) error {
	createInputSource := func(ctx *cli.Context) (altsrc.InputSourceContext, error) {
		if ctx.IsSet("config") {
			filePath := ctx.String("config")
			return altsrc.NewYamlSourceFromFile(filePath)
		}

		if err := ctx.Set("config", configPath); err != nil {
			return nil, err
		}

		if system.IsExists(configPath, true) {
			return altsrc.NewYamlSourceFromFile(configPath)
		} else {
			return &altsrc.MapInputSource{}, nil
		}
	}

	inputSource, err := createInputSource(ctx)
	if err != nil {
		return err
	}

	return altsrc.ApplyInputSourceValues(ctx, inputSource, flags)
}
