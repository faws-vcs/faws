package init

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var clone_cmd = cobra.Command{
	Use:     "clone remote [directory]",
	Short:   helpinfo.Text["clone"],
	GroupID: "remote",
	Run:     run_clone_cmd,
}

func init() {
	root.RootCmd.AddCommand(&clone_cmd)
}

func run_clone_cmd(cmd *cobra.Command, args []string) {
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
	var params = repository.CloneParams{
		Directory: working_directory,
		Remote:    args[0],
	}

	// use the second argument as repository location, if supplied
	if len(args) > 1 {
		params.Directory = args[1]
	}

	repository.Clone(&params)
}
