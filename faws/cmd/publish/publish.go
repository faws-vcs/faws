package publish

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var publish_cmd = cobra.Command{
	Use:     "publish",
	Short:   helpinfo.Text["publish"],
	GroupID: "remote",
	Run:     run_publish_cmd,
}

func init() {
	flag := publish_cmd.Flags()
	flag.StringP("sign", "s", "", "specify a signing identity other than your current primary")
	root.RootCmd.AddCommand(&publish_cmd)
}

func run_publish_cmd(cmd *cobra.Command, args []string) {
	var err error
	var working_directory string
	// use working directory as default repository location
	working_directory, err = os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	// publish from the repository
	var params = repository.PublishParams{
		Directory: working_directory,
	}
	flag := cmd.Flags()
	params.TrackerURL = os.Getenv("FAWS_TRACKER")
	params.Sign, err = flag.GetString("sign")
	if err != nil {
		app.Fatal(err)
		return
	}
	repository.Publish(&params)
}
