package cmd

import (
	"fmt"
	"os"
	"path/filepath"
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
	credentials, err := k.Conf.AwsConfigure.GetECRCredentials(k.Conf.AWSECRRegion)
	if err != nil {
		return err
	}

	k.SpecCMD.Envs = append(k.SpecCMD.Envs, "K3D_NAME="+k.Conf.Name)

	if token, ok := credentials[k.Conf.AWSECRUserName]; !ok {
		return fmt.Errorf("failed to get ECR token")
	} else {
		k.SpecCMD.Envs = append(k.SpecCMD.Envs, "K3D_AWS_ECR_USER="+k.Conf.AWSECRUserName, "K3D_AWS_ECR_PASSWORD="+token)
	}

	if len(k.Ctx.String("k3d-volume-host-path")) > 0 {
		k.SpecCMD.Envs = append(k.SpecCMD.Envs, "K3D_VOLUME_HOST_PATH="+k.Ctx.String("k3d-volume-host-path"))
		return nil
	}

	k.SpecCMD.Envs = append(k.SpecCMD.Envs, "K3D_VOLUME_HOST_PATH="+util.GetPwdPath(""))

	return nil
}

func (k *K3DCommands) createDeleteK3DCluster() error {
	var k3dDst string

	k.K3DCluster = true
	if _, _, err := k.getKubeContext(); err != nil {
		return err
	}

	for name, pkg := range k.Conf.Clusters {
		if strings.HasPrefix(name, util.K3DConfigPrefix) {
			k3dDst = pkg.DstPath
			break
		}
	}

	if len(k3dDst) == 0 {
		return fmt.Errorf("cluster provider with name %s not found", util.K3DConfigPrefix)
	}

	match, err := util.WalkMatch(k3dDst, util.K3DConfigPrefix+".yaml")
	if err != nil {
		return err
	}

	if len(match) == 0 {
		return fmt.Errorf("configuration file for %s not found", util.K3DConfigPrefix)
	}

	if err := k.prepareK3D("cluster", k.Ctx.Command.Name, "--config", match[0]); err != nil {
		return err
	}

	// Creating specific dir for k3d registry configuration
	k3dRegistryHostPath := filepath.Join(filepath.Dir(match[0]), util.K3DConfigPrefix)
	k.SpecCMD.Envs = append(k.SpecCMD.Envs, "K3D_REGISTRY_HOST_PATH="+k3dRegistryHostPath)

	if err := os.RemoveAll(k3dRegistryHostPath); err != nil {
		return err
	}

	if err := os.MkdirAll(k3dRegistryHostPath, 0755); err != nil {
		return err
	}

	return runner(k).runCMD()
}

func (k *K3DCommands) importImageToK3DCluster() error {
	if err := k.prepareK3D(append(append([]string{}, "image", "import", "--cluster", k.Conf.Name, "--keep-tools"),
		k.Ctx.StringSlice("k3d-import-image")...)...); err != nil {
		return err
	}

	return runner(k).runCMD()
}

func (k *K3DCommands) listK3DClusters() error {
	k.K3DCluster = true
	if _, _, err := k.getKubeContext(); err != nil {
		return err
	}

	if err := k.prepareK3D("cluster", k.Ctx.Command.Name); err != nil {
		return err
	}

	return runner(k).runCMD()
}

func (k *K3DCommands) startStopK3DCluster() error {
	k.K3DCluster = true
	if _, _, err := k.getKubeContext(); err != nil {
		return err
	}

	if err := k.prepareK3D("cluster", k.Ctx.Command.Name, k.Conf.Name); err != nil {
		return err
	}

	return runner(k).runCMD()
}

func K3DCreateAction(conf *config.Config) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		if err := resolveDependencies(conf.InitConfig(false), c, false); err != nil {
			return err
		}

		return newK3DCommands(conf, c, util.GetPwdPath("")).createDeleteK3DCluster()
	}
}

func K3DAction(conf *config.Config, action func(k3dRunner K3DRunner) error) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		return action(newK3DCommands(conf, c, util.GetPwdPath("")))
	}
}
