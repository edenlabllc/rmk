package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"rmk/config"
	"rmk/git_handler"
	"rmk/system"
)

type ConfigCommands struct {
	*ReleaseCommands
}

func newConfigCommands(conf *config.Config, ctx *cli.Context, workDir string) *ConfigCommands {
	return &ConfigCommands{&ReleaseCommands{Conf: conf, Ctx: ctx, WorkDir: workDir}}
}

func (c *ConfigCommands) awsConfigure(profile string) *system.SpecCMD {
	return &system.SpecCMD{
		Args: []string{"configure", "--profile", profile},
		Envs: []string{
			"AWS_CONFIG_FILE=" + strings.Join(c.Conf.AWSSharedConfigFile(profile), ""),
			"AWS_SHARED_CREDENTIALS_FILE=" + strings.Join(c.Conf.AWSSharedCredentialsFile(profile), ""),
		},
		Command: "aws",
		Ctx:     c.Ctx.Context,
		Dir:     c.WorkDir,
		Debug:   false,
	}
}

func (c *ConfigCommands) helmPlugin() *system.SpecCMD {
	return &system.SpecCMD{
		Args:          []string{"plugin"},
		Command:       "helm",
		Dir:           c.WorkDir,
		Ctx:           c.Ctx.Context,
		DisableStdOut: true,
		Debug:         false,
	}
}

func (c *ConfigCommands) rmkConfigInit() *system.SpecCMD {
	exRMK, err := os.Executable()
	if err != nil {
		panic(err)
	}

	return &system.SpecCMD{
		Args:    []string{"config", "init"},
		Command: exRMK,
		Dir:     c.WorkDir,
		Ctx:     c.Ctx.Context,
		Debug:   true,
	}
}

func (c *ConfigCommands) checkAwsEnv() (map[string]string, bool) {
	awsEnvs := map[string]string{
		"region":                "AWS_REGION",
		"aws_access_key_id":     "AWS_ACCESS_KEY_ID",
		"aws_secret_access_key": "AWS_SECRET_ACCESS_KEY",
		"aws_session_token":     "AWS_SESSION_TOKEN",
	}

	for key, val := range awsEnvs {
		value, ok := os.LookupEnv(val)
		if !ok {
			delete(awsEnvs, key)
		} else {
			awsEnvs[key] = value
		}
	}

	if len(awsEnvs) > 0 {
		return awsEnvs, true
	} else {
		return nil, false
	}
}

func (c *ConfigCommands) configAws() error {
	if awsEnvs, ok := c.checkAwsEnv(); !ok {
		c.SpecCMD = c.awsConfigure(c.Conf.Profile)
		return runner(c).runCMD()
	} else {
		for key, val := range awsEnvs {
			c.SpecCMD = c.awsConfigure(c.Conf.Profile)
			c.SpecCMD.Args = append(c.SpecCMD.Args, "set", key, val)
			if err := runner(c).runCMD(); err != nil {
				return err
			}
		}

		zap.S().Infof("AWS profile by name %s was created", c.Conf.Profile)
		return nil
	}
}

func (c *ConfigCommands) configAwsMFA() error {
	var tokenExpiration time.Time
	currentTime := time.Now()
	regularProfile := c.Conf.Profile

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

	if err := c.Conf.GetMFADevicesSerialNumbers(); err != nil {
		return err
	}

	timeDiff := time.Time{}.Add(tokenExpiration.Sub(currentTime)).Format("15:04:05")

	if len(c.Conf.MFADeviceSerialNumber) > 0 {
		zap.S().Infof("MFA device SerialNumber: %s", c.Conf.MFADeviceSerialNumber)
	}

	if currentTime.Before(tokenExpiration) {
		zap.S().Infof("MFA remaining time for token validity: %s", timeDiff)
	}

	if len(c.Conf.MFADeviceSerialNumber) > 0 && currentTime.After(tokenExpiration) {
		if err := c.Conf.GetMFASessionToken(); err != nil {
			return err
		}

		c.Conf.AWSMFAProfile = regularProfile + "-mfa"
		c.Conf.AWSMFATokenExpiration = strconv.FormatInt(c.Conf.Expiration.Unix(), 10)

		MFAProfileArgs := map[string]string{
			"aws_access_key_id":     c.Conf.MFAProfileCredentials.AccessKeyID,
			"aws_secret_access_key": c.Conf.MFAProfileCredentials.SecretAccessKey,
			"output":                "text",
			"region":                c.Conf.Region,
		}

		regularProfileArgs := map[string]string{
			"aws_access_key_id":     c.Conf.MFAToken.AccessKeyId,
			"aws_secret_access_key": c.Conf.MFAToken.SecretAccessKey,
			"aws_session_token":     c.Conf.MFAToken.SessionToken,
		}

		for key, val := range MFAProfileArgs {
			c.SpecCMD = c.awsConfigure(c.Conf.AWSMFAProfile)
			c.SpecCMD.Args = append(c.SpecCMD.Args, "set", key, val)
			if err := runner(c).runCMD(); err != nil {
				return err
			}
		}

		for key, val := range regularProfileArgs {
			c.SpecCMD = c.awsConfigure(regularProfile)
			c.SpecCMD.Args = append(c.SpecCMD.Args, "set", key, val)
			if err := runner(c).runCMD(); err != nil {
				return err
			}
		}
	}

	c.Conf.AwsConfigure.Profile = regularProfile

	return nil
}

