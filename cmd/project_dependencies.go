package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"rmk/config"
	"rmk/util"
)

const (
	gitProvider  = "git"
	httpProvider = "http"
)

var (
	TenantPrDependenciesDir = filepath.Join(util.TenantProjectDIR, "dependencies")
	TenantPrInventoryDir    = filepath.Join(util.TenantProjectDIR, "inventory")
	TenantPrInvClustersDir  = filepath.Join(TenantPrInventoryDir, "clusters")
	TenantPrInvHooksDir     = filepath.Join(TenantPrInventoryDir, "hooks")
)

type SpecDownload struct {
	Conf     *config.Config
	Ctx      *cli.Context
	PkgUrl   string
	PkgDst   string
	PkgFile  string
	PkgName  string
	Header   *http.Header
	Type     string
	rmOldDir bool
}

type InventoryState struct {
	clustersState    map[string]struct{}
	helmPluginsState map[string]struct{}
	toolsState       map[string]struct{}
}

func (s *SpecDownload) parseUrlByType() error {
	var getter = "git::"
	s.Header = &http.Header{}
	if strings.HasPrefix(s.PkgUrl, getter) {
		s.Type = gitProvider
	} else {
		s.Type = httpProvider
	}

	if s.Type == gitProvider {
		parse, err := url.Parse(strings.TrimPrefix(s.PkgUrl, getter))
		if err != nil {
			return err
		}

		if parse.Query().Has("depth") {
			return fmt.Errorf("for Git provider in project file %s repository only 'ref' argument can be used", s.PkgName)
		}

		query := parse.Query()
		query.Add("depth", "1")
		s.PkgUrl = getter + parse.Scheme + "://user:" + s.Conf.GitHubToken + "@" + parse.Host + parse.Path + "?" + query.Encode()
	}

	return nil
}

func (s *SpecDownload) downloadErrorHandler(err error) error {
	switch {
	case err != nil && s.Type == gitProvider:
		return fmt.Errorf("failed to download %s object %s, potential reasons: "+
			"object not found, permission denied, credentials expired, URL format in project file not correct", s.Type, s.PkgName)
	case err != nil && s.Type == httpProvider:
		return fmt.Errorf("failed to download %s object %s, potential reasons: "+
			"object not found, URL format in project file not correct", s.Type, s.PkgName)
	}

	return err
}

func (s *SpecDownload) download(silent bool) error {
	if err := s.parseUrlByType(); err != nil {
		return err
	}

	if !silent {
		zap.S().Infof("starting package download: %s", s.PkgName)
	}

	return s.downloadErrorHandler(
		util.DownloadArtifact(s.PkgUrl, s.PkgDst, s.PkgName, s.Header, silent, s.Conf.ProgressBar, s.Ctx.Context),
	)
}

// depsGC - deleting old deps dirs with not actual versions
func hooksGC(hooks []config.HookMapping) error {
	if !util.IsExists(util.GetPwdPath(TenantPrInvHooksDir), false) {
		return nil
	}

	allDirs, _, err := util.ListDir(util.GetPwdPath(TenantPrInvHooksDir), true)
	if err != nil {
		return err
	}

	var diff = make(map[string]struct{}, len(allDirs))
	for _, val := range allDirs {
		diff[val] = struct{}{}
	}

	for _, hook := range hooks {
		if _, ok := diff[hook.DstPath]; ok {
			delete(diff, hook.DstPath)
		}
	}

	for dir := range diff {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}
	}

	return nil
}

// resolveHooks - resolve hooks version according to nested project.yaml file
func resolveHooks(hooks map[string]*config.Package, tenant string, conf *config.Config) error {
	if len(hooks) > 0 {
		for _, hook := range hooks {
			conf.HooksMapping = append(conf.HooksMapping,
				config.HookMapping{
					Tenant:  tenant,
					Exists:  true,
					Package: hook,
				},
			)
		}
	} else if len(conf.Dependencies) > 0 {
		conf.HooksMapping = append(conf.HooksMapping,
			config.HookMapping{
				Tenant:  tenant,
				Exists:  false,
				Package: &config.Package{},
			},
		)
	}

	return nil
}

