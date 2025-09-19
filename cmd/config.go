package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"rmk/config"
	"rmk/git_handler"
	"rmk/providers/aws_provider"
	"rmk/providers/azure_provider"
	"rmk/providers/google_provider"
	"rmk/providers/onprem_provider"
	"rmk/util"
)

type ConfigCommands struct {
	*ReleaseCommands
}

func newConfigCommands(conf *config.Config, ctx *cli.Context, workDir string) *ConfigCommands {
	return &ConfigCommands{&ReleaseCommands{Conf: conf, Ctx: ctx, WorkDir: workDir}}
}

func (c *ConfigCommands) helmPlugin() *util.SpecCMD {
	return &util.SpecCMD{
		Args:          []string{"plugin"},
		Command:       "helm",
		Dir:           c.WorkDir,
		Ctx:           c.Ctx,
		DisableStdOut: true,
		Debug:         false,
	}
}

func (c *ConfigCommands) configAws(profile string) error {
	ac := aws_provider.NewAwsConfigure(c.Ctx.Context, profile)
	ac.ConfigSource = strings.Join(ac.AWSSharedConfigFile(profile), "")
	ac.CredentialsSource = strings.Join(ac.AWSSharedCredentialsFile(profile), "")

	if !util.IsExists(c.Ctx.String("config"), true) &&
		util.IsExists(ac.ConfigSource, true) &&
		util.IsExists(ac.CredentialsSource, true) {
		if err := os.RemoveAll(ac.ConfigSource); err != nil {
			return err
		}

		if err := os.RemoveAll(ac.CredentialsSource); err != nil {
			return err
		}
	}

	if util.IsExists(ac.ConfigSource, true) {
		if err := ac.ReadAWSConfigProfile(); err != nil {
			return err
		}
	}

	if c.Ctx.IsSet("aws-access-key-id") && c.Ctx.IsSet("aws-secret-access-key") {
		ac.AwsCredentialsProfile.AccessKeyID = c.Ctx.String("aws-access-key-id")
		ac.AwsCredentialsProfile.SecretAccessKey = c.Ctx.String("aws-secret-access-key")
	} else if c.Ctx.IsSet("aws-access-key-id") {
		ac.AwsCredentialsProfile.AccessKeyID = c.Ctx.String("aws-access-key-id")
		ac.AwsCredentialsProfile.SecretAccessKey = ""
	} else if c.Ctx.IsSet("aws-secret-access-key") {
		ac.AwsCredentialsProfile.AccessKeyID = ""
		ac.AwsCredentialsProfile.SecretAccessKey = c.Ctx.String("aws-secret-access-key")
	}

	if c.Ctx.IsSet("aws-region") {
		ac.Region = c.Ctx.String("aws-region")
	}

	if c.Ctx.IsSet("aws-session-token") {
		ac.AwsCredentialsProfile.SessionToken = c.Ctx.String("aws-session-token")
	}

	if err := ac.ValidateAWSCredentials(); err != nil {
		return err
	} else {
		c.Conf.AwsConfigure = ac
	}

	if err := ac.WriteAWSConfigProfile(); err != nil {
		return err
	}

	return nil
}

