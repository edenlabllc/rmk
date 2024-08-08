package util

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func ValidateGitHubToken(c *cli.Context, errorMsg string) error {
	if !c.IsSet("github-token") {
		if errorMsg == "" {
			return fmt.Errorf(ConfigNotInitializedErrorText)
		} else {
			return fmt.Errorf(errorMsg)
		}
	}

	return nil
}

func ValidateNArg(c *cli.Context, expectedNArg int) error {
	if c.NArg() != expectedNArg {
		return fmt.Errorf("exactly %d argument(s) required for '%s' command", expectedNArg, c.Command.Name)
	}

	return nil
}
