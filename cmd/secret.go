package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"rmk/config"
	"rmk/providers/aws_provider"
	"rmk/providers/azure_provider"
	"rmk/providers/google_provider"
	"rmk/util"
)

type SecretRunner interface {
	helmSecretsEncrypt() error
	helmSecretsDecrypt() error
	helmSecretsView() error
	helmSecretsEdit() error
}

type SecretCommands struct {
	*ReleaseCommands
}

type SopsConfigFile struct {
	CreationRules []CreationRule `yaml:"creation_rules"`
}

type CreationRule struct {
	PathRegex string `yaml:"path_regex"`
	Age       string `yaml:"age"`
}

func newSecretCommands(conf *config.Config, ctx *cli.Context, workDir string) *SecretCommands {
	return &SecretCommands{&ReleaseCommands{Conf: conf, Ctx: ctx, WorkDir: workDir}}
}

func (sc *SecretCommands) ageKeygen(args ...string) *util.SpecCMD {
	return &util.SpecCMD{
		Args:    args,
		Command: "age-keygen",
		Ctx:     sc.Ctx.Context,
		Dir:     sc.WorkDir,
		Debug:   false,
	}
}

func (sc *SecretCommands) helm() *util.SpecCMD {
	return &util.SpecCMD{
		Args:          []string{"secrets"},
		Command:       "helm",
		Ctx:           sc.Ctx.Context,
		Dir:           sc.WorkDir,
		Envs:          []string{"SOPS_AGE_KEY_FILE=" + filepath.Join(sc.Conf.SopsAgeKeys, util.SopsAgeKeyFile)},
		Debug:         true,
		DisableStdOut: true,
	}
}

func (sc *SecretCommands) createAgeKey(scope string) error {
	keyPath := filepath.Join(sc.Conf.SopsAgeKeys, sc.Conf.Tenant+"-"+scope+".txt")

	if util.IsExists(keyPath, true) {
		return fmt.Errorf("key for scope %s exists, if you want to recreate, delete this file %s "+
			"and run the command again", scope, keyPath)
	}

	sc.SpecCMD = sc.ageKeygen("-o", keyPath)
	sc.SpecCMD.DisableStdOut = true
	if err := releaseRunner(sc).runCMD(); err != nil {
		return err
	}

	sc.SpecCMD = sc.ageKeygen("-y", keyPath)
	sc.SpecCMD.DisableStdOut = true
	return releaseRunner(sc).runCMD()
}

func (sc *SecretCommands) CreateKeys() error {
	if !util.IsExists(util.GetPwdPath(util.TenantValuesDIR), false) {
		return fmt.Errorf("'%s' directory not exist in project structure, please generate structure "+
			"by running command: 'rmk project generate'", util.TenantValuesDIR)
	}

	scopes, err := os.ReadDir(util.GetPwdPath(util.TenantValuesDIR))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(sc.Conf.SopsAgeKeys, 0775); err != nil {
		return err
	}

	for _, scope := range scopes {
		if scope.IsDir() && !strings.Contains(scope.Name(), "cluster") {
			if err := sc.createAgeKey(scope.Name()); err != nil {
				return err
			}

			zap.S().Infof("generate age key for scope: %s", scope.Name())

			sopsConfigFiles, err := util.WalkInDir(util.GetPwdPath(util.TenantValuesDIR, scope.Name()),
				"secrets", util.SopsConfigFile)
			if err != nil {
				return err
			}

			if len(sopsConfigFiles) == 0 {
				secretSpecFiles, err := util.WalkInDir(util.GetPwdPath(util.TenantValuesDIR, scope.Name()),
					"secrets", util.SecretSpecFile)
				if err != nil {
					return err
				}

				for _, specFile := range secretSpecFiles {
					dirSpecFile, _ := filepath.Split(specFile)
					sopsConfigFiles = append(sopsConfigFiles, filepath.Join(dirSpecFile, util.SopsConfigFile))
				}
			}

			for _, configFile := range sopsConfigFiles {
				sops := &SopsConfigFile{}
				sops.CreationRules = append(sops.CreationRules,
					CreationRule{
						PathRegex: ".+\\.yaml$",
						Age:       strings.ReplaceAll(sc.SpecCMD.StdoutBuf.String(), "\n", ""),
					})

				var data bytes.Buffer
				encoder := yaml.NewEncoder(&data)
				encoder.SetIndent(2)
				if err := encoder.Encode(&sops); err != nil {
					return err
				}

				if err := os.WriteFile(configFile, data.Bytes(), 0644); err != nil {
					return err
				}

				zap.S().Infof("update SOPS config file: %s", configFile)
			}
		}
	}

	return nil
}

