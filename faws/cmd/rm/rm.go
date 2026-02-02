package rm

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var rm_cmd = cobra.Command{
	Use:     "rm pathspec",
	Short:   helpinfo.Text["rm"],
	GroupID: "repo",
	Run:     run_rm_cmd,
}

func init() {
	root.RootCmd.AddCommand(&rm_cmd)
}

func run_rm_cmd(cmd *cobra.Command, args []string) {
	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.RemoveFileParams{
		Directory: working_directory,
	}
	if len(args) != 0 {
		params.Pattern = args[0]
	}
	repository.RemoveFile(&params)
}
