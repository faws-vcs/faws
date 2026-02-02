package add

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/spf13/cobra"
)

var chmod_cmd = cobra.Command{
	Use:     "chmod pathspec mode",
	Short:   helpinfo.Text["chmod"],
	GroupID: "repo",
	Run:     run_chmod_cmd,
}

func init() {
	root.RootCmd.AddCommand(&chmod_cmd)
}

func run_chmod_cmd(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Help()
		os.Exit(1)
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
	}

	var params repository.ChmodParams
	params.Directory = working_directory
	params.Path = args[0]
	params.Mode, err = revision.ParseFileMode(args[1])
	if err != nil {
		app.Fatal(err)
	}

	repository.Chmod(&params)
}
