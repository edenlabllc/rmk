package commands

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"

	"rmk/config"
	"rmk/git_handler"
	"rmk/notification"
	"rmk/system"
)

type ProjectCommands struct {
	*parseContent
	projectFile *config.ProjectFile
	*projectSpec
	*ReleaseCommands
}

type environment struct {
	globalsPath  string
	releasesPath string
	secretsPath  string
	valuesPath   string
}

type scope struct {
	name         string
	environments map[string]*environment
}

type parseContent struct {
	Dependencies       []string
	EnvironmentName    string
	HelmfileParts      []string
	RepoName           string
	Scopes             []string
	TenantName         string
	TenantNameEnvStyle string
}

type projectSpec struct {
	scopes []scope
	owners string
}

func newProjectCommand(conf *config.Config, ctx *cli.Context, workDir string) *ProjectCommands {
	return &ProjectCommands{
		&parseContent{
			TenantName:         conf.Tenant,
			TenantNameEnvStyle: strings.ToUpper(regexp.MustCompile(`[\-.]`).ReplaceAllString(conf.Tenant, "_")),
		},
		&config.ProjectFile{},
		&projectSpec{owners: codeOwners},
		&ReleaseCommands{Conf: conf, Ctx: ctx, WorkDir: workDir},
	}
}

func (p *ProjectCommands) readProjectFile() error {
	if !system.IsExists(system.GetPwdPath(system.TenantProjectFile), true) {
		return fmt.Errorf("%s file not found", system.GetPwdPath(system.TenantProjectFile))
	}

	data, err := os.ReadFile(system.GetPwdPath(system.TenantProjectFile))
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &p.projectFile); err != nil {
		return err
	}

	return nil
}

func (p *ProjectCommands) serializeProjectFile() ([]byte, error) {
	var count = 0

	if err := p.readProjectFile(); err != nil {
		return nil, err
	}

	for key, pkg := range p.projectFile.Dependencies {
		if pkg.Name == p.Ctx.String("dependency") && pkg.Version != p.Ctx.String("version") {
			p.projectFile.Dependencies[key].Version = p.Ctx.String("version")
			count++
			zap.S().Infof("version changed for dependency %s, affected file: %s",
				pkg.Name,
				system.GetPwdPath(system.TenantProjectFile))
			break
		}
	}

	if count == 0 {
		zap.S().Infof("version %s for dependency %s is not updated",
			p.Ctx.String("version"), p.Ctx.String("dependency"))
		return nil, nil
	}

	var buf bytes.Buffer

	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&p.projectFile); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p *ProjectCommands) genMsgCommit() string {
	msg := fmt.Sprintf("Auto version update %s for dependency: %s",
		p.Ctx.String("version"), p.Ctx.String("dependency"))

	if p.Ctx.Bool("skip-ci") {
		return fmt.Sprintf("%s %s", "[skip ci]", msg)
	}

	return msg
}

func (p *ProjectCommands) updateProjectFile(gitSpec *git_handler.GitSpec) error {
	data, err := p.serializeProjectFile()
	if err != nil {
		return err
	}

	if data != nil {
		if err := os.WriteFile(system.GetPwdPath(system.TenantProjectFile), data, 0644); err != nil {
			return err
		}

		if !p.Ctx.Bool("skip-commit") {
			if err := gitSpec.GitCommitPush(
				system.GetPwdPath(system.TenantProjectFile),
				p.genMsgCommit(), p.Conf.GitHubToken); err != nil {
				return err
			}
		}

		tmp := &notification.TmpUpdate{Config: p.Conf, Context: p.Ctx}
		tmp.ChangesList = append(tmp.ChangesList, p.Ctx.String("dependency"))
		tmp.PathToFile = system.TenantProjectFile
		if err := notification.SlackInit(tmp,
			notification.SlackTmp(tmp).TmpProjectUpdateMsg()).SlackDeclareNotify(); err != nil {
			return err
		}
	}

	return nil
}

func (p *ProjectCommands) writeProjectFiles(path, data string) error {
	if system.IsExists(path, true) {
		zap.S().Warnf("file %s already exists", path)
		return nil
	}

	if len(data) == 0 {
		return nil
	}

	if err := os.WriteFile(path, []byte(data), 0644); err != nil {
		return err
	}

	zap.S().Infof("file %s generated", path)

	return nil
}

