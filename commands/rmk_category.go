package commands

import (
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"rmk/system"
)

func completionAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		fmt.Println(system.CompletionZshScript)

		return nil
	}
}

func docGenerateAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		man, err := c.App.ToMarkdown()
		if err != nil {
			return nil
		}

		fmt.Println(man)

		return nil
	}
}

func updateAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := system.ValidateNArg(c, 0); err != nil {
			return err
		}

		var version string
		latestPath := "latest"

		if c.Bool("release-candidate") {
			latestPath = "latest-rc"
		}

		metadata, err := getRMKArtifactMetadata(latestPath)
		if err != nil {
			return err
		}

		if len(c.String("version")) > 0 {
			v, err := semver.NewVersion(c.String("version"))
			if err != nil {
				return err
			}

			version = v.Original()
		}

		verCurrent, _ := semver.NewVersion(fmt.Sprintf("%v", c.App.Metadata["version"]))
		verFound, _ := semver.NewVersion(metadata.Version)
		binaryName := fmt.Sprintf("%s", c.App.Metadata["binaryName"])
		if verCurrent.LessThan(verFound) && len(version) == 0 {
			zap.S().Infof("newer release version RMK available: %s", verFound.Original())
			if err := updateRMK(binaryName, latestPath, false, true, c); err != nil {
				return err
			}
		} else if len(version) > 0 {
			zap.S().Infof("update current RMK version from %s to %s",
				c.App.Metadata["version"], version)
			if err := updateRMK(binaryName, version, false, true, c); err != nil {
				return err
			}
		} else {
			zap.S().Infof("installed RMK version %s is up-to-date", verCurrent.Original())
		}

		return nil
	}
}