func (c *ConfigCommands) copyAWSProfile(profile string) error {
	profileArgs := map[string]string{
		"aws_access_key_id":     c.Conf.MFAProfileCredentials.AccessKeyID,
		"aws_secret_access_key": c.Conf.MFAProfileCredentials.SecretAccessKey,
		"output":                "text",
		"region":                c.Conf.Region,
	}

	for key, val := range profileArgs {
		c.SpecCMD = c.awsConfigure(profile)
		c.SpecCMD.Args = append(c.SpecCMD.Args, "set", key, val)
		if err := runner(c).runCMD(); err != nil {
			return err
		}
	}

	return nil
}

func (c *ConfigCommands) uninstallHelmPlugin(plugin config.Package) error {
	c.SpecCMD = c.helmPlugin()
	c.SpecCMD.Args = append(c.SpecCMD.Args, "list")
	plSemVer, _ := semver.NewVersion(plugin.Version)

	if err := runner(c).runCMD(); err != nil {
		return fmt.Errorf("get Helm plugin list failed: %s", c.SpecCMD.StderrBuf.String())
	}

	for _, v := range strings.Split(c.SpecCMD.StdoutBuf.String(), "\n") {
		if strings.Contains(v, plugin.Name) && !strings.Contains(v, plSemVer.String()) {
			zap.S().Infof("Helm plugin %s detect new version %s from config", plugin.Name, plugin.Version)
			c.SpecCMD = c.helmPlugin()
			c.SpecCMD.Args = append(c.SpecCMD.Args, "uninstall", plugin.Name)
			if err := runner(c).runCMD(); err != nil {
				return fmt.Errorf("Helm plugin %s uninstallation failed: \n%s",
					plugin.Name, c.SpecCMD.StderrBuf.String())
			}

			break
		}
	}

	return nil
}

func (c *ConfigCommands) installHelmPlugin(plugin config.Package, args ...string) error {
	c.SpecCMD = c.helmPlugin()
	c.SpecCMD.Args = append(c.SpecCMD.Args, args...)
	if err := runner(c).runCMD(); err != nil {
		if !strings.Contains(c.SpecCMD.StderrBuf.String(), system.HelmPluginExist) {
			return fmt.Errorf("Helm plugin %s installation failed: \n%s", plugin.Name, c.SpecCMD.StderrBuf.String())
		}
	}

	if !strings.Contains(c.SpecCMD.StderrBuf.String(), system.HelmPluginExist) {
		zap.S().Infof("installing Helm plugin: %s", plugin.Name)
	}

	return nil
}

func (c *ConfigCommands) configHelmPlugins() error {
	for _, plugin := range c.Conf.HelmPlugins {
		if err := c.uninstallHelmPlugin(*plugin); err != nil {
			return err
		}

		if err := c.installHelmPlugin(*plugin, "install", plugin.Url, "--version="+plugin.Version); err != nil {
			return err
		}
	}

	return nil
}

func (c *ConfigCommands) rmkConfig() error {
	c.SpecCMD = c.rmkConfigInit()
	return runner(c).runCMD()
}

