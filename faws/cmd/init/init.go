package init

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var init_cmd = cobra.Command{
	Use:     "init [remote]",
	Short:   helpinfo.Text["init"],
	GroupID: "repo",
	Run:     run_init_cmd,
}

func init() {
	root.RootCmd.AddCommand(&init_cmd)
}

func run_init_cmd(cmd *cobra.Command, args []string) {
	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	// initialize the repository
	var params = repository.InitParams{
		Directory: working_directory,
	}
	if len(args) > 0 {
		params.Remote = args[0]
	}
	repository.Init(&params)
}
