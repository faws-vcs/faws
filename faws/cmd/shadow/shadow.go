package shadow

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var shadow_cmd = cobra.Command{
	Use:     "shadow remote ref",
	Short:   helpinfo.Text["shadow"],
	GroupID: "remote",
	Run:     run_shadow_cmd,
}

func init() {
	root.RootCmd.AddCommand(&shadow_cmd)
}

func run_shadow_cmd(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
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

	// initialize the repository
	var params = repository.ShadowParams{
		Directory: working_directory,
		Remote:    args[0],
		Ref:       args[1],
	}
	repository.Shadow(&params)
}
