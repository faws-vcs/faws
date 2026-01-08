package tracker

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/p2p"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var tracker_cmd = cobra.Command{
	Use:     "tracker",
	Short:   helpinfo.Text["tracker"],
	GroupID: "remote",
	Run:     run_tracker_cmd,
}

func init() {
	root.RootCmd.AddCommand(&tracker_cmd)
}

func run_tracker_cmd(cmd *cobra.Command, args []string) {
	var err error
	var working_directory string
	// use working directory as default tracker server
	working_directory, err = os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = p2p.RunTrackerServerParams{
		Directory: working_directory,
	}
	p2p.RunTrackerServer(&params)
}
