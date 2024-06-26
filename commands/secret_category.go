package commands

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
	"rmk/system"
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

func (sc *SecretCommands) ageKeygen(args ...string) *system.SpecCMD {
	return &system.SpecCMD{
		Args:    args,
		Command: "age-keygen",
		Ctx:     sc.Ctx.Context,
		Dir:     sc.WorkDir,
		Debug:   false,
	}
}

func (sc *SecretCommands) helm() *system.SpecCMD {
	return &system.SpecCMD{
		Args:          []string{"secrets"},
		Command:       "helm",
		Ctx:           sc.Ctx.Context,
		Dir:           sc.WorkDir,
		Envs:          []string{"SOPS_AGE_KEY_FILE=" + filepath.Join(sc.Conf.SopsAgeKeys, system.SopsAgeKeyFile)},
		Debug:         true,
		DisableStdOut: true,
	}
}

func (sc *SecretCommands) createAgeKey(scope string) error {
	keyPath := filepath.Join(sc.Conf.SopsAgeKeys, sc.Conf.Tenant+"-"+scope+".txt")

	if system.IsExists(keyPath, true) {
		return fmt.Errorf("key for scope %s exists, if you want to recreate, delete this file %s "+
			"and run the command again", scope, keyPath)
	}

	sc.SpecCMD = sc.ageKeygen("-o", keyPath)
	sc.SpecCMD.DisableStdOut = true
	if err := runner(sc).runCMD(); err != nil {
		return err
	}

	sc.SpecCMD = sc.ageKeygen("-y", keyPath)
	sc.SpecCMD.DisableStdOut = true
	return runner(sc).runCMD()
}

