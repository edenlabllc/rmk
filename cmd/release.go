package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
	"mvdan.cc/sh/v3/shell"

	"rmk/config"
	"rmk/git_handler"
	"rmk/notification"
	"rmk/providers/aws_provider"
	"rmk/providers/azure_provider"
	"rmk/providers/google_provider"
	"rmk/providers/onprem_provider"
	"rmk/util"
)

type releaseRunner interface {
	runCMD() error
}

type ReleaseCommands struct {
	Conf          *config.Config
	Ctx           *cli.Context
	SpecCMD       *util.SpecCMD
	Scope         string
	WorkDir       string
	ValuesPath    string
	UpdateContext bool
	APICluster    bool
	K3DCluster    bool
}

type ReleaseStruct struct {
	Enabled bool
	Image   struct {
		Repository string
		Tag        string
	} `yaml:"image,omitempty"`
}

type ReleasesList struct {
	NodeYAML yaml.Node
	Releases map[string]*ReleaseStruct
	Changes  struct {
		List  map[string][]string
		Count int64
	}
}

type SpecRelease struct {
	ReleaseCommands
	ReleasesList
	ReleasesPaths []string
}

type HelmfileList []struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Enabled   bool   `json:"enabled"`
	Installed bool   `json:"installed"`
	Labels    string `json:"labels"`
	Chart     string `json:"chart"`
	Version   string `json:"version"`
}

type HelmStatus struct {
	Name string `json:"name"`
	Info struct {
		FirstDeployed time.Time `json:"first_deployed"`
		LastDeployed  time.Time `json:"last_deployed"`
		Deleted       string    `json:"deleted"`
		Description   string    `json:"description"`
		Status        string    `json:"status"`
		Notes         string    `json:"notes"`
	} `json:"info"`
	Version   int    `json:"version"`
	Namespace string `json:"namespace"`
}

func (rc *ReleaseCommands) runCMD() error {
	if err := rc.SpecCMD.AddOSEnv(); err != nil {
		return err
	}

	if err := rc.SpecCMD.ExecCMD(); err != nil {
		rc.debugLevel()

		return err
	}

	rc.debugLevel()

	return nil
}

func (rc *ReleaseCommands) debugLevel() {
	if rc.SpecCMD.Debug {
		zap.S().Debugf("command: %s", rc.SpecCMD.CommandStr)
		zap.S().Debugf("path: %s", rc.SpecCMD.Dir)
		for _, val := range rc.SpecCMD.Envs {
			if len(rc.Conf.GitHubToken) > 0 {
				zap.S().Debugf("env: %s", strings.ReplaceAll(val, rc.Conf.GitHubToken, "[rmk_sensitive]"))
			} else {
				zap.S().Debugf("env: %s", val)
			}
		}
	}
}

func (rc *ReleaseCommands) nestedHelmfiles(envs ...string) []string {
	var hfPath, hfVersion []string

	for _, val := range rc.Conf.Dependencies {
		hfPath = append(hfPath, fmt.Sprintf(`{"path":"%s"}`, val.DstPath))
		keyVerEnv := regexp.MustCompile(`[\-.]`).ReplaceAllString(val.Name, "_")
		hfVersion = append(hfVersion, "HELMFILE_"+strings.ToUpper(keyVerEnv)+"_VERSION="+val.Version)
	}

	keyTenantEnv := regexp.MustCompile(`[\-.]`).ReplaceAllString(rc.Conf.Tenant, "_")
	envs = append(envs, "HELMFILE_"+strings.ToUpper(keyTenantEnv)+"_PATHS=["+strings.Join(hfPath, ",")+"]")
	return append(envs, hfVersion...)
}