func (c *ConfigCommands) configAwsMFA() error {
	var (
		tokenExpiration                 time.Time
		regularProfile                  string
		regularProfileConfigSource      string
		regularProfileCredentialsSource string
	)
	currentTime := time.Now()

	if strings.HasSuffix(c.Conf.Profile, "-mfa") {
		regularProfile = strings.TrimSuffix(c.Conf.Profile, "-mfa")
		regularProfileConfigSource = strings.TrimSuffix(c.Conf.ConfigSource, "-mfa")
		regularProfileCredentialsSource = strings.TrimSuffix(c.Conf.CredentialsSource, "-mfa")
	} else {
		regularProfile = c.Conf.Profile
		regularProfileConfigSource = c.Conf.ConfigSource
		regularProfileCredentialsSource = c.Conf.CredentialsSource
	}

	if len(c.Conf.AWSMFATokenExpiration) > 0 {
		unixTime, err := strconv.ParseInt(c.Conf.AWSMFATokenExpiration, 10, 64)
		if err != nil {
			return err
		}

		tokenExpiration = time.Unix(unixTime, 0)
	}

	if len(c.Conf.AWSMFAProfile) > 0 {
		c.Conf.AwsConfigure.Profile = c.Conf.AWSMFAProfile
	}

	if err := c.Conf.GetAWSMFADevicesSerialNumbers(); err != nil {
		return err
	}

	timeDiff := time.Time{}.Add(tokenExpiration.Sub(currentTime)).Format("15:04:05")

	if len(c.Conf.MFADeviceSerialNumber) > 0 {
		zap.S().Infof("MFA device SerialNumber: %s", c.Conf.MFADeviceSerialNumber)
	} else {
		if len(c.Conf.AWSMFAProfile) > 0 && len(c.Conf.AWSMFATokenExpiration) > 0 {
			ac := aws_provider.NewAwsConfigure(c.Ctx.Context, c.Conf.AWSMFAProfile)
			ac.ConfigSource = strings.Join(ac.AWSSharedConfigFile(c.Conf.AWSMFAProfile), "")
			ac.CredentialsSource = strings.Join(ac.AWSSharedCredentialsFile(c.Conf.AWSMFAProfile), "")
			if util.IsExists(ac.ConfigSource, true) {
				if err := ac.ReadAWSConfigProfile(); err != nil {
					return err
				}
			}

			ac.Profile = strings.TrimSuffix(regularProfile, "-mfa")
			ac.ConfigSource = strings.TrimSuffix(regularProfileConfigSource, "-mfa")
			ac.CredentialsSource = strings.TrimSuffix(regularProfileCredentialsSource, "-mfa")

			if err := ac.WriteAWSConfigProfile(); err != nil {
				return err
			}

			if err := os.RemoveAll(strings.Join(c.Conf.AWSSharedConfigFile(c.Conf.AWSMFAProfile), "")); err != nil {
				return err
			}

			if err := os.RemoveAll(strings.Join(c.Conf.AWSSharedCredentialsFile(c.Conf.AWSMFAProfile), "")); err != nil {
				return err
			}

			c.Conf.AWSMFAProfile = ""
			c.Conf.AWSMFATokenExpiration = ""
			c.Conf.AwsConfigure.Profile = ac.Profile
			c.Conf.AwsConfigure.ConfigSource = ac.ConfigSource
			c.Conf.AwsConfigure.CredentialsSource = ac.CredentialsSource

			return nil
		}
	}

	if currentTime.Before(tokenExpiration) {
		zap.S().Infof("MFA remaining time for token validity: %s", timeDiff)
	}

	if len(c.Conf.MFADeviceSerialNumber) > 0 && currentTime.After(tokenExpiration) {
		if err := c.Conf.GetAWSMFASessionToken(); err != nil {
			return err
		}

		c.Conf.AWSMFAProfile = regularProfile + "-mfa"
		c.Conf.AWSMFATokenExpiration = strconv.FormatInt(c.Conf.Expiration.Unix(), 10)

		acMFA := aws_provider.NewAwsConfigure(c.Ctx.Context, regularProfile+"-mfa")
		acMFA.AwsCredentialsProfile.AccessKeyID = c.Conf.MFAProfileCredentials.AccessKeyID
		acMFA.AwsCredentialsProfile.SecretAccessKey = c.Conf.MFAProfileCredentials.SecretAccessKey
		acMFA.ConfigSource = regularProfileConfigSource + "-mfa"
		acMFA.CredentialsSource = regularProfileCredentialsSource + "-mfa"
		acMFA.Region = c.Conf.Region

		if err := acMFA.WriteAWSConfigProfile(); err != nil {
			return err
		}

		acRegular := aws_provider.NewAwsConfigure(c.Ctx.Context, regularProfile)
		acRegular.AwsCredentialsProfile.AccessKeyID = c.Conf.MFAToken.AccessKeyID
		acRegular.AwsCredentialsProfile.SecretAccessKey = c.Conf.MFAToken.SecretAccessKey
		acRegular.AwsCredentialsProfile.SessionToken = c.Conf.MFAToken.SessionToken
		acRegular.ConfigSource = regularProfileConfigSource
		acRegular.CredentialsSource = regularProfileCredentialsSource
		acRegular.Region = c.Conf.Region

		if err := acRegular.WriteAWSConfigProfile(); err != nil {
			return err
		}
	}

	c.Conf.AwsConfigure.Profile = strings.TrimSuffix(regularProfile, "-mfa")
	c.Conf.AwsConfigure.ConfigSource = strings.TrimSuffix(regularProfileConfigSource, "-mfa")
	c.Conf.AwsConfigure.CredentialsSource = strings.TrimSuffix(regularProfileCredentialsSource, "-mfa")

	return nil
}

