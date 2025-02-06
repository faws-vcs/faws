package ls_tree

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var ls_tree_cmd = cobra.Command{
	Use:     "ls-tree [-r] object",
	Short:   helpinfo.Text["ls-tree"],
	GroupID: "repo",
	Run:     run_ls_tree_cmd,
}

func init() {
	flags := ls_tree_cmd.Flags()
	flags.BoolP("recurse", "r", false, "recurse into subdirectories")
	root.RootCmd.AddCommand(&ls_tree_cmd)
}

func run_ls_tree_cmd(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		return
	}

	recurse, err := cmd.Flags().GetBool("recurse")
	if err != nil {
		panic(err)
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.ListTreeParams{
		Directory: working_directory,
		Ref:       args[0],
		Recurse:   recurse,
	}
	repository.ListTree(&params)
}
