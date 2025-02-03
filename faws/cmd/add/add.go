package add

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var add_cmd = cobra.Command{
	Use:     "add path origin",
	Short:   helpinfo.Text["add"],
	GroupID: "repo",
	Run:     run_add_cmd,
}

func init() {
	root.RootCmd.AddCommand(&add_cmd)
}

func run_add_cmd(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Help()
		os.Exit(1)
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.AddFileParams{
		Directory: working_directory,
		Path:      args[0],
		Origin:    args[1],
	}
	repository.AddFile(&params)
}
