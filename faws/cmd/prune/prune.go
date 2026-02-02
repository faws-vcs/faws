package prune

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var prune_cmd = cobra.Command{
	Use:     "prune",
	Short:   helpinfo.Text["prune"],
	GroupID: "repo",
	Run:     run_prune_cmd,
}

func init() {
	// flags := prune_cmd.Flags()
	root.RootCmd.AddCommand(&prune_cmd)
}

func run_prune_cmd(cmd *cobra.Command, args []string) {
	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params repository.PruneCacheParams
	params.Directory = working_directory
	repository.PruneCache(&params)
}
