package reset

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var reset_cmd = cobra.Command{
	Use:               "reset [commit-ref]",
	Short:             helpinfo.Text["reset"],
	GroupID:           "repo",
	Run:               run_reset_cmd,
	ValidArgsFunction: repository.InferenceRefArg(0),
}

func init() {
	root.RootCmd.AddCommand(&reset_cmd)
}

func run_reset_cmd(cmd *cobra.Command, args []string) {
	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.ResetParams{
		Directory: working_directory,
	}
	if len(args) > 0 {
		params.Ref = args[0]
	}
	repository.Reset(&params)
}
