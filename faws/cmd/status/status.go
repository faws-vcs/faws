package status

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var status_cmd = cobra.Command{
	Use:     "status",
	Short:   helpinfo.Text["status"],
	GroupID: "repo",
	Run:     run_status_cmd,
}

func init() {
	flag := status_cmd.Flags()
	flag.BoolP("show-lazy", "l", false, "show signatures of lazy files in the index")
	root.RootCmd.AddCommand(&status_cmd)
}

func run_status_cmd(cmd *cobra.Command, args []string) {
	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	flag := cmd.Flags()
	// show the current repository status
	var params = repository.StatParams{
		Directory: working_directory,
	}
	params.ShowLazyFiles, err = flag.GetBool("show-lazy")
	if err != nil {
		return
	}

	repository.Stat(&params)
}
