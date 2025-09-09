package cmd

import (
	"fmt"
	"sort"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
	"go.uber.org/zap"

	"rmk/config"
	"rmk/git_handler"
	"rmk/util"
)

type Flags map[string][]cli.Flag

func Commands() []*cli.Command {
	conf := &config.Config{}
	gitSpec := &git_handler.GitSpec{
		DefaultBranches: []string{
			git_handler.DefaultDevelop,
			git_handler.DefaultStaging,
			git_handler.DefaultProduction,
		},
	}
	flags := Flags{
		"clusterK3DCreate":          flagsClusterK3DCreate(),
		"clusterK3DImport":          flagsClusterK3DImport(),
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
					Description: util.CompletionZshDescription,
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
					Name:         "init",
					Usage:        "Initialize configuration for current project and selected environment",
					Aliases:      []string{"i"},
					Before:       initInputSourceWithContext(gitSpec, flags["config"]),
					Flags:        flags["config"],
					Category:     "config",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       configInitAction(conf, gitSpec),
				},
				{
					Name:         "delete",
					Usage:        "Delete configuration for selected environment",
					Aliases:      []string{"d"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "config",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       configDeleteAction(conf),
				},
				{
					Name:         "list",
					Usage:        "List available configurations for current project",
					Aliases:      []string{"l"},
					Flags:        flags["configList"],
					Category:     "config",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       configListAction(conf, gitSpec),
				},
				{
					Name:         "view",
					Usage:        "View configuration for selected environment",
					Aliases:      []string{"v"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "config",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       configViewAction(conf),
				},
			},
		},
		{
			Name:  "cluster",
			Usage: "Cluster management",
			Subcommands: []*cli.Command{
				{
					Name:     "capi",
					Usage:    "CAPI cluster management",
					Aliases:  []string{"c"},
					Category: "cluster",
					Subcommands: []*cli.Command{
						{
							Name:         "create",
							Usage:        "Create CAPI management cluster",
							Aliases:      []string{"c"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "capi",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       K3DCreateAction(conf),
							After:        CAPIInitAction(conf, gitSpec),
						},
						{
							Name:         "delete",
							Usage:        "Delete CAPI management cluster",
							Aliases:      []string{"d"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "capi",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       K3DAction(conf, K3DRunner.createDeleteK3DCluster),
						},
						{
							Name:         "destroy",
							Usage:        "Destroy K8S target (workload) cluster",
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "capi",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       CAPIProvisionDestroyAction(conf),
						},
						{
							Name:         "list",
							Usage:        "List CAPI management clusters",
							Aliases:      []string{"l"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "capi",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       K3DAction(conf, K3DRunner.listK3DClusters),
						},
						{
							Name:         "provision",
							Usage:        "Provision K8S target (workload) cluster",
							Aliases:      []string{"p"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "capi",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       CAPIProvisionDestroyAction(conf),
						},
						{
							Name:         "update",
							Usage:        "Update CAPI management cluster",
							Aliases:      []string{"u"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "capi",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       CAPIUpdateAction(conf),
						},
					},
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
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       K3DCreateAction(conf),
						},
						{
							Name:         "delete",
							Usage:        "Delete K3D cluster",
							Aliases:      []string{"d"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "k3d",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       K3DAction(conf, K3DRunner.createDeleteK3DCluster),
						},
						{
							Name:         "import",
							Usage:        "Import images from docker to K3D cluster",
							Aliases:      []string{"i"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["clusterK3DImport"]),
							Flags:        flags["clusterK3DImport"],
							Category:     "k3d",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       K3DAction(conf, K3DRunner.importImageToK3DCluster),
						},
						{
							Name:         "list",
							Usage:        "List K3D clusters",
							Aliases:      []string{"l"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "k3d",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       K3DAction(conf, K3DRunner.listK3DClusters),
						},
						{
							Name:         "start",
							Usage:        "Start K3D cluster",
							Aliases:      []string{"s"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							BashComplete: util.ShellCompleteCustomOutput,
							Category:     "k3d",
							Action:       K3DAction(conf, K3DRunner.startStopK3DCluster),
						},
						{
							Name:         "stop",
							Usage:        "Stop K3D cluster",
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "k3d",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       K3DAction(conf, K3DRunner.startStopK3DCluster),
						},
					},
				},
				{
					Name:         "switch",
					Usage:        "Switch Kubernetes context to project cluster",
					Aliases:      []string{"s"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["clusterSwitch"]),
					Flags:        flags["clusterSwitch"],
					Category:     "cluster",
					BashComplete: util.ShellCompleteCustomOutput,
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
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       projectGenerateAction(conf, gitSpec),
				},
				{
					Name:         "update",
					Usage:        "Update project file with specific dependencies version",
					Aliases:      []string{"u"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["projectUpdate"]),
					Flags:        flags["projectUpdate"],
					Category:     "project",
					BashComplete: util.ShellCompleteCustomOutput,
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
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       releaseHelmfileAction(conf),
				},
				{
					Name:         "destroy",
					Usage:        "Destroy releases",
					Aliases:      []string{"d"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseHelmfile"]),
					Flags:        flags["releaseHelmfile"],
					Category:     "release",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       releaseHelmfileAction(conf),
				},
				{
					Name:         "list",
					Usage:        "List releases",
					Aliases:      []string{"l"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseHelmfileWithOutput"]),
					Flags:        flags["releaseHelmfileWithOutput"],
					Category:     "release",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       releaseHelmfileAction(conf),
				},
				{
					Name:         "rollback",
					Usage:        "Rollback specific releases to latest stable state",
					Aliases:      []string{"r"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseRollback"]),
					Flags:        flags["releaseRollback"],
					Category:     "release",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       releaseRollbackAction(conf),
				},
				{
					Name:         "sync",
					Usage:        "Sync releases",
					Aliases:      []string{"s"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseHelmfile"]),
					Flags:        flags["releaseHelmfile"],
					Category:     "release",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       releaseHelmfileAction(conf),
				},
				{
					Name:         "template",
					Usage:        "Template releases",
					Aliases:      []string{"t"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseHelmfile"]),
					Flags:        flags["releaseHelmfile"],
					Category:     "release",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       releaseHelmfileAction(conf),
				},
				{
					Name:         "update",
					Usage:        "Update releases file with specific environment values",
					Aliases:      []string{"u"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["releaseUpdate"]),
					Flags:        flags["releaseUpdate"],
					Category:     "release",
					BashComplete: util.ShellCompleteCustomOutput,
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
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       secretMgrEncryptDecryptAction(conf),
						},
						{
							Name:         "encrypt",
							Usage:        "Encrypt secrets batch for selected scope and environment",
							Aliases:      []string{"e"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["secretManager"]),
							Flags:        flags["secretManager"],
							Category:     "manager",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       secretMgrEncryptDecryptAction(conf),
						},
						{
							Name:         "generate",
							Usage:        "Generate secrets batch for selected scope and environment",
							Aliases:      []string{"g"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["secretGenerate"]),
							Flags:        flags["secretGenerate"],
							Category:     "manager",
							BashComplete: util.ShellCompleteCustomOutput,
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
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       secretKeysCreateAction(conf),
						},
						{
							Name:         "download",
							Usage:        "Download SOPS age keys from S3 bucket",
							Aliases:      []string{"d"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "keys",
							BashComplete: util.ShellCompleteCustomOutput,
							Action:       secretKeysDownloadAction(conf),
						},
						{
							Name:         "upload",
							Usage:        "Upload SOPS age keys to S3 bucket",
							Aliases:      []string{"u"},
							Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
							Flags:        flags["hidden"],
							Category:     "keys",
							BashComplete: util.ShellCompleteCustomOutput,
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
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       secretAction(conf, SecretRunner.helmSecretsEncrypt),
				},
				{
					Name:         "decrypt",
					Usage:        "Decrypt secret file",
					Aliases:      []string{"d"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "secret",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       secretAction(conf, SecretRunner.helmSecretsDecrypt),
				},
				{
					Name:         "view",
					Usage:        "View secret file",
					Aliases:      []string{"v"},
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "secret",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       secretAction(conf, SecretRunner.helmSecretsView),
				},
				{
					Name:         "edit",
					Usage:        "Edit secret file",
					Before:       readInputSourceWithContext(gitSpec, conf, flags["hidden"]),
					Flags:        flags["hidden"],
					Category:     "secret",
					BashComplete: util.ShellCompleteCustomOutput,
					Action:       secretAction(conf, SecretRunner.helmSecretsEdit),
				},
			},
		},
		{
			Name:         "update",
			Usage:        "Update RMK CLI to a new version",
			Flags:        flags["update"],
			BashComplete: util.ShellCompleteCustomOutput,
			Action:       updateAction(),
		},
	}
}

func initInputSourceWithContext(gitSpec *git_handler.GitSpec, flags []cli.Flag) cli.BeforeFunc {
	return func(ctx *cli.Context) error {
		if err := gitSpec.GenerateID(); err != nil {
			return err
		}

		return inputSourceContext(ctx, flags, util.GetHomePath(util.RMKDir, util.RMKConfig, gitSpec.ID+".yaml"))
	}
}

func readInputSourceWithContext(gitSpec *git_handler.GitSpec, conf *config.Config, flags []cli.Flag) cli.BeforeFunc {
	return func(ctx *cli.Context) error {
		if err := gitSpec.GenerateID(); err != nil {
			return err
		}

		configPath := util.GetHomePath(util.RMKDir, util.RMKConfig, gitSpec.ID+".yaml")
		detectPrGenCommand := ctx.Command.Category == "project" && ctx.Command.Name == "generate"
		isProjectDefined := detectPrGenCommand && util.IsExists(util.GetPwdPath(util.TenantProjectFile), true)
		isProjectUndefined := detectPrGenCommand && !util.IsExists(util.GetPwdPath(util.TenantProjectFile), true)

		if !util.IsExists(configPath, true) {
			if isProjectUndefined {
				return nil
			} else if isProjectDefined {
				return fmt.Errorf(
					"%s file exists, please eather delete it or run 'rmk config init' command to regenerate project",
					util.TenantProjectFile)
			}
		}

		if err := conf.ReadConfigFile(configPath); err != nil {
			zap.S().Errorf(util.ConfigNotInitializedErrorText)
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

		if util.IsExists(configPath, true) {
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