// uniqueHooksMapping - casts a list of hooks to unique values and recursively inherits hook version values
func uniqueHooksMapping(hooks []config.HookMapping) []config.HookMapping {
	var uniqueHooksMapping []config.HookMapping

	for _, hook := range hooks {
		skip := false
		for _, uniqueHook := range uniqueHooksMapping {
			if hook.Exists == uniqueHook.Exists && hook.Tenant == uniqueHook.Tenant {
				skip = true
				break
			}
		}

		if !skip {
			uniqueHooksMapping = append(uniqueHooksMapping, hook)
		}
	}

	numberHook := 0
	compareHooks := make(map[int]*semver.Version)
	for key, hook := range uniqueHooksMapping {
		if hook.Exists {
			ver, _ := semver.NewVersion(hook.Version)
			compareHooks[key] = ver
		}
	}

	for key, ver := range compareHooks {
		if len(compareHooks) > 1 {
			for _, v := range compareHooks {
				if ver.GreaterThan(v) {
					numberHook = key
				}
			}
		} else {
			numberHook = key
		}
	}

	for key, hook := range uniqueHooksMapping {
		if !hook.Exists {
			uniqueHooksMapping[key].Package = uniqueHooksMapping[numberHook].Package
			uniqueHooksMapping[key].InheritedFrom = uniqueHooksMapping[numberHook].Tenant
		}
	}

	return uniqueHooksMapping
}

func (is *InventoryState) saveState(inv config.Inventory) {
	is.clustersState = make(map[string]struct{})
	for key := range inv.Clusters {
		is.clustersState[key] = struct{}{}
	}

	is.helmPluginsState = make(map[string]struct{})
	for key := range inv.HelmPlugins {
		is.helmPluginsState[key] = struct{}{}
	}

	is.toolsState = make(map[string]struct{})
	for key := range inv.Tools {
		is.toolsState[key] = struct{}{}
	}
}

func (is *InventoryState) resolveClusters(invPkg map[string]*config.Package, conf *config.Config) (map[string]*config.Package, error) {
	if len(conf.Clusters) == 0 {
		conf.Clusters = make(map[string]*config.Package)
	}

	for key, pkg := range invPkg {
		vPkg, _ := semver.NewVersion(pkg.Version)
		if _, ok := conf.Clusters[key]; !ok {
			conf.Clusters[key] = pkg
		} else if _, found := is.clustersState[key]; !found {
			vP, _ := semver.NewVersion(conf.Clusters[key].Version)
			if vPkg.GreaterThan(vP) {
				conf.Clusters[key] = pkg
			}
		}
	}

	return conf.Clusters, nil
}

func (is *InventoryState) resolveHelmPlugins(invPkg map[string]*config.Package, conf *config.Config) (map[string]*config.Package, error) {
	if len(conf.HelmPlugins) == 0 {
		conf.HelmPlugins = make(map[string]*config.Package)
	}

	for key, pkg := range invPkg {
		vPkg, _ := semver.NewVersion(pkg.Version)
		if _, ok := conf.HelmPlugins[key]; !ok {
			conf.HelmPlugins[key] = pkg
		} else if _, found := is.helmPluginsState[key]; !found {
			vP, _ := semver.NewVersion(conf.HelmPlugins[key].Version)
			if vPkg.GreaterThan(vP) {
				conf.HelmPlugins[key] = pkg
			}
		}
	}

	return conf.HelmPlugins, nil
}

func (is *InventoryState) resolveTools(invPkg map[string]*config.Package, conf *config.Config) (map[string]*config.Package, error) {
	if len(conf.Tools) == 0 {
		conf.Tools = make(map[string]*config.Package)
	}

	for key, pkg := range invPkg {
		vPkg, _ := semver.NewVersion(pkg.Version)
		if _, ok := conf.Tools[key]; !ok {
			conf.Tools[key] = pkg
		} else if _, found := is.toolsState[key]; !found {
			vP, _ := semver.NewVersion(conf.Tools[key].Version)
			if vPkg.GreaterThan(vP) {
				conf.Tools[key] = pkg
			}
		}
	}

	return conf.Tools, nil
}