func (c *ConfigCommands) installHelmPlugin(plugin config.Package, args ...string) error {
	c.SpecCMD = c.helmPlugin()
	c.SpecCMD.Args = append(c.SpecCMD.Args, args...)
	if err := releaseRunner(c).runCMD(); err != nil {
		if !strings.Contains(c.SpecCMD.StderrBuf.String(), util.HelmPluginExist) {
			return fmt.Errorf("Helm plugin %s installation failed: \n%s", plugin.Name, c.SpecCMD.StderrBuf.String())
		}
	}

	if !strings.Contains(c.SpecCMD.StderrBuf.String(), util.HelmPluginExist) {
		zap.S().Infof("installing Helm plugin: %s", plugin.Name)
	}

	return nil
}

func (c *ConfigCommands) configHelmPlugins() error {
	var (
		helmPluginsUpdate    = make(map[string]*config.Package)
		helmPluginsInstalled = make(map[string]*config.Package)
	)

	c.SpecCMD = c.helmPlugin()
	c.SpecCMD.Args = append(c.SpecCMD.Args, "list")

	if err := releaseRunner(c).runCMD(); err != nil {
		return fmt.Errorf("get Helm plugin list failed: %s", c.SpecCMD.StderrBuf.String())
	}

	for _, val := range strings.Split(c.SpecCMD.StdoutBuf.String(), "\n")[1:] {
		reg, _ := regexp.Compile(`\s+`)
		plugin := strings.Split(reg.ReplaceAllString(val, "|"), "|")
		if len(plugin) > 1 {
			helmPluginsInstalled[plugin[0]] = &config.Package{
				Name:    plugin[0],
				Version: plugin[1],
			}
		}
	}

	for name, plugin := range c.Conf.HelmPlugins {
		plSemVer, _ := semver.NewVersion(plugin.Version)
		for _, pl := range helmPluginsInstalled {
			plSV, _ := semver.NewVersion(pl.Version)
			if pl.Name == plugin.Name && !plSemVer.Equal(plSV) {
				helmPluginsUpdate[name] = plugin
				break
			}
		}
	}

	for _, plugin := range helmPluginsUpdate {
		zap.S().Infof("Helm plugin %s detect new version %s from %s", plugin.Name, plugin.Version, util.TenantProjectFile)
		c.SpecCMD = c.helmPlugin()
		c.SpecCMD.Args = append(c.SpecCMD.Args, "uninstall", plugin.Name)
		if err := releaseRunner(c).runCMD(); err != nil {
			return fmt.Errorf("Helm plugin %s uninstallation failed: \n%s",
				plugin.Name, c.SpecCMD.StderrBuf.String())
		}

		if err := c.installHelmPlugin(*plugin, "install", plugin.Url, "--version="+plugin.Version); err != nil {
			return err
		}
	}

	for name, plugin := range c.Conf.HelmPlugins {
		if _, ok := helmPluginsInstalled[name]; !ok {
			if err := c.installHelmPlugin(*plugin, "install", plugin.Url, "--version="+plugin.Version); err != nil {
				return err
			}
		}
	}

	return nil
}

