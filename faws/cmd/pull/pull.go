package init

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var pull_cmd = cobra.Command{
	Use:     "pull remote",
	Short:   helpinfo.Text["pull"],
	GroupID: "remote",
	Run:     run_pull_cmd,
}

func init() {
	root.RootCmd.AddCommand(&pull_cmd)
}

func run_pull_cmd(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		os.Exit(1)
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	// initialize the repository
	var params = repository.PullParams{
		Directory: working_directory,
		Remote:    args[0],
	}
	repository.Pull(&params)
}
