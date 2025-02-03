package root

import (
	"github.com/faws-vcs/faws/faws/cmd/help"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/spf13/cobra"
)

var RootCmd = cobra.Command{
	Use:   "faws",
	Short: "Faws is a toy version control system (VCS).",

	Run: run_root_cmd,
}

func init() {
	for _, category := range helpinfo.Categories {
		RootCmd.AddGroup(&cobra.Group{
			ID:    category.CategoryID,
			Title: category.Description,
		})
	}
}

func run_root_cmd(cmd *cobra.Command, args []string) {
	help.HelpCmd.Execute()
}
