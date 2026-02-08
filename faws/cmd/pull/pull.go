package pull

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var pull_cmd = cobra.Command{
	Use:               "pull [--tag] ...",
	Short:             helpinfo.Text["pull"],
	GroupID:           "remote",
	Run:               run_pull_cmd,
	ValidArgsFunction: repository.InferenceRefLastArg,
}

func init() {
	flag := pull_cmd.Flags()
	flag.BoolP("tag", "t", false, "pull the named tags instead of objects. If no tags are named, all tags from the origin will get pulled")
	flag.BoolP("verbose", "v", false, "display extra information")
	flag.BoolP("quiet", "q", false, "shut up the interactive Hud")
	root.RootCmd.AddCommand(&pull_cmd)
}

func run_pull_cmd(cmd *cobra.Command, args []string) {
	var err error
	var working_directory string
	// use working directory as default repository location
	working_directory, err = os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	// pull from the repository
	var params = repository.PullParams{
		Directory: working_directory,
	}
	if len(args) > 0 {
		params.Ref = args
	}
	flag := cmd.Flags()
	params.Tags, err = flag.GetBool("tag")
	if err != nil {
		app.Fatal(err)
		return
	}

	if !params.Tags {
		if len(args) < 1 {
			cmd.Help()
			os.Exit(1)
		}
	}

	params.TrackerURL = os.Getenv("FAWS_TRACKER")
	params.Verbose, err = flag.GetBool("verbose")
	if err != nil {
		app.Fatal(err)
		return
	}
	params.Quiet, err = flag.GetBool("quiet")
	if err != nil {
		app.Fatal(err)
		return
	}
	repository.Pull(&params)
}
