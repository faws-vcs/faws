package rewrite_commits

import (
	"os"
	"strconv"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/spf13/cobra"
)

var mass_revise_cmd = cobra.Command{
	Use:     "mass-revise",
	Short:   helpinfo.Text["mass-revise"],
	GroupID: "repo",
	Run:     run_mass_revise_cmd,
}

func init() {
	flag := mass_revise_cmd.Flags()
	flag.Bool("destroy", false, "destroy all current commits and replace with new ones")
	flag.StringP("mode", "m", "", "the new filemode")
	root.RootCmd.AddCommand(&mass_revise_cmd)
}

func run_mass_revise_cmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cmd.Help()
		os.Exit(1)
	}

	flag := cmd.Flags()
	m, err := flag.GetString("mode")
	if err != nil {
		cmd.Help()
		os.Exit(1)
	}

	destroy, err := cmd.Flags().GetBool("destroy")
	if err != nil {
		cmd.Help()
		os.Exit(1)
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
	}

	var params = repository.MassReviseParams{
		Directory:   working_directory,
		Destructive: destroy,
	}

	if m != "" {
		mode, err := strconv.ParseUint(m, 10, 8)
		if err != nil {
			app.Fatal(err)
		}
		params.SetFileMode = true
		params.NewFileMode = revision.FileMode(mode)
	}

	repository.MassRevise(&params)
}