func initAWSProfile(c *cli.Context, conf *config.Config, gitSpec *git_handler.GitSpec) error {
	var profile string

	conf.AwsConfigure = aws_provider.NewAwsConfigure(c.Context, gitSpec.ID)
	conf.AWSMFAProfile = c.String("aws-mfa-profile")
	conf.AWSMFATokenExpiration = c.String("aws-mfa-token-expiration")

	// Detect if MFA is enabled
	if len(conf.AWSMFAProfile) > 0 && len(conf.AWSMFATokenExpiration) > 0 {
		profile = conf.AWSMFAProfile
	} else {
		profile = conf.Profile
	}

	if err := newConfigCommands(conf, c, util.GetPwdPath("")).configAws(profile); err != nil {
		return err
	}

	if ok, err := conf.AwsConfigure.GetAwsConfigure(profile); err != nil && ok {
		zap.S().Warnf("%s", err.Error())
	} else if !ok && err != nil {
		return err
	}

	if err := newConfigCommands(conf, c, util.GetPwdPath("")).configAwsMFA(); err != nil {
		return err
	}

	secrets, err := aws_provider.NewAwsConfigure(c.Context, conf.Profile).GetAWSSecrets(conf.Tenant)
	if err != nil {
		return err
	}

	return newSecretCommands(conf, c, util.GetPwdPath("")).
		WriteKeysInRootDir(secrets, "AWS Secrets Manager")
}

func initAzureProfile(c *cli.Context, conf *config.Config, gitSpec *git_handler.GitSpec) error {
	ac := azure_provider.NewAzureConfigure()
	asp := azure_provider.NewRawSP()

	if util.IsExists(
		util.GetHomePath(azure_provider.AzureHomeDir, azure_provider.AzurePrefix+gitSpec.ID+".json"), true) {
		if err := ac.ReadSPCredentials(gitSpec.ID); err != nil {
			return err
		}
	}

	if c.Bool("azure-service-principle") {
		if err := json.NewDecoder(os.Stdin).Decode(&asp); err != nil {
			return fmt.Errorf("unable to deserialize JSON from STDIN: %s", err.Error())
		}

		ac.MergeAzureRawSP(asp)
	}

	if c.IsSet("azure-client-id") {
		ac.ClientID = c.String("azure-client-id")
	}

	if c.IsSet("azure-client-secret") {
		ac.ClientSecret = c.String("azure-client-secret")
	}

	if c.IsSet("azure-location") {
		ac.Location = c.String("azure-location")
	}

	if c.IsSet("azure-subscription-id") {
		ac.SubscriptionID = c.String("azure-subscription-id")
	}

	if c.IsSet("azure-tenant-id") {
		ac.TenantID = c.String("azure-tenant-id")
	}

	if err := ac.ValidateSPCredentials(); err != nil {
		return err
	} else {
		conf.AzureConfigure = ac
	}

	if err := ac.WriteSPCredentials(gitSpec.ID); err != nil {
		return err
	}

	if err := ac.NewAzureClient(c.Context, gitSpec.ID); err != nil {
		return err
	}

	if c.IsSet("azure-key-vault-resource-group-name") {
		conf.AzureKeyVaultResourceGroup = c.String("azure-key-vault-resource-group-name")
		ac.ResourceGroupName = c.String("azure-key-vault-resource-group-name")
	} else {
		if err := ac.DefineAzureKeyVaultResourceGroup(conf.Tenant); err != nil {
			return err
		}
	}

	ok, err := ac.GetAzureKeyVault(conf.Tenant)
	if err != nil {
		return err
	}

	if !ok {
		err := ac.CreateAzureKeyVault(conf.Tenant)
		if err != nil {
			return err
		}
	} else {
		secrets, err := ac.GetAzureSecrets()
		if err != nil {
			return err
		}

		if err := newSecretCommands(conf, c, util.GetPwdPath("")).
			WriteKeysInRootDir(secrets, "Azure Key Vault"); err != nil {
			return err
		}
	}

	return nil
}

func initGCPProfile(c *cli.Context, conf *config.Config, gitSpec *git_handler.GitSpec) error {
	gcp := google_provider.NewGCPConfigure(c.Context,
		util.GetHomePath(google_provider.GoogleHomeDir, google_provider.GoogleCredentialsPrefix+gitSpec.ID+".json"))

	if c.IsSet("google-application-credentials") {
		gcp.AppCredentialsPath = c.String("google-application-credentials")
		if err := gcp.ReadSACredentials(); err != nil {
			return err
		}

		if err := gcp.CopySACredentials(gitSpec.ID); err != nil {
			return err
		}
	} else {
		if err := gcp.ReadSACredentials(); err != nil {
			return err
		}
	}

	if !c.IsSet("gcp-region") {
		return fmt.Errorf("GCP provider option gcp-region required")
	}

	conf.GCPRegion = c.String("gcp-region")
	conf.GCPConfigure = gcp

	secrets, err := gcp.GetGCPSecrets(conf.Tenant)
	if err != nil {
		return err
	}

	return newSecretCommands(conf, c, util.GetPwdPath("")).
		WriteKeysInRootDir(secrets, "GCP Secrets Manager")
}