func (rc *ReleaseCommands) prepareHelmfile(args ...string) *util.SpecCMD {
	var sensKeyWords []string

	defaultArgs := []string{"--environment", rc.Conf.Environment}

	// generating common environment variables
	envs := append([]string{},
		"NAME="+rc.Conf.Name,
		"ROOT_DOMAIN="+rc.Conf.RootDomain,
		"SOPS_AGE_KEY_FILE="+filepath.Join(rc.Conf.SopsAgeKeys, util.SopsAgeKeyFile),
		"TENANT="+rc.Conf.Tenant,
	)

	if len(rc.Conf.GitHubToken) > 0 {
		sensKeyWords = []string{rc.Conf.GitHubToken}
		envs = append(envs, "GITHUB_TOKEN="+rc.Conf.GitHubToken)
	} else {
		sensKeyWords = []string{}
	}

	// generating additional environment variables to specific cluster provider
	switch rc.Conf.ClusterProvider {
	case aws_provider.AWSClusterProvider:
		envs = append(envs,
			"AWS_ACCOUNT_ID="+rc.Conf.AccountID,
			"AWS_CLUSTER=true",
			"AWS_CONFIG_FILE="+strings.Join(rc.Conf.AWSSharedConfigFile(rc.Conf.Profile), ""),
			"AWS_PROFILE="+rc.Conf.Profile,
			"AWS_REGION="+rc.Conf.Region,
			"AWS_SHARED_CREDENTIALS_FILE="+strings.Join(rc.Conf.AWSSharedCredentialsFile(rc.Conf.Profile), ""),
		)
	case azure_provider.AzureClusterProvider:
		envs = append(envs,
			"AZURE_CLUSTER=true",
			"AZURE_LOCATION="+rc.Conf.AzureConfigure.Location,
			"AZURE_SUBSCRIPTION_ID="+rc.Conf.AzureConfigure.SubscriptionID,
		)
	case google_provider.GoogleClusterProvider:
		envs = append(envs,
			"GCP_CLUSTER=true",
			"GCP_PROJECT_ID="+rc.Conf.GCPConfigure.ProjectID,
			"GCP_REGION="+rc.Conf.GCPRegion,
			"GOOGLE_APPLICATION_CREDENTIALS="+rc.Conf.GCPConfigure.AppCredentialsPath,
		)
	case onprem_provider.OnPremClusterProvider:
		envs = append(envs,
			"ONPREM_CLUSTER=true",
		)
	}

	for _, val := range rc.Conf.HooksMapping {
		keyTenantEnv := regexp.MustCompile(`[\-.]`).ReplaceAllString(val.Tenant, "_")
		envs = append(envs, "HELMFILE_"+strings.ToUpper(keyTenantEnv)+"_HOOKS_DIR="+val.DstPath)
	}

	// generating additional environment variables to nested helmfiles
	envs = rc.nestedHelmfiles(envs...)

	switch {
	case rc.APICluster:
		envs = append(envs, "CAPI_CLUSTER="+strconv.FormatBool(rc.APICluster))
	case rc.K3DCluster:
		envs = append(envs, "K3D_CLUSTER="+strconv.FormatBool(rc.K3DCluster))
	}

	if len(rc.Ctx.String("helmfile-log-level")) > 0 {
		defaultArgs = append(defaultArgs, "--log-level", rc.Ctx.String("helmfile-log-level"))
	}

	return &util.SpecCMD{
		Args:         append(defaultArgs, args...),
		Command:      "helmfile",
		Ctx:          rc.Ctx,
		Dir:          rc.WorkDir,
		Envs:         envs,
		Debug:        true,
		SensKeyWords: sensKeyWords,
	}
}

func (rc *ReleaseCommands) releaseMiddleware() error {
	if len(rc.Conf.Dependencies) == 0 {
		if err := os.RemoveAll(filepath.Join(rc.WorkDir, TenantPrDependenciesDir)); err != nil {
			return err
		}
	}

	if err := util.MergeAgeKeys(rc.Conf.SopsAgeKeys); err != nil {
		return err
	}

	if _, currentContext, err := clusterRunner(&ClusterCommands{rc}).getKubeContext(); err != nil {
		return err
	} else {
		switch {
		case strings.Contains(currentContext, util.K3DPrefix) && !strings.Contains(currentContext, util.CAPI):
			rc.K3DCluster = true
		case currentContext == util.K3DPrefix+"-"+util.CAPI:
			rc.APICluster = true
		}
	}

	return nil
}

