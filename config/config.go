package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"rmk/git_handler"
	"rmk/providers/aws_provider"
	"rmk/providers/azure_provider"
	"rmk/providers/google_provider"
	"rmk/util"
)

type Config struct {
	Name                           string   `yaml:"name,omitempty"`
	Tenant                         string   `yaml:"tenant,omitempty"`
	Environment                    string   `yaml:"environment,omitempty"`
	RootDomain                     string   `yaml:"root-domain,omitempty"`
	GitHubToken                    string   `yaml:"github-token,omitempty"`
	ClusterProvider                string   `yaml:"cluster-provider"`
	SlackNotifications             bool     `yaml:"slack-notifications"`
	SlackWebHook                   string   `yaml:"slack-webhook,omitempty"`
	SlackChannel                   string   `yaml:"slack-channel,omitempty"`
	SlackMsgDetails                []string `yaml:"slack-message-details,omitempty"`
	SopsAgeKeys                    string   `yaml:"sops-age-keys,omitempty"`
	SopsBucketName                 string   `yaml:"sops-bucket-name,omitempty"`
	AWSMFAProfile                  string   `yaml:"aws-mfa-profile,omitempty"`
	AWSMFATokenExpiration          string   `yaml:"aws-mfa-token-expiration,omitempty"`
	*aws_provider.AwsConfigure     `yaml:"aws,omitempty"`
	*azure_provider.AzureConfigure `yaml:"azure,omitempty"`
	*google_provider.GCPConfigure  `yaml:"gcp,omitempty"`
	ProgressBar                    bool `yaml:"progress-bar"`
	ProjectFile                    `yaml:"project-file"`
}

type HookMapping struct {
	Tenant        string `yaml:"tenant,omitempty"`
	Exists        bool   `yaml:"-"`
	InheritedFrom string `yaml:"inherited-from,omitempty"`
	*Package
}

type Package struct {
	Name           string   `yaml:"name,omitempty"`
	Version        string   `yaml:"version,omitempty"`
	Url            string   `yaml:"url,omitempty"`
	Checksum       string   `yaml:"checksum,omitempty"`
	Artifacts      []string `yaml:"-"`
	HelmfileTenant string   `yaml:"-"`
	OsLinux        string   `yaml:"os-linux,omitempty"`
	OsMac          string   `yaml:"os-mac,omitempty"`
	Os             string   `yaml:"-"`
	Rename         bool     `yaml:"rename,omitempty"`
	GitHubToken    string   `yaml:"-"`
	DstPath        string   `yaml:"dst-path,omitempty"`
}

type Inventory struct {
	HelmPlugins map[string]*Package `yaml:"helm-plugins,omitempty"`
	Hooks       map[string]*Package `yaml:"hooks,omitempty"`
	Tools       map[string]*Package `yaml:"tools,omitempty"`
}

type Project struct {
	Dependencies []Package     `yaml:"dependencies,omitempty"`
	HooksMapping []HookMapping `yaml:"hooks-mapping,omitempty"`
	Spec         struct {
		Environments map[string]*ProjectRootDomain `yaml:"environments,omitempty"`
		Owners       []string                      `yaml:"owners,omitempty"`
		Scopes       []string                      `yaml:"scopes,omitempty"`
	} `yaml:"spec,omitempty"`
}

type ProjectFile struct {
	Project   `yaml:"project,omitempty"`
	Inventory `yaml:"inventory,omitempty"`
}

type ProjectRootDomain struct {
	RootDomain string `yaml:"root-domain,omitempty"`
}

func (conf *Config) InitConfig() *Config {
	conf.ProjectFile = ProjectFile{}
	if err := conf.ReadProjectFile(util.GetPwdPath(util.TenantProjectFile)); err != nil {
		zap.S().Fatal(err)
	}

	return conf
}

func (conf *Config) SerializeConfig() ([]byte, error) {
	var data bytes.Buffer
	encoder := yaml.NewEncoder(&data)
	encoder.SetIndent(2)
	err := encoder.Encode(&conf)
	return data.Bytes(), err
}

func (conf *Config) SerializeJsonConfig() ([]byte, error) {
	return json.Marshal(&conf)
}