func (sc *SecretCommands) DownloadKeys() error {
	switch sc.Conf.ClusterProvider {
	case aws_provider.AWSClusterProvider:
		secrets, err := aws_provider.NewAwsConfigure(sc.Ctx.Context, sc.Conf.Profile).GetAWSSecrets(sc.Conf.Tenant)
		if err != nil {
			return err
		}

		if len(secrets) == 0 {
			zap.S().Warnf("SOPS Age keys contents for tenant %s not found in %s secrets",
				sc.Conf.Tenant, strings.ToUpper(aws_provider.AWSClusterProvider))
		}

		for key, val := range secrets {
			zap.S().Infof("download AWS secret %s to %s",
				key, filepath.Join(sc.Conf.SopsAgeKeys, key+util.SopsAgeKeyExt))
			if err := os.WriteFile(filepath.Join(sc.Conf.SopsAgeKeys, key+util.SopsAgeKeyExt), val, 0644); err != nil {
				return err
			}
		}

		return nil
	case azure_provider.AzureClusterProvider:
		if err := sc.Conf.NewAzureClient(sc.Ctx.Context, sc.Conf.Name); err != nil {
			return err
		}

		secrets, err := sc.Conf.GetAzureSecrets()
		if err != nil {
			return err
		}

		if len(secrets) == 0 {
			zap.S().Warnf("SOPS Age keys contents for tenant %s not found in %s secrets",
				sc.Conf.Tenant, strings.ToUpper(aws_provider.AWSClusterProvider))
		}

		for key, val := range secrets {
			zap.S().Infof("download Azure key vault secret %s to %s",
				key, filepath.Join(sc.Conf.SopsAgeKeys, key+util.SopsAgeKeyExt))
			if err := os.WriteFile(filepath.Join(sc.Conf.SopsAgeKeys, key+util.SopsAgeKeyExt), val, 0644); err != nil {
				return err
			}
		}

		return nil
	case google_provider.GoogleClusterProvider:
		gcp := google_provider.NewGCPConfigure(sc.Ctx.Context, sc.Conf.GCPConfigure.AppCredentialsPath)

		secrets, err := gcp.GetGCPSecrets(sc.Conf.Tenant)
		if err != nil {
			return err
		}

		if len(secrets) == 0 {
			zap.S().Warnf("SOPS Age keys contents for tenant %s not found in %s secrets",
				sc.Conf.Tenant, strings.ToUpper(aws_provider.AWSClusterProvider))
		}

		for key, val := range secrets {
			zap.S().Infof("download GCP secret %s to %s",
				key, filepath.Join(sc.Conf.SopsAgeKeys, key+util.SopsAgeKeyExt))
			if err := os.WriteFile(filepath.Join(sc.Conf.SopsAgeKeys, key+util.SopsAgeKeyExt), val, 0644); err != nil {
				return err
			}
		}

		return nil
	default:
		return nil
	}
}

func (sc *SecretCommands) UploadKeys() error {
	switch sc.Conf.ClusterProvider {
	case aws_provider.AWSClusterProvider:
		a := aws_provider.NewAwsConfigure(sc.Ctx.Context, sc.Conf.Profile)

		walkMatch, err := util.WalkMatch(sc.Conf.SopsAgeKeys, sc.Conf.Tenant+"*"+util.SopsAgeKeyExt)
		if err != nil {
			return err
		}

		for _, val := range walkMatch {
			file, err := os.ReadFile(val)
			if err != nil {
				return err
			}

			keyName := strings.TrimSuffix(filepath.Base(val), util.SopsAgeKeyExt)

			if err := a.SetAWSSecret(sc.Conf.Tenant, keyName, file); err != nil {
				return err
			}
		}

		return nil
	case azure_provider.AzureClusterProvider:
		if err := sc.Conf.NewAzureClient(sc.Ctx.Context, sc.Conf.Name); err != nil {
			return err
		}

		walkMatch, err := util.WalkMatch(sc.Conf.SopsAgeKeys, sc.Conf.Tenant+"*"+util.SopsAgeKeyExt)
		if err != nil {
			return err
		}

		for _, val := range walkMatch {
			file, err := os.ReadFile(val)
			if err != nil {
				return err
			}

			keyName := strings.TrimSuffix(filepath.Base(val), util.SopsAgeKeyExt)
			value := string(file)

			if err := sc.Conf.SetAzureSecret(keyName, value); err != nil {
				return err
			}
		}

		return nil
	case google_provider.GoogleClusterProvider:
		gcp := google_provider.NewGCPConfigure(sc.Ctx.Context, sc.Conf.GCPConfigure.AppCredentialsPath)

		walkMatch, err := util.WalkMatch(sc.Conf.SopsAgeKeys, sc.Conf.Tenant+"*"+util.SopsAgeKeyExt)
		if err != nil {
			return err
		}

		for _, val := range walkMatch {
			file, err := os.ReadFile(val)
			if err != nil {
				return err
			}

			keyName := strings.TrimSuffix(filepath.Base(val), util.SopsAgeKeyExt)

			if err := gcp.SetGCPSecret(sc.Conf.Tenant, sc.Conf.GCPRegion, keyName, file); err != nil {
				return err
			}
		}

		return nil
	default:
		return nil
	}
}

