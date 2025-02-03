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
	Use:     "rm [-r] path",
	Short:   helpinfo.Text["rm"],
	GroupID: "repo",
	Run:     run_rm_cmd,
}

func init() {
	flags := rm_cmd.Flags()
	flags.BoolP("recurse", "r", false, "recursively remove non-empty directories with all contents")
	root.RootCmd.AddCommand(&rm_cmd)
}

func run_rm_cmd(cmd *cobra.Command, args []string) {
	app.OpenConfiguration()

	recurse, err := cmd.Flags().GetBool("recurse")
	if err != nil {
		app.Fatal(err)
		return
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.RemoveFileParams{
		Directory: working_directory,
		Recurse:   recurse,
	}
	if len(args) != 0 {
		params.Path = args[0]
	}
	repository.RemoveFile(&params)
}