func resolveDependencies(conf *config.Config, ctx *cli.Context, silent bool) error {
	var (
		recursivelyDownload func() error
		invErr              error
	)

	if err := updateDependencies(conf, ctx, silent); err != nil {
		return err
	}

	if err := resolveHooks(conf.Hooks, conf.Tenant, conf); err != nil {
		return err
	}

	invState := &InventoryState{}
	invState.saveState(conf.Inventory)

	recursivelyDownload = func() error {
		for _, val := range conf.Dependencies {
			projectFile := &config.ProjectFile{}

			depsDir := util.FindDir(util.GetPwdPath(TenantPrDependenciesDir), val.Name)
			if err := projectFile.ReadProjectFile(util.GetPwdPath(TenantPrDependenciesDir, depsDir, util.TenantProjectFile)); err != nil {
				return err
			}

			// Resolve and recursively download repositories containing clusters
			if conf.Clusters, invErr = invState.resolveClusters(projectFile.Clusters, conf); invErr != nil {
				return invErr
			}

			// Resolve and recursively download repositories containing helm plugins
			if conf.HelmPlugins, invErr = invState.resolveHelmPlugins(projectFile.HelmPlugins, conf); invErr != nil {
				return invErr
			}

			// Resolve repositories containing hooks
			if len(strings.Split(depsDir, ".")) > 0 {
				if err := resolveHooks(projectFile.Hooks, strings.Split(depsDir, ".")[0], conf); err != nil {
					return err
				}
			}

			// Resolve and recursively download repositories containing tools
			if conf.Tools, invErr = invState.resolveTools(projectFile.Tools, conf); invErr != nil {
				return invErr
			}

			// Recursively downloading repositories containing helmfiles
			foundDeps := 0
			compare := make(map[string]struct{}, len(projectFile.Dependencies))
			for _, dep := range projectFile.Dependencies {
				compare[dep.Name] = struct{}{}
			}

			for _, dep := range conf.Dependencies {
				if _, ok := compare[dep.Name]; ok {
					foundDeps++
				}
			}

			if len(projectFile.Dependencies) == 0 {
				foundDeps++
			}

			if foundDeps == 0 {
				conf.Dependencies = append(projectFile.Dependencies, val)
				if err := updateDependencies(conf, ctx, silent); err != nil {
					return err
				}

				if err := recursivelyDownload(); err != nil {
					return err
				}
			}
		}

		return nil
	}

	if err := recursivelyDownload(); err != nil {
		return err
	}

	if err := updateClusters(conf, ctx, silent); err != nil {
		return err
	}

	// Finding unique versions of hooks in HooksMapping
	conf.HooksMapping = uniqueHooksMapping(conf.HooksMapping)

	if err := updateHooks(conf, ctx, silent); err != nil {
		return err
	}

	// Old hooks dirs garbage collection
	if err := hooksGC(conf.HooksMapping); err != nil {
		return err
	}

	if err := updateTools(conf, ctx, silent); err != nil {
		return err
	}

	if err := newConfigCommands(conf, ctx, util.GetPwdPath("")).configHelmPlugins(); err != nil {
		return err
	}

	if err := conf.CreateConfigFile(); err != nil {
		return err
	}

	return nil
}