func (rc *ReleaseCommands) releaseHelmfile(args ...string) error {
	if err := rc.releaseMiddleware(); err != nil {
		return err
	}

	rc.SpecCMD = rc.prepareHelmfile(args...)

	return rc.runCMD()
}

func (sr *SpecRelease) searchReleasesPath() error {
	paths, err := util.WalkInDir(util.GetPwdPath(util.TenantValuesDIR), sr.Conf.Environment, util.ReleasesFileName)
	if err != nil {
		return err
	}

	sr.ReleasesPaths = paths
	return nil
}

func (sr *SpecRelease) readReleasesFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, &sr.NodeYAML); err != nil {
		return err
	}

	if err := sr.NodeYAML.Decode(&sr.Releases); err != nil {
		return err
	}

	for key, val := range sr.Releases {
		if len(val.Image.Repository) > 0 && val.Image.Repository == sr.Ctx.String("repository") &&
			val.Image.Tag != sr.Ctx.String("tag") {
			val.Image.Tag = sr.Ctx.String("tag")
			sr.Changes.List[path] = append(sr.Changes.List[path], key)
			sr.Changes.Count++
		}
	}

	sort.Strings(sr.Changes.List[path])

	return nil
}

func (sr *SpecRelease) serializeReleasesStruct() ([]byte, error) {
	headComment := sr.NodeYAML.HeadComment
	footComment := sr.NodeYAML.FootComment

	if err := sr.NodeYAML.Encode(&sr.Releases); err != nil {
		return nil, err
	}

	sr.NodeYAML.HeadComment = fmt.Sprintf("%s\r\n\n", headComment)
	sr.NodeYAML.FootComment = footComment

	var data bytes.Buffer

	encoder := yaml.NewEncoder(&data)
	encoder.SetIndent(2)
	err := encoder.Encode(&sr.NodeYAML)
	return data.Bytes(), err
}

func (sr *SpecRelease) updateReleasesFile(g *git_handler.GitSpec) error {
	if err := sr.searchReleasesPath(); err != nil {
		return err
	}

	if len(sr.ReleasesPaths) == 0 {
		return fmt.Errorf("no files %s found", util.ReleasesFileName)
	}

	sr.Changes.List = make(map[string][]string)

	for _, path := range sr.ReleasesPaths {
		sr.Releases = make(map[string]*ReleaseStruct)
		if err := sr.readReleasesFile(path); err != nil {
			return err
		}

		for key, val := range sr.Changes.List {
			if key == path {
				data, err := sr.serializeReleasesStruct()
				if err != nil {
					return err
				}

				zap.S().Infof("tag changed for next releases %s, "+
					"affected file: %s", strings.Join(val, " "), path)

				if err := os.WriteFile(path, data, 0644); err != nil {
					return err
				}
			}
		}
	}

	if sr.Changes.Count == 0 {
		zap.S().Info("no image tag found to update by repositories URL")
		return nil
	} else if sr.Ctx.Bool("deploy") || sr.Ctx.Bool("commit") {
		tmp := &notification.TmpUpdate{Config: sr.Conf, Context: sr.Ctx}
		for key := range sr.Changes.List {
			if err := g.GitCommitPush(key, sr.genMsgCommit(key), sr.Conf.GitHubToken); err != nil {
				return err
			}

			tmp.PathToFile = key
			tmp.ChangesList = sr.Changes.List[key]
			if err := notification.SlackInit(tmp,
				notification.SlackTmp(tmp).TmpReleaseUpdateMsg()).SlackDeclareNotify(); err != nil {
				return err
			}
		}

		if sr.Ctx.Bool("deploy") {
			if errDep := sr.deployUpdatedReleases(); errDep != nil {
				if err := notification.SlackInit(tmp,
					notification.SlackTmp(tmp).TmpReleaseUpdateFailedMsg(errDep)).SlackFailNotify(); err != nil {
					return err
				}

				return errDep
			}

			if err := notification.SlackInit(tmp,
				notification.SlackTmp(tmp).TmpReleaseUpdateSuccessMsg()).SlackSuccessNotify(); err != nil {
				return err
			}
		}
	}

	return nil
}