func (p *ProjectCommands) generateReadme(gitSpec *git_handler.GitSpec) error {
	p.RepoName = gitSpec.RepoName

	readmeF, err := p.Conf.ParseTemplate(template.New("README"), &p.parseContent, readmeFile)
	if err != nil {
		return err
	}

	if err := p.writeProjectFiles(system.GetPwdPath(system.ReadmeFileName), readmeF); err != nil {
		return err
	}

	return nil
}

func (p *ProjectCommands) generateHelmfile() error {
	sort.Strings(p.projectFile.Spec.Environments)

	for key, name := range p.projectFile.Spec.Environments {
		p.EnvironmentName = name
		hEnvironments, err := p.Conf.ParseTemplate(template.New("Helmfile"), &p.parseContent, helmfileEnvironments)
		if err != nil {
			return err
		}

		if key == 0 {
			p.HelmfileParts = append(p.HelmfileParts, fmt.Sprintf("environments:\n%s", hEnvironments))
		} else {
			p.HelmfileParts = append(p.HelmfileParts, hEnvironments)
		}
	}

	p.HelmfileParts = append(p.HelmfileParts, helmDefaults)

	hInclude, err := p.Conf.ParseTemplate(template.New("Helmfile"), &p.parseContent, helmfiles)
	if err != nil {
		return err
	}

	p.HelmfileParts = append(p.HelmfileParts, hInclude, helmfileMissingFileHandler)

	hCommonLabels, err := p.Conf.ParseTemplate(template.New("Helmfile"), &p.parseContent, helmfileCommonLabels)
	if err != nil {
		return err
	}

	p.HelmfileParts = append(p.HelmfileParts, hCommonLabels, helmfileTemplates)

	hReleases, err := p.Conf.ParseTemplate(template.New("Helmfile"), &p.parseContent, helmfileReleases)
	if err != nil {
		return err
	}

	p.HelmfileParts = append(p.HelmfileParts, hReleases)

	if err := p.writeProjectFiles(system.GetPwdPath(system.HelmfileGoTmplName), strings.Join(p.HelmfileParts, "\n")); err != nil {
		return err
	}

	return nil
}

func (p *ProjectCommands) generateProjectFiles(gitSpec *git_handler.GitSpec) error {
	for _, sc := range p.scopes {
		for _, env := range sc.environments {
			switch sc.name {
			case "clusters":
				if err := p.writeProjectFiles(filepath.Join(env.valuesPath, system.TerraformVarsFile), clusterVariables); err != nil {
					return err
				}

				if err := p.writeProjectFiles(filepath.Join(env.valuesPath, system.TerraformWGFile), clusterWorkerGroups); err != nil {
					return err
				}
			case p.TenantName:
				tGlobals, err := p.Conf.ParseTemplate(template.New("TenantGlobals"), &p.parseContent, tenantGlobals)
				if err != nil {
					return err
				}

				if err := p.writeProjectFiles(env.globalsPath, tGlobals); err != nil {
					return err
				}

				tReleases, err := p.Conf.ParseTemplate(template.New("TenantReleases"), &p.parseContent, tenantReleasesFile)
				if err != nil {
					return err
				}

				if err := p.writeProjectFiles(env.releasesPath, tReleases); err != nil {
					return err
				}

				tSecretSpec, err := p.Conf.ParseTemplate(template.New("TenantSecretSpec"), &p.parseContent, tenantSecretSpecFile)
				if err != nil {
					return err
				}

				if err := p.writeProjectFiles(filepath.Join(env.secretsPath, system.SecretSpecFile), tSecretSpec); err != nil {
					return err
				}

				if err := p.writeProjectFiles(filepath.Join(env.valuesPath, p.TenantName+"-app.yaml.gotmpl"), tenantValuesExample); err != nil {
					return err
				}
			default:
				if err := p.writeProjectFiles(env.globalsPath, globals); err != nil {
					return err
				}

				if err := p.writeProjectFiles(env.releasesPath, releasesFile); err != nil {
					return err
				}

				if err := p.writeProjectFiles(filepath.Join(env.secretsPath, system.SecretSpecFile), secretSpecFile); err != nil {
					return err
				}
			}

			if sc.name != "clusters" {
				if err := p.writeProjectFiles(filepath.Join(env.secretsPath, system.SopsConfigFile), sopsConfigFile); err != nil {
					return err
				}
			}
		}
	}

	if err := p.writeProjectFiles(system.GetPwdPath(system.TenantProjectGitIgn), gitignore); err != nil {
		return err
	}

	if err := p.writeProjectFiles(system.GetPwdPath(system.TenantProjectCodeOwners), p.owners); err != nil {
		return err
	}

	if err := p.generateHelmfile(); err != nil {
		return err
	}

	if err := p.generateReadme(gitSpec); err != nil {
		return err
	}

	return nil
}