func (sc *SecretCommands) CreateKeys() error {
	if !system.IsExists(system.GetPwdPath(system.TenantValuesDIR), false) {
		return fmt.Errorf("'%s' directory not exist in project structure, please generate structure "+
			"by running command: 'rmk project generate'", system.TenantValuesDIR)
	}

	scopes, err := os.ReadDir(system.GetPwdPath(system.TenantValuesDIR))
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

			sopsConfigFiles, err := system.WalkInDir(system.GetPwdPath(system.TenantValuesDIR, scope.Name()),
				"secrets", system.SopsConfigFile)
			if err != nil {
				return err
			}

			if len(sopsConfigFiles) == 0 {
				secretSpecFiles, err := system.WalkInDir(system.GetPwdPath(system.TenantValuesDIR, scope.Name()),
					"secrets", system.SecretSpecFile)
				if err != nil {
					return err
				}

				for _, specFile := range secretSpecFiles {
					dirSpecFile, _ := filepath.Split(specFile)
					sopsConfigFiles = append(sopsConfigFiles, filepath.Join(dirSpecFile, system.SopsConfigFile))
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

func (sc *SecretCommands) getOptionFiles(option string) ([]string, error) {
	var optionFiles, optionPaths []string
	var check int

	switch {
	case !sc.Ctx.IsSet("scope") && !sc.Ctx.IsSet("environment"):
		var err error
		optionFiles, err = system.WalkInDir(system.GetPwdPath(system.TenantValuesDIR, filepath.Join(optionPaths...)),
			"secrets", option)
		if err != nil {
			return nil, err
		}
	case sc.Ctx.IsSet("scope") && !sc.Ctx.IsSet("environment"):
		for _, scope := range sc.Ctx.StringSlice("scope") {
			sopsFiles, err := system.WalkInDir(system.GetPwdPath(system.TenantValuesDIR, scope),
				"secrets", option)
			if err != nil {
				return nil, err
			}

			optionFiles = append(optionFiles, sopsFiles...)
		}
	case !sc.Ctx.IsSet("scope") && sc.Ctx.IsSet("environment"):
		for _, environment := range sc.Ctx.StringSlice("environment") {
			for _, env := range sc.Conf.Project.Spec.Environments {
				if environment == env {
					check++
				}
			}

			if check == 0 {
				return nil, fmt.Errorf("environment %s do not exist in project.spec.environments", environment)
			}

			sopsFiles, err := system.WalkInDir(system.GetPwdPath(system.TenantValuesDIR),
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
				sopsFiles, err := system.WalkInDir(system.GetPwdPath(system.TenantValuesDIR, filepath.Join(optionPaths...)),
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
		secrets, err := system.WalkMatch(tempPath, "*.yaml")
		if err != nil {
			return nil, err
		}

		for _, secretPath := range secrets {
			_, file := filepath.Split(secretPath)
			if file != system.SopsConfigFile && file != system.SecretSpecFile {
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

	if err := system.MergeAgeKeys(sc.Conf.SopsAgeKeys); err != nil {
		return err
	}

	for _, secret := range secretPaths {
		sc.SpecCMD = sc.helm()

		switch sc.Ctx.Command.Name {
		case "decrypt":
			sc.SpecCMD.Args = append(sc.SpecCMD.Args, sc.Ctx.Command.Name, "-i", secret)
			if err := runner(sc).runCMD(); err != nil {
				if strings.Contains(sc.SpecCMD.StderrBuf.String(), system.HelmSecretsIsNotEncrypted+secret) {
					zap.S().Warnf(strings.ToLower(system.HelmSecretsIsNotEncrypted)+"%s", secret)
					continue
				} else {
					return fmt.Errorf(sc.SpecCMD.StderrBuf.String())
				}
			}

			zap.S().Infof("decrypting: %s", secret)
		case "encrypt":
			sc.SpecCMD.Args = append(sc.SpecCMD.Args, sc.Ctx.Command.Name, "-i", secret)
			if err := runner(sc).runCMD(); err != nil {
				if strings.Contains(sc.SpecCMD.StderrBuf.String(), system.HelmSecretsAlreadyEncrypted+filepath.Base(secret)) {
					zap.S().Warnf(strings.ToLower(system.HelmSecretsAlreadyEncrypted)+"%s", secret)
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
	if err := system.MergeAgeKeys(sc.Conf.SopsAgeKeys); err != nil {
		return err
	}

	sc.SpecCMD = sc.helm()
	sc.SpecCMD.Args = append(sc.SpecCMD.Args, sc.Ctx.Command.Name, "-i", sc.Ctx.Args().First())

	return sc.runHelmSecretsCMD(sc.Ctx.Args().First(), false)
}

func (sc *SecretCommands) helmSecretsDecrypt() error {
	if err := system.MergeAgeKeys(sc.Conf.SopsAgeKeys); err != nil {
		return err
	}

	sc.SpecCMD = sc.helm()
	sc.SpecCMD.Args = append(sc.SpecCMD.Args, sc.Ctx.Command.Name, "-i", sc.Ctx.Args().First())

	return sc.runHelmSecretsCMD(sc.Ctx.Args().First(), false)
}

func (sc *SecretCommands) helmSecretsView() error {
	if err := system.MergeAgeKeys(sc.Conf.SopsAgeKeys); err != nil {
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
	if err := system.MergeAgeKeys(sc.Conf.SopsAgeKeys); err != nil {
		return err
	}

	sc.SpecCMD = sc.helm()
	sc.SpecCMD.Args = append(sc.SpecCMD.Args, sc.Ctx.Command.Name, sc.Ctx.Args().First())
	sc.SpecCMD.DisableStdOut = false

	return sc.runHelmSecretsCMD(sc.Ctx.Args().First(), true)
}

func (sc *SecretCommands) runHelmSecretsCMD(secretFilePath string, returnCMDError bool) error {
	if !system.IsExists(secretFilePath, true) {
		return fmt.Errorf("file does not exist: %s", secretFilePath)
	}

	if err := runner(sc).runCMD(); err != nil {
		if returnCMDError {
			return err
		}

		out := sc.SpecCMD.StderrBuf.String()

		// suppress help message of helm secrets
		if strings.Contains(out, system.HelpFlagFull) {
			return fmt.Errorf(system.UnknownErrorText, "Helm secrets")
		}

		// remove unneeded text from helm secrets
		out = strings.ReplaceAll(out, system.HelmSecretsOutputPrefix, "")
		out = strings.ReplaceAll(out, system.HelmSecretsError, "")
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
		if err := system.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(false), c, false); err != nil {
			return err
		}

		return newSecretCommands(conf, c, system.GetPwdPath("")).SecretManager(system.SopsConfigFile)
	}
}

func secretMgrGenerateAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		return newSecretCommands(conf, c, system.GetPwdPath("")).SecretManager(system.SecretSpecFile)
	}
}
func secretKeysCreateAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(false), c, false); err != nil {
			return err
		}

		return newSecretCommands(conf, c, system.GetPwdPath("")).CreateKeys()
	}
}

func secretKeysDownloadAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		return conf.DownloadFromBucket("", conf.SopsBucketName, conf.SopsAgeKeys, conf.Tenant)
	}
}

func secretKeysUploadAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		return conf.UploadToBucket(conf.SopsBucketName, conf.SopsAgeKeys, "*"+system.SopsAgeKeyExt)
	}
}

func secretAction(conf *config.Config, action func(secretRunner SecretRunner) error) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := system.ValidateNArg(c, 1); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(false), c, false); err != nil {
			return err
		}

		return action(newSecretCommands(conf, c, system.GetPwdPath("")))
	}
}