func (sr *SpecRelease) genMsgCommit(path string) string {
	if sr.Ctx.Bool("skip-ci") {
		return fmt.Sprintf("%s Auto version update %s for releases: %s",
			"[skip ci]",
			sr.Ctx.String("tag"),
			strings.Join(sr.Changes.List[path], ","))
	}

	return fmt.Sprintf("Auto version update %s for releases: %s",
		sr.Ctx.String("tag"),
		strings.Join(sr.Changes.List[path], ","))
}

func (sr *SpecRelease) deployUpdatedReleases() error {
	if err := sr.releaseMiddleware(); err != nil {
		return err
	}

	if err := sr.checkStatusRelease(); err != nil {
		return err
	}

	sr.SpecCMD = sr.prepareHelmfile()
	for _, values := range sr.Changes.List {
		for _, val := range values {
			sr.SpecCMD.Args = append(sr.SpecCMD.Args, "--selector", "app="+val)
		}
	}

	sr.SpecCMD.Args = append(sr.SpecCMD.Args, "sync")

	return sr.runCMD()
}

func (rc *ReleaseCommands) helmCommands(args ...string) *util.SpecCMD {
	return &util.SpecCMD{
		Args:          args,
		Command:       "helm",
		Ctx:           rc.Ctx,
		Dir:           rc.WorkDir,
		Debug:         true,
		DisableStdOut: true,
		Envs: append(rc.SpecCMD.Envs,
			"AWS_PROFILE="+rc.Conf.Profile,
			"AWS_CONFIG_FILE="+strings.Join(rc.Conf.AWSSharedConfigFile(rc.Conf.Profile), ""),
			"AWS_SHARED_CREDENTIALS_FILE="+strings.Join(rc.Conf.AWSSharedCredentialsFile(rc.Conf.Profile), ""),
		),
	}
}

func (sr *SpecRelease) deserializeHelmStatus() *HelmStatus {
	helmStatus := &HelmStatus{}

	regex, err := regexp.Compile("\n\n")
	if err != nil {
		zap.S().Fatal(err)
	}

	if len(sr.SpecCMD.StdoutBuf.String()) == 0 {
		return helmStatus
	}

	if err := json.Unmarshal([]byte(regex.ReplaceAllString(sr.SpecCMD.StdoutBuf.String(), "\n")), &helmStatus); err != nil {
		zap.S().Fatalf("can't deserialize Helm status output: %v", err)
	}

	return helmStatus
}

func (sr *SpecRelease) getNamespaceViaHelmfileList(releaseName string) (string, error) {
	sr.SpecCMD = sr.prepareHelmfile("--selector", "name="+releaseName, "list", "--output", "json")
	sr.SpecCMD.DisableStdOut = true
	sr.SpecCMD.Debug = true
	if err := sr.runCMD(); err != nil {
		return "", fmt.Errorf("Helmfile failed to get release %s namespace\n%s", releaseName, sr.SpecCMD.StderrBuf.String())
	}

	helmfileList := HelmfileList{}

	regex, err := regexp.Compile("\n\n")
	if err != nil {
		return "", err
	}

	if len(sr.SpecCMD.StdoutBuf.String()) == 0 {
		return "", nil
	}

	if err := json.Unmarshal([]byte(regex.ReplaceAllString(sr.SpecCMD.StdoutBuf.String(), "\n")), &helmfileList); err != nil {
		return "", fmt.Errorf("can't deserialize Helmfile list command output: %v", err)
	}

	if len(helmfileList) > 0 {
		return helmfileList[0].Namespace, nil
	} else {
		return "", fmt.Errorf("Helmfile return empty list for release: %s", releaseName)
	}
}