func (p *ProjectCommands) generateProject(gitSpec *git_handler.GitSpec) error {
	if err := p.readProjectFile(); err != nil {
		return err
	}

	for _, pkg := range p.projectFile.Dependencies {
		p.Dependencies = append(p.Dependencies, pkg.Name)
	}

	if len(p.projectFile.Spec.Owners) > 0 {
		p.owners = p.owners + fmt.Sprintf("* @%s\n", strings.Join(p.projectFile.Spec.Owners, " @"))
	} else {
		p.owners = ""
	}

	if reflect.ValueOf(p.projectFile.Spec).IsZero() {
		return fmt.Errorf("'spec' option required in %s", system.TenantProjectFile)
	}

	switch {
	case len(p.projectFile.Spec.Scopes) == 0 && len(p.projectFile.Spec.Environments) > 0:
		return fmt.Errorf("'scopes' option required, if 'environments' specified in %s", system.TenantProjectFile)
	case len(p.projectFile.Spec.Scopes) > 0 && len(p.projectFile.Spec.Environments) == 0:
		return fmt.Errorf("'environments' option required, if 'scopes' specified in %s", system.TenantProjectFile)
	case len(p.projectFile.Spec.Scopes) == 0 && len(p.projectFile.Spec.Environments) == 0:
		return fmt.Errorf("'scopes', 'environments' options required in %s", system.TenantProjectFile)
	}

	for sKey, sc := range p.projectFile.Spec.Scopes {
		p.Scopes = append(p.Scopes, sc)
		p.scopes = append(p.scopes, scope{name: sc, environments: make(map[string]*environment)})
		for _, env := range p.projectFile.Spec.Environments {
			if sc == "clusters" {
				p.scopes[sKey].environments[env] = &environment{
					secretsPath: system.GetPwdPath(system.TenantValuesDIR, sc, p.Conf.ClusterProvider, env, "secrets"),
					valuesPath:  system.GetPwdPath(system.TenantValuesDIR, sc, p.Conf.ClusterProvider, env, "values"),
				}

				continue
			}

			p.scopes[sKey].environments[env] = &environment{
				globalsPath:  system.GetPwdPath(system.TenantValuesDIR, sc, env, system.GlobalsFileName),
				releasesPath: system.GetPwdPath(system.TenantValuesDIR, sc, env, system.ReleasesFileName),
				secretsPath:  system.GetPwdPath(system.TenantValuesDIR, sc, env, "secrets"),
				valuesPath:   system.GetPwdPath(system.TenantValuesDIR, sc, env, "values"),
			}
		}
	}

	for _, sc := range p.scopes {
		for _, env := range sc.environments {
			if err := os.MkdirAll(env.secretsPath, 0755); err != nil {
				return err
			}

			if err := os.MkdirAll(env.valuesPath, 0755); err != nil {
				return err
			}
		}
	}

	if err := os.MkdirAll(system.GetPwdPath("docs"), 0755); err != nil {
		return err
	}

	if err := p.generateProjectFiles(gitSpec); err != nil {
		return err
	}

	if p.Ctx.Bool("create-sops-age-keys") {
		if err := newSecretCommands(p.Conf, p.Ctx, system.GetPwdPath()).CreateKeys(); err != nil {
			return err
		}
	}

	return nil
}

func projectGenerateAction(conf *config.Config, gitSpec *git_handler.GitSpec) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := newProjectCommand(conf, c, system.GetPwdPath()).generateProject(gitSpec); err != nil {
			return err
		}

		return resolveDependencies(conf.InitConfig(false), c, false)
	}
}

func projectUpdateAction(conf *config.Config, gitSpec *git_handler.GitSpec) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		return newProjectCommand(conf, c, system.GetPwdPath()).updateProjectFile(gitSpec)
	}
}