func initAWSProfile(c *cli.Context, conf *config.Config, gitSpec *git_handler.GitSpec) error {
	var profile string

	// Detect if MFA is enabled
	if len(conf.AWSMFAProfile) > 0 && len(conf.AWSMFATokenExpiration) > 0 {
		profile = conf.AWSMFAProfile
	} else {
		profile = conf.Profile
	}

	if c.Bool("aws-reconfigure") {
		if err := os.RemoveAll(strings.Join(conf.AWSSharedCredentialsFile(conf.Profile), "")); err != nil {
			return err
		}

		// Reconfigure regular AWS profile
		if err := newConfigCommands(conf, c, system.GetPwdPath("")).configAws(); err != nil {
			return err
		}

		// Get CallerIdentity and region for regular AWS profile
		if _, err := conf.AwsConfigure.GetAwsConfigure(conf.Profile); err != nil {
			return err
		}

		// Delete MFA profile
		if strings.Contains(profile, "-mfa") {
			if err := os.RemoveAll(strings.Join(conf.AWSSharedConfigFile(profile), "")); err != nil {
				return err
			}

			if err := os.RemoveAll(strings.Join(conf.AWSSharedCredentialsFile(profile), "")); err != nil {
				return err
			}
		}

		// Reset ConfigFrom value for config for current environment
		conf.ConfigFrom = gitSpec.ID
		// Reset AWSMFAProfile value for config for current environment
		conf.AWSMFAProfile = ""
		// Reset AWSMFATokenExpiration value for config for current environment
		conf.AWSMFATokenExpiration = ""
		// Returning a regular profile value
		profile = conf.Profile

		// Create new MFA profile
		if err := newConfigCommands(conf, c, system.GetPwdPath("")).configAwsMFA(); err != nil {
			return err
		}
	}

	if ok, err := conf.AwsConfigure.GetAwsConfigure(profile); err != nil && ok {
		zap.S().Warnf("%s", err.Error())
		if err := newConfigCommands(conf, c, system.GetPwdPath("")).configAws(); err != nil {
			return err
		}

		if _, err := conf.AwsConfigure.GetAwsConfigure(profile); err != nil {
			return err
		}

		if err := newConfigCommands(conf, c, system.GetPwdPath("")).configAwsMFA(); err != nil {
			return err
		}
	} else if !c.Bool("aws-reconfigure") {
		if err := newConfigCommands(conf, c, system.GetPwdPath("")).configAwsMFA(); err != nil {
			return err
		}
	} else if !ok && err != nil {
		return err
	}

	return nil
}

func getConfigFromEnvironment(c *cli.Context, conf *config.Config, gitSpec *git_handler.GitSpec) error {
	if len(c.String("config-from-environment")) > 0 {
		configPath := system.GetHomePath(system.RMKDir, system.RMKConfig,
			gitSpec.RepoPrefixName+"-"+c.String("config-from-environment")+".yaml")

		if err := conf.ReadConfigFile(configPath); err != nil {
			zap.S().Errorf("RMK config %s.yaml not initialized, please checkout to branch %s "+
				"and run command 'rmk config init' with specific parameters",
				c.String("config-from-environment"), c.String("config-from-environment"))
			return err
		}

		if err := c.Set("config-from", conf.Name); err != nil {
			return err
		}

		// Delete regular profile
		if err := os.RemoveAll(strings.Join(conf.AWSSharedCredentialsFile(gitSpec.ID), "")); err != nil {
			return err
		}

		if len(conf.AWSMFAProfile) > 0 && len(conf.AWSMFATokenExpiration) > 0 {
			regularProfile := conf.Profile

			// Get MFA profile credentials.
			conf.AwsConfigure.Profile = conf.AWSMFAProfile
			if err := conf.GetAWSCredentials(); err != nil {
				return err
			}

			// Copy MFA profile for current environment
			conf.AwsConfigure.Profile = regularProfile
			if err := newConfigCommands(conf, c, system.GetPwdPath("")).copyAWSProfile(gitSpec.ID); err != nil {
				return err
			}
		} else {
			// Delete config MFA profile
			if err := os.RemoveAll(strings.Join(conf.AWSSharedConfigFile(gitSpec.ID+"-mfa"), "")); err != nil {
				return err
			}

			// Delete credentials MFA profile
			if err := os.RemoveAll(strings.Join(conf.AWSSharedCredentialsFile(gitSpec.ID+"-mfa"), "")); err != nil {
				return err
			}

			// Get regular profile credentials.
			if err := conf.GetAWSCredentials(); err != nil {
				return err
			}

			// Copy regular profile for current environment
			if err := newConfigCommands(conf, c, system.GetPwdPath("")).copyAWSProfile(gitSpec.ID); err != nil {
				return err
			}
		}

		// Reset AWSMFAProfile value for config for current environment
		if err := c.Set("aws-mfa-profile", ""); err != nil {
			return err
		}

		// Reset AWSMFATokenExpiration value for config for current environment
		if err := c.Set("aws-mfa-token-expiration", ""); err != nil {
			return err
		}

		conf.ConfigFrom = c.String("config-from")
		conf.AwsConfigure.Profile = gitSpec.ID

		return nil
	}

	if err := system.ValidateArtifactModeDefault(c, "required parameter --github-token not set"); err != nil {
		return err
	}

	if err := system.ValidateNArg(c, 0); err != nil {
		return err
	}

	if !c.IsSet("config-from") {
		if err := c.Set("config-from", gitSpec.ID); err != nil {
			return err
		}
	}

	conf.ConfigFrom = c.String("config-from")
	conf.AwsConfigure.Profile = gitSpec.ID
	conf.CloudflareToken = c.String("cloudflare-token")
	conf.GitHubToken = c.String("github-token")

	return nil
}

func configDeleteAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		// Delete MFA profile
		if len(conf.AWSMFAProfile) > 0 && len(conf.AWSMFATokenExpiration) > 0 {
			if err := os.RemoveAll(strings.Join(conf.AWSSharedConfigFile(conf.AWSMFAProfile), "")); err != nil {
				return err
			}

			if err := os.RemoveAll(strings.Join(conf.AWSSharedCredentialsFile(conf.AWSMFAProfile), "")); err != nil {
				return err
			}
		}

		// Delete config MFA profile
		if err := os.RemoveAll(strings.Join(conf.AWSSharedConfigFile(conf.Profile), "")); err != nil {
			return err
		}

		// Delete credentials MFA profile
		if err := os.RemoveAll(strings.Join(conf.AWSSharedCredentialsFile(conf.Profile), "")); err != nil {
			return err
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
		if err := getConfigFromEnvironment(c, conf, gitSpec); err != nil {
			return err
		}

		zap.S().Infof("loaded config file by path: %s", c.String("config"))

		start := time.Now()

		conf.Name = gitSpec.ID
		conf.Tenant = gitSpec.RepoPrefixName
		conf.Environment = gitSpec.DefaultBranch
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

		conf.ArtifactMode = c.String("artifact-mode")
		conf.ProgressBar = c.Bool("progress-bar")
		conf.Terraform.BucketKey = system.TenantBucketKey
		conf.ClusterProvisionerSL = c.Bool("cluster-provisioner-state-locking")
		conf.S3ChartsRepoRegion = c.String("s3-charts-repo-region")
		conf.ClusterProvider = c.String("cluster-provider")
		conf.AWSMFAProfile = c.String("aws-mfa-profile")
		conf.AWSMFATokenExpiration = c.String("aws-mfa-token-expiration")
		conf.AWSECRHost = c.String("aws-ecr-host")
		conf.AWSECRRegion = c.String("aws-ecr-region")
		conf.AWSECRUserName = c.String("aws-ecr-user-name")

		// AWS Profile init configuration with support MFA
		if err := initAWSProfile(c, conf, gitSpec); err != nil {
			return err
		}

		//Formation of a unique bucket name, consisting of the prefix tenant of the repository,
		//constant and the first 3 and last 2 numbers AWS account id
		awsUID := conf.AccountID[0:3] + conf.AccountID[len(conf.AccountID)-2:]
		conf.SopsAgeKeys = system.GetHomePath(system.RMKDir, system.SopsRootName, conf.Tenant+"-"+system.SopsRootName+"-"+awsUID)
		conf.SopsBucketName = conf.Tenant + "-" + system.SopsRootName + "-" + awsUID
		conf.Terraform.BucketName = conf.Tenant + "-" + system.TenantBucketName + "-" + awsUID
		conf.Terraform.DDBTableName = system.TenantDDBTablePrefix + "-" + awsUID

		if err := conf.InitConfig(true).SetRootDomain(c, gitSpec.ID); err != nil {
			return err
		}

		if err := conf.CreateConfigFile(); err != nil {
			return err
		}

		if conf.ClusterProvider == system.AWSClusterProvider {
			if conf.ClusterProvisionerSL {
				// create dynamodb table for backend terraform
				if err := conf.CreateDynamoDBTable(conf.Terraform.DDBTableName); err != nil {
					return err
				}
			}

			// create s3 bucket for backend terraform
			if err := conf.CreateBucket(conf.Terraform.BucketName); err != nil {
				return err
			}

			//create s3 bucket for sops age keys
			if err := conf.CreateBucket(conf.SopsBucketName); err != nil {
				return err
			}

			if err := conf.DownloadFromBucket("", conf.SopsBucketName, conf.SopsAgeKeys, conf.Tenant); err != nil {
				return err
			}
		}

		zap.S().Infof("time spent on initialization: %.fs", time.Since(start).Seconds())

		return nil
	}
}

func configListAction(conf *config.Config, gitSpec *git_handler.GitSpec) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateNArg(c, 0); err != nil {
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
		if err := system.ValidateNArg(c, 0); err != nil {
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