func (sc *SecretCommands) getOptionFiles(option string) ([]string, error) {
	var optionFiles, optionPaths []string
	var check int

	switch {
	case !sc.Ctx.IsSet("scope") && !sc.Ctx.IsSet("environment"):
		var err error
		optionFiles, err = util.WalkInDir(util.GetPwdPath(util.TenantValuesDIR, filepath.Join(optionPaths...)),
			"secrets", option)
		if err != nil {
			return nil, err
		}
	case sc.Ctx.IsSet("scope") && !sc.Ctx.IsSet("environment"):
		for _, scope := range sc.Ctx.StringSlice("scope") {
			sopsFiles, err := util.WalkInDir(util.GetPwdPath(util.TenantValuesDIR, scope),
				"secrets", option)
			if err != nil {
				return nil, err
			}

			optionFiles = append(optionFiles, sopsFiles...)
		}
	case !sc.Ctx.IsSet("scope") && sc.Ctx.IsSet("environment"):
		for _, environment := range sc.Ctx.StringSlice("environment") {
			for env := range sc.Conf.Project.Spec.Environments {
				if environment == env {
					check++
				}
			}

			if check == 0 {
				return nil, fmt.Errorf("environment %s do not exist in project.spec.environments", environment)
			}

			sopsFiles, err := util.WalkInDir(util.GetPwdPath(util.TenantValuesDIR),
				environment, filepath.Join("secrets", option))
			if err != nil {
				return nil, err
			}

			optionFiles = append(optionFiles, sopsFiles...)
		}
	case sc.Ctx.IsSet("scope") && sc.Ctx.IsSet("environment"):
		for _, scope := range sc.Ctx.StringSlice("scope") {
			for _, environment := range sc.Ctx.StringSlice("environment") {
				optionPaths = append(optionPaths, scope, environment)
				sopsFiles, err := util.WalkInDir(util.GetPwdPath(util.TenantValuesDIR, filepath.Join(optionPaths...)),
					"secrets", option)
				if err != nil {
					return nil, err
				}

				optionPaths = []string{}
				optionFiles = append(optionFiles, sopsFiles...)
			}
		}
	}

	return optionFiles, nil
}

func (sc *SecretCommands) getSecretPaths(optionFiles []string) ([]string, error) {
	var tempPaths, secretPaths []string

	for _, configFile := range optionFiles {
		path := strings.ReplaceAll(configFile, string(filepath.Separator)+filepath.Base(configFile), "")
		tempPaths = append(tempPaths, path)
	}

	for _, tempPath := range tempPaths {
		secrets, err := util.WalkMatch(tempPath, "*.yaml")
		if err != nil {
			return nil, err
		}

		for _, secretPath := range secrets {
			_, file := filepath.Split(secretPath)
			if file != util.SopsConfigFile && file != util.SecretSpecFile {
				secretPaths = append(secretPaths, secretPath)
			}
		}
	}

	return secretPaths, nil
}

func (sc *SecretCommands) SecretManager(option string) error {
	optionFiles, err := sc.getOptionFiles(option)
	if err != nil {
		return err
	}

	if sc.Ctx.Command.Name == "generate" {
		if err := sc.genSpecSecrets(optionFiles); err != nil {
			return err
		}

		return nil
	}

	secretPaths, err := sc.getSecretPaths(optionFiles)
	if err != nil {
		return err
	}

	if err := util.MergeAgeKeys(sc.Conf.SopsAgeKeys); err != nil {
		return err
	}

	for _, secret := range secretPaths {
		sc.SpecCMD = sc.helm()

		switch sc.Ctx.Command.Name {
		case "decrypt":
			sc.SpecCMD.Args = append(sc.SpecCMD.Args, sc.Ctx.Command.Name, "-i", secret)
			if err := releaseRunner(sc).runCMD(); err != nil {
				if strings.Contains(sc.SpecCMD.StderrBuf.String(), util.HelmSecretsIsNotEncrypted+secret) {
					zap.S().Warnf(strings.ToLower(util.HelmSecretsIsNotEncrypted)+"%s", secret)
					continue
				} else {
					return fmt.Errorf(sc.SpecCMD.StderrBuf.String())
				}
			}

			zap.S().Infof("decrypting: %s", secret)
		case "encrypt":
			sc.SpecCMD.Args = append(sc.SpecCMD.Args, sc.Ctx.Command.Name, "-i", secret)
			if err := releaseRunner(sc).runCMD(); err != nil {
				if strings.Contains(sc.SpecCMD.StderrBuf.String(), util.HelmSecretsAlreadyEncrypted+filepath.Base(secret)) {
					zap.S().Warnf(strings.ToLower(util.HelmSecretsAlreadyEncrypted)+"%s", secret)
					continue
				} else {
					return fmt.Errorf(sc.SpecCMD.StderrBuf.String())
				}
			}

			zap.S().Infof("encrypting: %s", secret)
		}
	}

	return nil
}