func removeOldDir(pwd string, pkg config.Package) error {
	if !util.IsExists(pwd, false) {
		return nil
	}

	oldDir := util.FindDir(pwd, pkg.Name)
	if len(strings.Split(oldDir, "-")) > 1 {
		oldVer := strings.SplitN(oldDir, "-", 2)[1]
		if oldVer != pkg.Version {
			if err := os.RemoveAll(filepath.Join(pwd, pkg.Name+"-"+oldVer)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *SpecDownload) batchUpdate(pwd string, pkg config.Package, silent bool) error {
	s.PkgUrl = pkg.Url
	s.PkgName = pkg.Name + "-" + strings.ReplaceAll(pkg.Version, "/", "_")
	s.PkgDst = filepath.Join(s.PkgDst, s.PkgName)
	pkgExists := util.IsExists(s.PkgDst, false)
	if !pkgExists {
		if s.rmOldDir {
			if err := removeOldDir(pwd, pkg); err != nil {
				return err
			}
		}

		if err := s.download(silent); err != nil {
			return err
		}
	}

	return nil
}

func updateDependencies(conf *config.Config, ctx *cli.Context, silent bool) error {
	pwd := util.GetPwdPath(TenantPrDependenciesDir)

	for key, val := range conf.Dependencies {
		spec := &SpecDownload{Conf: conf, Ctx: ctx, PkgDst: pwd, rmOldDir: true}
		if err := spec.batchUpdate(pwd, val, silent); err != nil {
			return err
		}

		// needed if all packages from project.yaml were downloaded earlier
		spec.PkgUrl = val.Url
		if err := spec.parseUrlByType(); err != nil {
			return err
		}

		switch {
		case util.IsExists(filepath.Join(spec.PkgDst, util.HelmfileFileName), true):
			conf.Dependencies[key].DstPath = filepath.Join(spec.PkgDst, util.HelmfileFileName)
		case util.IsExists(filepath.Join(spec.PkgDst, util.HelmfileGoTmplName), true):
			conf.Dependencies[key].DstPath = filepath.Join(spec.PkgDst, util.HelmfileGoTmplName)
		default:
			return fmt.Errorf("%s or %s not found in dependent project %s",
				util.HelmfileFileName, util.HelmfileGoTmplName, spec.PkgName)
		}
	}

	return nil
}

func updateClusters(conf *config.Config, ctx *cli.Context, silent bool) error {
	pwd := util.GetPwdPath(TenantPrInvClustersDir)

	for key, val := range conf.Clusters {
		spec := &SpecDownload{Conf: conf, Ctx: ctx, PkgDst: pwd, rmOldDir: true}
		if err := spec.batchUpdate(pwd, *val, silent); err != nil {
			return err
		}

		conf.Clusters[key].DstPath = spec.PkgDst
	}

	return nil
}

func updateHooks(conf *config.Config, ctx *cli.Context, silent bool) error {
	pwd := util.GetPwdPath(TenantPrInvHooksDir)

	for key, val := range conf.HooksMapping {
		spec := &SpecDownload{Conf: conf, Ctx: ctx, PkgDst: pwd}
		if err := spec.batchUpdate(pwd, *val.Package, silent); err != nil {
			return err
		}

		conf.HooksMapping[key].DstPath = spec.PkgDst
	}

	return nil
}

func match(dir string, patterns []string) ([]string, error) {
	var (
		unique []string
		find   []string
	)

	for _, val := range patterns {
		match, err := util.WalkMatch(dir, val)
		if err != nil {
			return nil, err
		}

		find = append(find, match...)
	}

	for _, val := range find {
		skip := false
		for _, uniq := range unique {
			if val == uniq {
				skip = true
				break
			}
		}

		if !skip {
			unique = append(unique, val)
		}
	}

	return unique, nil
}

func overwriteFiles(path, pattern, name string) error {
	var data []byte

	oldFilePath, err := util.WalkMatch(path, pattern)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(strings.Join(oldFilePath, "")); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(path, name), data, 0755)
}

func updateTools(conf *config.Config, ctx *cli.Context, silent bool) error {
	spec := &SpecDownload{
		Conf:   conf,
		Ctx:    ctx,
		PkgDst: util.GetHomePath(util.RMKDir, util.RMKToolsDir, util.ToolsTmpDir),
		Type:   httpProvider,
	}

	toolsVersionPath := util.GetHomePath(util.RMKDir, util.RMKToolsDir, util.ToolsVersionDir)
	toolsTmpPath := util.GetHomePath(util.RMKDir, util.RMKToolsDir, util.ToolsTmpDir)
	toolsBinPath := util.GetHomePath(".local", util.ToolsBinDir)

	if err := os.MkdirAll(toolsVersionPath, 0755); err != nil {
		return err
	}

	// Cleaning previously downloaded artifacts state
	for _, pkg := range conf.Tools {
		pkg.Artifacts = []string{}
	}

	for key, val := range conf.Tools {
		version, _ := semver.NewVersion(val.Version)
		spec.PkgUrl = val.Url
		spec.PkgName = val.Name + "-" + version.String()
		if !util.IsExists(filepath.Join(toolsVersionPath, spec.PkgName), true) {
			err := spec.download(silent)
			if err != nil {
				return err
			}

			if err := overwriteFiles(toolsVersionPath, val.Name+"-*", spec.PkgName); err != nil {
				return err
			}

			if val.Rename {
				conf.Tools[key].Artifacts, err = match(toolsTmpPath,
					[]string{filepath.Base(val.Url), val.Name + "-*", val.Name + "_*"})
				if err != nil {
					return err
				}
			} else {
				conf.Tools[key].Artifacts, err = match(toolsTmpPath,
					[]string{val.Name, val.Name + "-*", val.Name + "_*"})
				if err != nil {
					return err
				}
			}
		} else {
			continue
		}
	}

	if err := os.MkdirAll(toolsBinPath, 0755); err != nil {
		return err
	}

	for _, pkg := range conf.Tools {
		for _, pathArt := range pkg.Artifacts {
			if pkg.Rename {
				if err := util.CopyFile(pathArt, filepath.Join(toolsBinPath, pkg.Name)); err != nil {
					return err
				}
			} else {
				if err := util.CopyFile(pathArt, filepath.Join(toolsBinPath, filepath.Base(pathArt))); err != nil {
					return err
				}
			}
		}
	}

	return os.RemoveAll(toolsTmpPath)
}
