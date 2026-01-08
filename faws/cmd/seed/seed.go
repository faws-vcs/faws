package seed

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var seed_cmd = cobra.Command{
	Use:     "seed topic-url",
	Short:   helpinfo.Text["seed"],
	GroupID: "remote",
	Run:     run_seed_cmd,
}

func init() {
	flag := seed_cmd.Flags()
	flag.StringP("sign", "s", "", "use a signing identity to identify yourself with the P2P network")
	root.RootCmd.AddCommand(&seed_cmd)
}

func run_seed_cmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cmd.Help()
		os.Exit(1)
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.SeedParams{
		Directory: working_directory,
		TopicURI:  args[0],
	}
	flag := cmd.Flags()
	params.TrackerURL = os.Getenv("FAWS_TRACKER")
	params.Sign, err = flag.GetString("sign")
	if err != nil {
		app.Fatal(err)
		return
	}
	repository.Seed(&params)
}