func (sc *SecretCommands) helmSecretsEncrypt() error {
	if err := util.MergeAgeKeys(sc.Conf.SopsAgeKeys); err != nil {
		return err
	}

	sc.SpecCMD = sc.helm()
	sc.SpecCMD.Args = append(sc.SpecCMD.Args, sc.Ctx.Command.Name, "-i", sc.Ctx.Args().First())

	return sc.runHelmSecretsCMD(sc.Ctx.Args().First(), false)
}

func (sc *SecretCommands) helmSecretsDecrypt() error {
	if err := util.MergeAgeKeys(sc.Conf.SopsAgeKeys); err != nil {
		return err
	}

	sc.SpecCMD = sc.helm()
	sc.SpecCMD.Args = append(sc.SpecCMD.Args, sc.Ctx.Command.Name, "-i", sc.Ctx.Args().First())

	return sc.runHelmSecretsCMD(sc.Ctx.Args().First(), false)
}

func (sc *SecretCommands) helmSecretsView() error {
	if err := util.MergeAgeKeys(sc.Conf.SopsAgeKeys); err != nil {
		return err
	}

	sc.SpecCMD = sc.helm()
	sc.SpecCMD.Args = append(sc.SpecCMD.Args, "decrypt", sc.Ctx.Args().First())

	if err := sc.runHelmSecretsCMD(sc.Ctx.Args().First(), false); err != nil {
		return err
	}

	fmt.Println(sc.SpecCMD.StdoutBuf.String())

	return nil
}

func (sc *SecretCommands) helmSecretsEdit() error {
	if err := util.MergeAgeKeys(sc.Conf.SopsAgeKeys); err != nil {
		return err
	}

	sc.SpecCMD = sc.helm()
	sc.SpecCMD.Args = append(sc.SpecCMD.Args, sc.Ctx.Command.Name, sc.Ctx.Args().First())
	sc.SpecCMD.DisableStdOut = false

	return sc.runHelmSecretsCMD(sc.Ctx.Args().First(), true)
}

func (sc *SecretCommands) runHelmSecretsCMD(secretFilePath string, returnCMDError bool) error {
	if !util.IsExists(secretFilePath, true) {
		return fmt.Errorf("file does not exist: %s", secretFilePath)
	}

	if err := releaseRunner(sc).runCMD(); err != nil {
		if returnCMDError {
			return err
		}

		out := sc.SpecCMD.StderrBuf.String()

		// suppress help message of helm secrets
		if strings.Contains(out, util.HelpFlagFull) {
			return fmt.Errorf(util.UnknownErrorText, "Helm secrets")
		}

		// remove unneeded text from helm secrets
		out = strings.ReplaceAll(out, util.HelmSecretsOutputPrefix, "")
		out = strings.ReplaceAll(out, util.HelmSecretsError, "")
		out = strings.TrimSpace(out)

		// make the first letter lowercase
		outRunes := []rune(out)
		outRunes[0] = unicode.ToLower(outRunes[0])
		out = string(outRunes)

		return fmt.Errorf(out)
	}

	return nil
}

func secretMgrEncryptDecryptAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		return newSecretCommands(conf, c, util.GetPwdPath("")).SecretManager(util.SopsConfigFile)
	}
}

func secretMgrGenerateAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		return newSecretCommands(conf, c, util.GetPwdPath("")).SecretManager(util.SecretSpecFile)
	}
}
func secretKeysCreateAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		return newSecretCommands(conf, c, util.GetPwdPath("")).CreateKeys()
	}
}

func secretKeysDownloadAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		return newSecretCommands(conf, c, util.GetPwdPath("")).DownloadKeys()
	}
}

func secretKeysUploadAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		return newSecretCommands(conf, c, util.GetPwdPath("")).UploadKeys()
	}
}

func secretAction(conf *config.Config, action func(secretRunner SecretRunner) error) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 1); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		return action(newSecretCommands(conf, c, util.GetPwdPath("")))
	}
}