func (conf *Config) GetConfigs(all bool) error {
	var (
		patternTenant  string
		patternTaskNum *regexp.Regexp
		patternSemVer  *regexp.Regexp
		patternBranch  *regexp.Regexp
	)

	configsPath := util.GetHomePath(util.RMKDir, util.RMKConfig)

	if all {
		patternTenant = ""
	} else {
		patternTenant = conf.Tenant

		patternBranch = regexp.MustCompile(`^` + patternTenant +
			`-(` + git_handler.DefaultDevelop + `|` + git_handler.DefaultStaging + `|` + git_handler.DefaultProduction + `)$`)
		patternSemVer = regexp.MustCompile(`^` + patternTenant + `-v\d+-\d+-\d+(-[a-z]+)?$`)
		patternTaskNum = regexp.MustCompile(`^` + patternTenant + `-[a-z]+-\d+$`)
	}

	match, err := util.WalkMatch(configsPath, patternTenant+"*.yaml")
	if err != nil {
		return err
	}

	for _, val := range match {
		rmkConfig := strings.TrimSuffix(filepath.Base(val), filepath.Ext(filepath.Base(val)))

		if all {
			fmt.Printf("- %s\n", rmkConfig)
		} else {
			switch {
			case patternBranch.MatchString(rmkConfig):
				fmt.Printf("- %s\n", rmkConfig)
			case patternSemVer.MatchString(rmkConfig):
				fmt.Printf("- %s\n", rmkConfig)
			case patternTaskNum.MatchString(rmkConfig):
				fmt.Printf("- %s\n", rmkConfig)
			}
		}
	}

	return nil
}

func (conf *Config) SetRootDomain(c *cli.Context, gitSpecID string) error {
	for env, val := range conf.Spec.Environments {
		check := strings.Split(val.RootDomain, "*.")
		if len(check) > 2 {
			return fmt.Errorf("root-domain not set correctly for environment %s", env)
		}
	}

	for env, val := range conf.Spec.Environments {
		if env == conf.Environment {
			if regexp.MustCompile(`^\*\.`).MatchString(val.RootDomain) {
				conf.RootDomain = strings.ReplaceAll(val.RootDomain, "*", gitSpecID)
			} else if len(val.RootDomain) > 0 {
				conf.RootDomain = val.RootDomain
			} else {
				conf.RootDomain = "localhost"
			}
		}
	}

	return nil
}

func (pf *ProjectFile) ReadProjectFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &pf); err != nil {
		return err
	}

	return pf.parseProjectFileData()
}

func (pf *ProjectFile) ParseTemplate(tpl *template.Template, data interface{}, value string) (string, error) {
	var text bytes.Buffer
	defer text.Reset()
	parse, err := tpl.Parse(value)
	if err != nil {
		return "", err
	}

	err = parse.Execute(&text, &data)
	if err != nil {
		return "", err
	}

	return text.String(), nil
}

func (pf *ProjectFile) parseProjectFileData() error {
	var err error

	for key, dep := range pf.Dependencies {
		pf.Dependencies[key].Url, err = pf.ParseTemplate(template.New("Dependencies"), pf.Dependencies[key], dep.Url)
		if err != nil {
			return err
		}
	}

	for key, plugin := range pf.HelmPlugins {
		if _, err := semver.NewVersion(plugin.Version); err != nil {
			return fmt.Errorf("%s %s for section inventory.helm-plugins", strings.ToLower(err.Error()), plugin.Version)
		}

		pf.HelmPlugins[key].Name = key
		pf.HelmPlugins[key].Url, err = pf.ParseTemplate(template.New("HelmPlugins"), pf.HelmPlugins[key], plugin.Url)
		if err != nil {
			return err
		}
	}

	for key, hook := range pf.Hooks {
		if _, err := semver.NewVersion(hook.Version); err != nil {
			return fmt.Errorf("%s %s for section inventory.hooks", strings.ToLower(err.Error()), hook.Version)
		}

		pf.Hooks[key].Name = key
		pf.Hooks[key].Url, err = pf.ParseTemplate(template.New("Hooks"), pf.Hooks[key], hook.Url)
		if err != nil {
			return err
		}
	}

	for key, tool := range pf.Tools {
		if _, err := semver.NewVersion(tool.Version); err != nil {
			return fmt.Errorf("%s %s for section inventory.tools", strings.ToLower(err.Error()), tool.Version)
		}

		osDetect := runtime.GOOS
		pf.Tools[key].Name = key

		switch osDetect {
		case "darwin":
			pf.Tools[key].Os = pf.Tools[key].OsMac
		case "linux":
			pf.Tools[key].Os = pf.Tools[key].OsLinux
		default:
			return fmt.Errorf("OS %s is not supported by RMK", osDetect)
		}

		pf.Tools[key].Url, err = pf.ParseTemplate(template.New("Tools"), pf.Tools[key], tool.Url)
		if err != nil {
			return err
		}

		pf.Tools[key].Checksum, err = pf.ParseTemplate(template.New("Tools"), pf.Tools[key], tool.Checksum)
		if err != nil {
			return err
		}
	}

	return nil
}

func (conf *Config) ReadConfigFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, &conf)
}

func (conf *Config) CreateConfigFile() error {
	if err := os.MkdirAll(util.GetHomePath(util.RMKDir, util.RMKConfig), 0755); err != nil {
		return err
	}

	data, err := conf.SerializeConfig()
	if err != nil {
		return err
	}

	return os.WriteFile(util.GetHomePath(util.RMKDir, util.RMKConfig, conf.Name+".yaml"), data, 0644)
}
