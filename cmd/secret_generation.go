package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"
	"rmk/providers/aws_provider"
	"rmk/providers/azure_provider"
	"rmk/providers/google_provider"
	"rmk/util"
)

// Custom name function for parsing template
const (
	FetchSecretValue = "fetchSecretValue"
	Prompt           = "prompt"
	RequiredEnv      = "requiredEnv"
)

type GenerationSpec struct {
	GenerationRules []GenerationRule `yaml:"generation-rules"`
	secretsDir      string
}

type GenerationFuncMap struct {
	preRender bool
	funcMap   template.FuncMap
	tplString *bytes.Buffer
}

type GenerationRule struct {
	Name     string `yaml:"name"`
	Template string `yaml:"template"`
}

func prompt(name string) (string, error) {
	fmt.Printf("Enter %s: ", name)
	passwd, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Printf("\n")
	if err != nil {
		return "", err
	}

	return string(passwd), nil
}

func requiredEnv(name string) (string, error) {
	if val, exists := os.LookupEnv(name); exists && len(val) > 0 {
		return val, nil
	}

	return "", fmt.Errorf("required env var %s is not set", name)
}

func (gf *GenerationFuncMap) createFuncMap() {
	gf.funcMap = sprig.TxtFuncMap()
	for key, val := range map[string]interface{}{
		FetchSecretValue: util.ValsFetchSecretValue,
		Prompt:           prompt,
		RequiredEnv:      requiredEnv,
	} {
		gf.funcMap[key] = val
	}
}

func (gf *GenerationFuncMap) newTemplate() *template.Template {
	gf.createFuncMap()
	tmpl := template.New("generate").Funcs(gf.funcMap)
	if gf.preRender {
		tmpl = tmpl.Option("missingkey=zero")
	} else {
		tmpl = tmpl.Option("missingkey=error")
	}

	return tmpl
}

func (gf *GenerationFuncMap) renderSpecTemplate(s string, data ...interface{}) error {
	t, err := gf.newTemplate().Parse(s)
	if err != nil {
		return err
	}

	var tplString bytes.Buffer
	var d interface{}
	if len(data) > 0 {
		d = data[0]
	}

	if err := t.Execute(&tplString, d); err != nil {
		return err
	}

	gf.tplString = &tplString

	return nil
}

func (g *GenerationSpec) writeSpecSecrets(force bool) error {
	for _, rule := range g.GenerationRules {
		if util.IsExists(filepath.Join(g.secretsDir, rule.Name+".yaml"), true) && !force {
			zap.S().Warnf("%s exists, new secret generation was skipped",
				filepath.Join(g.secretsDir, rule.Name+".yaml"))
			continue
		}

		if strings.Contains(rule.Template, Prompt) || strings.Contains(rule.Template, RequiredEnv) {
			genFunc := &GenerationFuncMap{preRender: false}
			if err := genFunc.renderSpecTemplate(rule.Template); err != nil {
				return err
			}

			if err := os.WriteFile(filepath.Join(g.secretsDir, rule.Name+".yaml"),
				genFunc.tplString.Bytes(), 0755); err != nil {
				return err
			}

			zap.S().Infof("generating: %s", filepath.Join(g.secretsDir, rule.Name+".yaml"))

			continue
		}

		if err := os.WriteFile(filepath.Join(g.secretsDir, rule.Name+".yaml"),
			[]byte(rule.Template), 0755); err != nil {
			return err
		}

		zap.S().Infof("generating: %s", filepath.Join(g.secretsDir, rule.Name+".yaml"))
	}

	return nil
}

func (sc *SecretCommands) genSpecSecrets(specFiles []string) error {
	switch sc.Conf.ClusterProvider {
	case aws_provider.AWSClusterProvider:
		if sc.Conf.AwsConfigure != nil {
			if err := sc.Conf.SetAWSCredentialsEnv(false); err != nil {
				return err
			}
		}
	case azure_provider.AzureClusterProvider:
		if sc.Conf.AzureConfigure != nil {
			if err := sc.Conf.AzureConfigure.ReadSPCredentials(sc.Conf.Name); err != nil {
				return err
			}

			if err := sc.Conf.SetAzureCredentialsEnv(false); err != nil {
				return err
			}
		}
	case google_provider.GoogleClusterProvider:
		if sc.Conf.GCPConfigure != nil {
			if err := sc.Conf.SetGCPCredentialsEnv(false); err != nil {
				return err
			}
		}
	}

	genSpec := &GenerationSpec{}

	for _, spec := range specFiles {
		data, err := os.ReadFile(spec)
		if err != nil {
			return err
		}

		genSpec.secretsDir, _ = filepath.Split(spec)

		genFunc := &GenerationFuncMap{preRender: true}
		if err := genFunc.renderSpecTemplate(string(data)); err != nil {
			return err
		}

		_, err = util.YamlDecodeWithComments(spec, genFunc.tplString.Bytes(), genSpec)
		if err != nil {
			return err
		}

		if err := genSpec.writeSpecSecrets(sc.Ctx.Bool("force")); err != nil {
			return err
		}
	}

	return nil
}
