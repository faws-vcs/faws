package log

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var log_cmd = cobra.Command{
	Use:     "log commit",
	Short:   helpinfo.Text["log"],
	GroupID: "repo",
	Run:     run_log_cmd,
}

func init() {
	root.RootCmd.AddCommand(&log_cmd)
}

func run_log_cmd(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Help()
		os.Exit(1)
	}

	app.OpenConfiguration()

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.ViewLogParams{
		Directory: working_directory,
		Ref:       args[0],
	}
	repository.ViewLog(&params)
}
