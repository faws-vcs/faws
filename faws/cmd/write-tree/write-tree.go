package write_tree

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var write_tree_cmd = cobra.Command{
	Use:     "write-tree",
	Short:   helpinfo.Text["write-tree"],
	GroupID: "repo",
	Run:     run_write_tree_cmd,
}

func init() {
	root.RootCmd.AddCommand(&write_tree_cmd)
}

func run_write_tree_cmd(cmd *cobra.Command, args []string) {
	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.WriteTreeParams{
		Directory: working_directory,
	}
	repository.WriteTree(&params)
}
