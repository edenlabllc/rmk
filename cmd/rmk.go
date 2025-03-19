package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"rmk/providers/aws_provider"
	"rmk/util"
)

type RMKArtifactMetadata struct {
	ProjectName string    `json:"project_name"`
	Tag         string    `json:"tag"`
	PreviousTag string    `json:"previous_tag"`
	Version     string    `json:"version"`
	Commit      string    `json:"commit"`
	Date        time.Time `json:"date"`
	Runtime     struct {
		Goos   string `json:"goos"`
		Goarch string `json:"goarch"`
	} `json:"runtime"`
}

func completionAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
			return err
		}

		fmt.Println(util.CompletionZshScript)

		return nil
	}
}

func docGenerateAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
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

func getRMKArtifactMetadata(keyPath string) (*RMKArtifactMetadata, error) {
	rmkArtifactMetadata := &RMKArtifactMetadata{}
	aws := &aws_provider.AwsConfigure{Region: util.RMKBucketRegion}
	data, err := aws.GetAWSBucketFileData(util.RMKBucketName, util.RMKBin+"/"+keyPath+"/metadata.json")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &rmkArtifactMetadata); err != nil {
		return nil, err
	}

	return rmkArtifactMetadata, nil
}

func rmkURLFormation(paths ...string) string {
	u, err := url.Parse("https://" + util.RMKBucketName + ".s3." + util.RMKBucketRegion + ".amazonaws.com")
	if err != nil {
		zap.S().Fatal(err)
	}

	p := append([]string{u.Path}, paths...)
	u.Path = path.Join(p...)
	return u.String()
}

func updateRMK(pkgName, version string, silent, progressBar bool, ctx *cli.Context) error {
	zap.S().Infof("starting package download: %s", pkgName)
	pkgDst := util.GetHomePath(util.ToolsLocalDir, util.ToolsBinDir)
	if err := util.DownloadArtifact(
		rmkURLFormation(util.RMKBin, version, pkgName),
		pkgDst,
		pkgName,
		&http.Header{},
		silent,
		progressBar,
		ctx.Context,
	); err != nil {
		return err
	}

	if err := os.Rename(filepath.Join(pkgDst, pkgName), filepath.Join(pkgDst, util.RMKBin)); err != nil {
		return err
	}

	if err := os.Chmod(filepath.Join(pkgDst, util.RMKBin), 0755); err != nil {
		return err
	}

	relPath := strings.ReplaceAll(util.RMKSymLinkPath, filepath.Base(util.RMKSymLinkPath), "")
	if syscall.Access(relPath, uint32(2)) == nil {
		if !util.IsExists(util.RMKSymLinkPath, true) {
			return os.Symlink(filepath.Join(pkgDst, util.RMKBin), util.RMKSymLinkPath)
		}
	} else {
		zap.S().Warnf("symlink was not created automatically due to permissions, "+
			"please complete installation by running command: \n"+
			"sudo ln -s %s %s", filepath.Join(pkgDst, util.RMKBin), util.RMKSymLinkPath)
	}

	return nil
}

func updateAction() cli.ActionFunc {
	return func(c *cli.Context) error {
		if err := util.ValidateNArg(c, 0); err != nil {
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
