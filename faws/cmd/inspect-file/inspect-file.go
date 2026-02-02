package inspect_file

import (
	"github.com/faws-vcs/faws/faws/app/about"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var inspect_file_cmd = cobra.Command{
	Use:     "inspect-file name",
	Short:   helpinfo.Text["inspect-file"],
	GroupID: "about",
	Run:     run_inspect_file_cmd,
}

func init() {
	root.RootCmd.AddCommand(&inspect_file_cmd)
}

func run_inspect_file_cmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help()
		return
	}

	var params about.InspectFileParams
	params.Name = args[0]
	about.InspectFile(&params)
}
