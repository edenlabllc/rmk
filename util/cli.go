package util

import (
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/urfave/cli/v2"
	"github.com/urfave/cli/v2/altsrc"
)

func cliArgContains(flagName string) bool {
	for _, name := range strings.Split(flagName, ",") {
		name = strings.TrimSpace(name)
		count := utf8.RuneCountInString(name)
		if count > 2 {
			count = 2
		}

		flag := fmt.Sprintf("%s%s", strings.Repeat("-", count), name)

		for _, a := range os.Args {
			if a == flag {
				return true
			}
		}
	}

	return false
}

func printFlagSuggestions(lastArg string, flags []cli.Flag, writer io.Writer) {
	cur := strings.TrimPrefix(lastArg, "-")
	cur = strings.TrimPrefix(cur, "-")
	for _, flag := range flags {
		// skip hidden flags for bool type
		if boolFlag, ok := flag.(*cli.BoolFlag); ok && boolFlag.Hidden {
			continue
		}
		// skip hidden flags for altsrc bool type
		if altsrcBoolFlag, ok := flag.(*altsrc.BoolFlag); ok && altsrcBoolFlag.Hidden {
			continue
		}
		// skip hidden flags for string type
		if stringFlag, ok := flag.(*cli.StringFlag); ok && stringFlag.Hidden {
			continue
		}
		// skip hidden flags for altsrc string type
		if altsrcStringFlag, ok := flag.(*altsrc.StringFlag); ok && altsrcStringFlag.Hidden {
			continue
		}

		for _, name := range flag.Names() {
			name = strings.TrimSpace(name)
			// this will get total count utf8 letters in flag name
			count := utf8.RuneCountInString(name)
			if count > 2 {
				count = 2 // reuse this count to generate single - or -- in flag completion
			}
			// if flag name has more than one utf8 letter and last argument in cli has -- prefix then
			// skip flag completion for short flags example -v or -x
			if strings.HasPrefix(lastArg, "--") && count == 1 {
				continue
			}
			// match if last argument matches this flag and it is not repeated
			if strings.HasPrefix(name, cur) && cur != name && !cliArgContains(name) {
				flagCompletion := fmt.Sprintf("%s%s", strings.Repeat("-", count), name)
				_, _ = fmt.Fprintln(writer, flagCompletion)
			}
		}
	}
}

func ShellCompleteCustomOutput(c *cli.Context) {
	if len(os.Args) > 2 {
		if os.Args[len(os.Args)-2] != "" && strings.HasPrefix(os.Args[len(os.Args)-2], "-") {
			printFlagSuggestions(os.Args[len(os.Args)-2], c.Command.Flags, c.App.Writer)

			return
		}
	}
}
