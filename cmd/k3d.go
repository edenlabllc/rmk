package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/urfave/cli/v2"

	"rmk/config"
	"rmk/util"
)

type K3DRunner interface {
	createDeleteK3DCluster() error
	importImageToK3DCluster() error
	listK3DClusters() error
	startStopK3DCluster() error
}

type K3DCommands struct {
	*ReleaseCommands
}

func newK3DCommands(conf *config.Config, ctx *cli.Context, workDir string) *K3DCommands {
	return &K3DCommands{&ReleaseCommands{Conf: conf, Ctx: ctx, WorkDir: workDir}}
}

func (k *K3DCommands) k3d(args ...string) *util.SpecCMD {
	return &util.SpecCMD{
		Args:          append([]string{}, args...),
		Command:       "k3d",
		Dir:           k.WorkDir,
		Ctx:           k.Ctx.Context,
		DisableStdOut: false,
		Debug:         false,
	}
}

func (k *K3DCommands) prepareK3D(args ...string) error {
	k.SpecCMD = k.k3d(args...)
	k.SpecCMD.Debug = true

	switch {
	case k.APICluster:
		k.SpecCMD.Envs = append(k.SpecCMD.Envs, "K3D_NAME="+util.CAPI)
	case k.K3DCluster:
		k.SpecCMD.Envs = append(k.SpecCMD.Envs, "K3D_NAME="+k.Conf.Name)
	}

	if len(k.Ctx.String("k3d-volume-host-path")) > 0 {
		k.SpecCMD.Envs = append(k.SpecCMD.Envs, "K3D_VOLUME_HOST_PATH="+k.Ctx.String("k3d-volume-host-path"))
		return nil
	}

	k.SpecCMD.Envs = append(k.SpecCMD.Envs, "K3D_VOLUME_HOST_PATH="+util.GetPwdPath(""))

	return nil
}

func (k *K3DCommands) selectCluster() {
	switch k.Ctx.Command.Category {
	case util.CAPI:
		k.APICluster = true
	case util.K3DPrefix:
		k.K3DCluster = true
	}
}

func (k *K3DCommands) createDeleteK3DCluster() error {
	k.selectCluster()

	if _, _, err := clusterRunner(&ClusterCommands{k.ReleaseCommands}).getKubeContext(); err != nil {
		return err
	}

	k.SpecCMD = k.prepareHelmfile("--log-level", "error", "-l", "cluster="+k.Ctx.Command.Category, "template")
	k.SpecCMD.DisableStdOut = true
	if err := releaseRunner(k).runCMD(); err != nil {
		return fmt.Errorf("Helmfile failed to render template by label release: cluster=%s\n%s",
			k.Ctx.Command.Category, k.SpecCMD.StderrBuf.String())
	}

	k3dConfig, err := util.CreateTempYAMLFile(os.TempDir(), k.Ctx.Command.Category+"-config", k.SpecCMD.StdoutBuf.Bytes())
	if err != nil {
		return err
	}

	if err := k.prepareK3D("cluster", k.Ctx.Command.Name, "--config", k3dConfig); err != nil {
		return err
	}

	if err := releaseRunner(k).runCMD(); err != nil {
		if err := os.RemoveAll(k3dConfig); err != nil {
			return err
		}

		return err
	}

	return os.RemoveAll(k3dConfig)
}

func (k *K3DCommands) importImageToK3DCluster() error {
	if err := k.prepareK3D(append(append([]string{}, "image", "import", "--cluster", k.Conf.Name, "--keep-tools"),
		k.Ctx.StringSlice("k3d-import-image")...)...); err != nil {
		return err
	}

	return releaseRunner(k).runCMD()
}

func (k *K3DCommands) listK3DClusters() error {
	k.selectCluster()

	if _, _, err := clusterRunner(&ClusterCommands{k.ReleaseCommands}).getKubeContext(); err != nil {
		return err
	}

	if k.APICluster {
		if err := k.prepareK3D("cluster", k.Ctx.Command.Name, util.CAPI); err != nil {
			return err
		}

		k.SpecCMD.DisableStdOut = true
		if err := releaseRunner(k).runCMD(); err != nil {
			if strings.Contains(k.SpecCMD.StderrBuf.String(), "No nodes found for given cluster") {
				return fmt.Errorf("%s management cluster not found", strings.ToUpper(util.CAPI))
			} else {
				return fmt.Errorf("%s", k.SpecCMD.StderrBuf.String())
			}
		}

		fmt.Printf("%s", k.SpecCMD.StdoutBuf.String())

		return nil
	}

	if err := k.prepareK3D("cluster", k.Ctx.Command.Name); err != nil {
		return err
	}

	return releaseRunner(k).runCMD()
}

func (k *K3DCommands) startStopK3DCluster() error {
	k.selectCluster()

	if _, _, err := clusterRunner(&ClusterCommands{k.ReleaseCommands}).getKubeContext(); err != nil {
		return err
	}

	if err := k.prepareK3D("cluster", k.Ctx.Command.Name, k.Conf.Name); err != nil {
		return err
	}

	return releaseRunner(k).runCMD()
}

func K3DCreateAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(), c, false); err != nil {
			return err
		}

		return newK3DCommands(conf, c, util.GetPwdPath("")).createDeleteK3DCluster()
	}
}

func K3DAction(conf *config.Config, action func(k3dRunner K3DRunner) error) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		return action(newK3DCommands(conf, c, util.GetPwdPath("")))
	}
}
