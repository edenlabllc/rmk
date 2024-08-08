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
	"gopkg.in/yaml.v3"

	"rmk/util"
)

// Custom name function for parsing template
const (
	Prompt      = "prompt"
	RequiredEnv = "requiredEnv"
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
	for key, val := range map[string]interface{}{RequiredEnv: requiredEnv, Prompt: prompt} {
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

		if err := yaml.Unmarshal(genFunc.tplString.Bytes(), &genSpec); err != nil {
			return err
		}

		if err := genSpec.writeSpecSecrets(sc.Ctx.Bool("force")); err != nil {
			return err
		}
	}

	return nil
}
