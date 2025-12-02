package help

import (
	"bytes"
	"fmt"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/spf13/cobra"
)

const faws_description = "Faws is an experimental version control system (VCS) built for game devs"

var HelpCmd = cobra.Command{
	Use:   "faws help",
	Short: "Faws is a toy version control system (VCS).",
	Run:   run_help_cmd,
}

func run_help_cmd(cmd *cobra.Command, args []string) {
	var m bytes.Buffer
	fmt.Fprintln(&m, faws_description)
	fmt.Fprintln(&m)

	max_command_length := int(0)

	// get max command length
	for _, category := range helpinfo.Categories {
		for _, command := range category.Commands {
			if len(command) > max_command_length {
				max_command_length = len(command)
			}
		}
	}

	for _, category := range helpinfo.Categories {
		fmt.Fprintln(&m, category.Description)

		minimum_width := max_command_length + 2

		for _, command := range category.Commands {
			fmt.Fprint(&m, "  ")
			command_length := len(command)
			fmt.Fprint(&m, command)
			for i := 0; i < (minimum_width - command_length); i++ {
				fmt.Fprint(&m, " ")
			}
			fmt.Fprint(&m, helpinfo.Text[command])
			fmt.Fprintln(&m)
		}

		fmt.Fprintln(&m)
	}

	app.Info(m.String())
}
