package util

import (
	"fmt"

	goyaml "github.com/goccy/go-yaml"
	"github.com/urfave/cli/v2"
)

func ValidateNArg(c *cli.Context, expectedNArg int) error {
	if c.NArg() != expectedNArg {
		return fmt.Errorf("exactly %d argument(s) required for '%s' command", expectedNArg, c.Command.Name)
	}

	return nil
}

func YamlValidate(filePath string, raw []byte, out any) error {
	if err := goyaml.UnmarshalWithOptions(raw, out, goyaml.Strict()); err != nil {
		return fmt.Errorf("file %s, line %w", YAMLRelativePath(filePath), err)
	}

	return nil
}