func (sr *SpecRelease) releaseStatus(releaseName string) error {
	namespace, err := sr.getNamespaceViaHelmfileList(releaseName)
	if err != nil {
		return err
	}

	sr.SpecCMD = sr.helmCommands("status", "--namespace",
		namespace,
		releaseName,
		"--output",
		"json",
	)

	if err := sr.runCMD(); err != nil {
		return fmt.Errorf("Helm failed to get release %s status\n%s", releaseName, sr.SpecCMD.StderrBuf.String())
	}

	return nil
}

func (sr *SpecRelease) releaseUnlock(status *HelmStatus) error {
	sr.SpecCMD = sr.helmCommands("rollback", status.Name,
		strconv.Itoa(status.Version),
		"-n",
		status.Namespace,
		"--history-max",
		"0",
	)

	if err := sr.runCMD(); err != nil {
		return fmt.Errorf("Helm failed to rollback release: %s\n%s", status.Name, sr.SpecCMD.StderrBuf.String())
	}

	return nil
}

func (sr *SpecRelease) checkStatusRelease() error {
	var helmStatuses []HelmStatus

	for _, values := range sr.Changes.List {
		for _, releaseName := range values {
			if err := sr.releaseStatus(releaseName); err != nil {
				return err
			}

			status := sr.deserializeHelmStatus().Info.Status

			if status == "pending-upgrade" || status == "pending-install" {
				helmStatuses = append(helmStatuses, *sr.deserializeHelmStatus())
			}
		}
	}

	for _, status := range helmStatuses {
		if err := sr.releaseUnlock(&status); err != nil {
			return err
		}

		zap.S().Infof("unlock release %s for namespace %s was done", status.Name, status.Namespace)
	}

	return nil
}

func releaseHelmfileAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		rc := &ReleaseCommands{
			Conf:    conf,
			Ctx:     c,
			WorkDir: util.GetPwdPath(""),
		}

		if !c.Bool("skip-context-switch") {
			if err := clusterRunner(&ClusterCommands{rc}).switchKubeContext(); err != nil {
				return err
			}
		}

		var args []string

		for _, selector := range c.StringSlice("selector") {
			args = append(args, "--selector", selector)
		}

		args = append(args, c.Command.Name)

		if c.IsSet("output") {
			args = append(args, "--output", c.String("output"))
		}

		if c.IsSet("helmfile-args") {
			// parse arguments using shell syntax (fully-compatible with any type of quotes)
			shArgs, err := shell.Fields(c.String("helmfile-args"), func(name string) string { return "" })

			if err != nil {
				return fmt.Errorf("--helmfile-args argument has invalid shell syntax")
			}

			args = append(args, shArgs...)
		}

		return rc.releaseHelmfile(args...)
	}
}

func releaseRollbackAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		sr := &SpecRelease{ReleasesList: ReleasesList{Changes: struct {
			List  map[string][]string
			Count int64
		}{List: make(map[string][]string)}}}
		sr.Conf = conf
		sr.Ctx = c
		sr.WorkDir = util.GetPwdPath("")

		if !c.Bool("skip-context-switch") {
			if err := clusterRunner(&ClusterCommands{&sr.ReleaseCommands}).switchKubeContext(); err != nil {
				return err
			}
		}

		sr.Changes.List["rollback"] = c.StringSlice("release-name")
		if err := sr.checkStatusRelease(); err != nil {
			return err
		}

		return nil
	}
}

func releaseUpdateAction(conf *config.Config, gitSpec *git_handler.GitSpec) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		sr := &SpecRelease{}
		sr.Conf = conf
		sr.Ctx = c
		sr.WorkDir = util.GetPwdPath("")

		if !c.Bool("skip-context-switch") {
			if err := clusterRunner(&ClusterCommands{&sr.ReleaseCommands}).switchKubeContext(); err != nil {
				return err
			}
		}

		return sr.updateReleasesFile(gitSpec)
	}
}