func initOnPremProfile(c *cli.Context, conf *config.Config, gitSpec *git_handler.GitSpec) error {
	confDiff := &config.Config{}
	configPath := util.GetHomePath(util.RMKDir, util.RMKConfig, gitSpec.ID+".yaml")
	if util.IsExists(configPath, true) {
		if err := confDiff.ReadConfigFile(configPath); err != nil {
			return err
		}

		if confDiff.ClusterProvider != onprem_provider.OnPremClusterProvider || confDiff.OnPremConfigure == nil {
			conf.OnPremConfigure = onprem_provider.NewOnPremConfigure()
		} else {
			conf.OnPremConfigure = confDiff.OnPremConfigure
		}
	} else {
		conf.OnPremConfigure = onprem_provider.NewOnPremConfigure()
	}

	if c.IsSet("onprem-ssh-init-server-host") {
		conf.OnPremConfigure.SSHInitServerHost = c.String("onprem-ssh-init-server-host")
	}

	if c.IsSet("onprem-ssh-private-key") && len(c.String("onprem-ssh-private-key")) > 0 {
		sshPrivateKey := c.String("onprem-ssh-private-key")
		if err := os.MkdirAll(util.GetHomePath(".ssh"), 0700); err != nil {
			return err
		}

		sshPKHomePath := util.GetHomePath(".ssh", filepath.Base(sshPrivateKey))
		if util.IsExists(sshPKHomePath, true) {
			if ok, err := util.CheckFilesSum(sshPrivateKey, sshPKHomePath); err != nil {
				return err
			} else if !ok {
				if err := util.CopyFile(sshPrivateKey, sshPKHomePath, 0600); err != nil {
					return err
				}
			}
		} else {
			if err := util.CopyFile(sshPrivateKey, sshPKHomePath, 0600); err != nil {
				return err
			}
		}

		conf.OnPremConfigure.SSHPrivateKey = sshPKHomePath
	} else if c.IsSet("onprem-ssh-private-key") && len(c.String("onprem-ssh-private-key")) == 0 {
		conf.OnPremConfigure.SSHPrivateKey = ""
	}

	if c.IsSet("onprem-ssh-user") {
		conf.OnPremConfigure.SSHUser = c.String("onprem-ssh-user")
	}

	if err := conf.OnPremConfigure.ValidateSSHCredentials(); err != nil {
		return err
	}

	return nil
}

func configDeleteAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		switch {
		case c.String("cluster-provider") == aws_provider.AWSClusterProvider:
			// Delete MFA profile
			if len(conf.AWSMFAProfile) > 0 && len(conf.AWSMFATokenExpiration) > 0 {
				if err := os.RemoveAll(strings.Join(conf.AWSSharedConfigFile(conf.AWSMFAProfile), "")); err != nil {
					return err
				}

				if err := os.RemoveAll(strings.Join(conf.AWSSharedCredentialsFile(conf.AWSMFAProfile), "")); err != nil {
					return err
				}
			}

			// Delete config regular profile
			if err := os.RemoveAll(strings.Join(conf.AWSSharedConfigFile(conf.Profile), "")); err != nil {
				return err
			}

			// Delete credentials regular profile
			if err := os.RemoveAll(strings.Join(conf.AWSSharedCredentialsFile(conf.Profile), "")); err != nil {
				return err
			}
		case c.String("cluster-provider") == azure_provider.AzureClusterProvider:
			if err := os.RemoveAll(util.GetHomePath(azure_provider.AzureHomeDir,
				azure_provider.AzurePrefix+conf.Name+".json")); err != nil {
				return err
			}
		case c.String("cluster-provider") == google_provider.GoogleClusterProvider:
			if err := os.RemoveAll(util.GetHomePath(google_provider.GoogleHomeDir,
				google_provider.GoogleCredentialsPrefix+conf.Name+".json")); err != nil {
				return err
			}
		case c.String("cluster-provider") == onprem_provider.OnPremClusterProvider:
			if util.IsExists(conf.OnPremConfigure.SSHPrivateKey, true) {
				if err := os.RemoveAll(conf.OnPremConfigure.SSHPrivateKey); err != nil {
					return err
				}
			}
		}

		if err := os.RemoveAll(c.String("config")); err != nil {
			return err
		}

		zap.S().Infof("deleted config file by path: %s", c.String("config"))

		return nil
	}
}

