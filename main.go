package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"rmk/commands"
	"rmk/logger"
)

var (
	builtBy   string
	commit    string
	date      string
	name      string
	target    string
	timestamp string
	version   string
)

// Override default help template parsing with add [GLOBAL OPTIONS] for all commands and subcommands
func init() {
	var output string

	for _, val := range commands.FlagsGlobal() {
		output += fmt.Sprintf("   %s\n", val.String())
	}

	cli.CommandHelpTemplate += `
GLOBAL OPTIONS:
` + output + `{{range $index, $option := .VisibleFlags}}{{if eq $option.Name "help"}}{{"   "}}{{$option}}{{end}}{{end}}
`

	cli.SubcommandHelpTemplate += `
GLOBAL OPTIONS:
` + output + `{{range $index, $option := .VisibleFlags}}{{if eq $option.Name "help"}}{{"   "}}{{$option}}{{end}}{{end}}
`
}

func runCLI() *cli.App {
	app := cli.NewApp()
	app.Name = name
	app.Description = "Command line tool for reduced management of the " +
		"provision of Kubernetes clusters in different environments and management of service releases." +
		fmt.Sprintf("\nBuiltBy: %s\nCommit: %s\nDate: %s\nTarget: %s",
			builtBy, commit, date, target)
	app.Usage = "Reduced management for Kubernetes"
	app.Version = version
	app.Metadata = map[string]interface{}{
		"binaryName": name + "_" + target,
		"timestamp":  timestamp,
		"version":    version,
	}

	app.Flags = commands.FlagsGlobal()
	app.Before = func(c *cli.Context) error {
		logger.Init(c.String("log-format"), c.String("log-level"))
		return nil
	}

	// Enable command auto-completion (the --generate-bash-completion flag is provided out of box)
	// Incompatible with UseShortOptionHandling option
	app.EnableBashCompletion = true
	// Enable flag and command suggestions
	app.Suggest = true

	app.Commands = commands.Commands()
	sort.Sort(cli.CommandsByName(app.Commands))

	return app
}

func main() {
	err := runCLI().Run(os.Args)
	if err != nil {
		zap.S().Fatal(err)
	}
}
