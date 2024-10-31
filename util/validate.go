package util

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

func ValidateNArg(c *cli.Context, expectedNArg int) error {
	if c.NArg() != expectedNArg {
		return fmt.Errorf("exactly %d argument(s) required for '%s' command", expectedNArg, c.Command.Name)
	}

	return nil
}