func configInitAction(conf *config.Config, gitSpec *git_handler.GitSpec) cli.ActionFunc {
	return func(c *cli.Context) error {
		zap.S().Infof("loaded config file by path: %s", c.String("config"))

		start := time.Now()

		conf.Name = gitSpec.ID
		conf.Tenant = gitSpec.RepoPrefixName
		conf.Environment = gitSpec.DefaultBranch
		conf.ClusterProvider = c.String("cluster-provider")
		conf.ProgressBar = c.Bool("progress-bar")
		conf.GitHubToken = c.String("github-token")
		conf.SopsAgeKeys = util.GetHomePath(util.RMKDir, util.SopsRootName, conf.Tenant)
		zap.S().Infof("RMK will use values for %s environment", conf.Environment)

		if c.Bool("slack-notifications") {
			conf.SlackNotifications = c.Bool("slack-notifications")
			if !c.IsSet("slack-webhook") || !c.IsSet("slack-channel") {
				return fmt.Errorf("parameters --slack-webhook, --slack-channel not set, " +
					"required if Slack notifications are enabled")
			} else {
				conf.SlackWebHook = c.String("slack-webhook")
				conf.SlackChannel = c.String("slack-channel")
				conf.SlackMsgDetails = c.StringSlice("slack-message-details")
			}
		}

		switch conf.ClusterProvider {
		case aws_provider.AWSClusterProvider:
			conf.AzureConfigure = nil
			conf.GCPConfigure = nil
			conf.OnPremConfigure = nil
			if err := initAWSProfile(c, conf, gitSpec); err != nil {
				return err
			}
		case azure_provider.AzureClusterProvider:
			conf.AwsConfigure = nil
			conf.GCPConfigure = nil
			conf.OnPremConfigure = nil
			if err := initAzureProfile(c, conf, gitSpec); err != nil {
				return err
			}
		case google_provider.GoogleClusterProvider:
			conf.AwsConfigure = nil
			conf.AzureConfigure = nil
			conf.OnPremConfigure = nil
			if err := initGCPProfile(c, conf, gitSpec); err != nil {
				return err
			}
		case util.LocalClusterProvider:
			conf.AwsConfigure = nil
			conf.AzureConfigure = nil
			conf.GCPConfigure = nil
			conf.OnPremConfigure = nil
		case onprem_provider.OnPremClusterProvider:
			conf.AwsConfigure = nil
			conf.AzureConfigure = nil
			conf.GCPConfigure = nil
			if err := initOnPremProfile(c, conf, gitSpec); err != nil {
				return err
			}
		}

		if err := conf.InitConfig().SetRootDomain(gitSpec.ID); err != nil {
			return err
		}

		if err := conf.CreateConfigFile(); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		zap.S().Infof("time spent on initialization: %.fs", time.Since(start).Seconds())

		return nil
	}
}

func configListAction(conf *config.Config, gitSpec *git_handler.GitSpec) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := gitSpec.GenerateID(); err != nil {
			return err
		}

		conf.Tenant = gitSpec.RepoPrefixName
		return conf.GetConfigs(c.Bool("all"))
	}
}

func configViewAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if c.String("log-format") == "json" {
			serializeJsonConfig, err := conf.SerializeJsonConfig()
			if err != nil {
				return err
			}

			zap.L().Info("RMK", zap.Any("config", json.RawMessage(serializeJsonConfig)))
			return nil
		}

		serializeConfig, err := conf.SerializeConfig()
		if err != nil {
			return err
		}

		zap.S().Infof("loaded config file by path: %s", c.String("config"))
		fmt.Printf("%s\n", string(serializeConfig))

		return nil
	}
}
