package fsck

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var fsck_cmd = cobra.Command{
	Use:     "fsck [ref]",
	Short:   helpinfo.Text["fsck"],
	GroupID: "repo",
	Run:     run_fsck_cmd,
}

func init() {
	flags := fsck_cmd.Flags()
	flags.BoolP("destructive", "d", false, "remove corrupted objects from the repository (cannot be undone!)")
	root.RootCmd.AddCommand(&fsck_cmd)
}

func run_fsck_cmd(cmd *cobra.Command, args []string) {
	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	flags := cmd.Flags()
	destructive, err := flags.GetBool("destructive")
	if err != nil {
		app.Fatal(err)
		return
	}

	var params repository.CheckObjectsParams
	params.Directory = working_directory
	if len(args) > 0 {
		params.Ref = args[0]
	}
	params.Destructive = destructive
	repository.CheckObjects(&params)
}
