package commands

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"rmk/config"
	"rmk/system"
)

type DockerRunner interface {
	dockerLogin() error
	dockerLogout() error
}

type CRCommands struct {
	*ReleaseCommands
}

func newCRCommands(conf *config.Config, ctx *cli.Context, workDir string) *CRCommands {
	return &CRCommands{&ReleaseCommands{Conf: conf, Ctx: ctx, WorkDir: workDir}}
}

func (cr *CRCommands) docker(args ...string) *system.SpecCMD {
	return &system.SpecCMD{
		Args:          append([]string{}, args...),
		Command:       "docker",
		Dir:           cr.WorkDir,
		Ctx:           cr.Ctx.Context,
		DisableStdOut: true,
		Debug:         false,
	}
}

func (cr *CRCommands) dockerLogin() error {
	credentials, err := cr.Conf.AwsConfigure.GetECRCredentials(cr.Conf.AWSECRRegion)
	if err != nil {
		return err
	}

	if token, ok := credentials[cr.Conf.AWSECRUserName]; !ok {
		return fmt.Errorf("failed to get ECR token")
	} else {
		if cr.Ctx.Bool("get-token") {
			fmt.Println(token)
			return nil
		}

		cr.SpecCMD = cr.docker("login", "--username", cr.Conf.AWSECRUserName, "--password", token,
			cr.Conf.AWSECRHost)
		if err := runner(cr).runCMD(); err != nil {
			return err
		}

		if !strings.Contains(cr.SpecCMD.StderrBuf.String(), "Using --password") {
			return fmt.Errorf(strings.ReplaceAll(cr.SpecCMD.StderrBuf.String(), "\n", ""))
		}

		zap.S().Info(strings.ReplaceAll(cr.SpecCMD.StdoutBuf.String(), "\n", ""))
	}

	return nil
}

func (cr *CRCommands) dockerLogout() error {
	cr.SpecCMD = cr.docker("logout", cr.Conf.AWSECRHost)
	if err := runner(cr).runCMD(); err != nil {
		return err
	}

	zap.S().Info(strings.ReplaceAll(cr.SpecCMD.StdoutBuf.String(), "\n", ""))

	return nil
}

func containerRegistryAction(conf *config.Config, action func(dockerRunner DockerRunner) error) cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateArtifactModeDefault(c, ""); err != nil {
			return err
		}

		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		return action(newCRCommands(conf, c, system.GetPwdPath("")))
	}
}
