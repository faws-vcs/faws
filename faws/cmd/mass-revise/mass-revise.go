package rewrite_commits

import (
	"encoding/hex"
	"os"
	"regexp"

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
	flag.String("match-tag", "", "only revise certain tags matching this regex")
	flag.String("match-file-magic", "", "only revise files beginning with these bytes")
	root.RootCmd.AddCommand(&mass_revise_cmd)
}

func run_mass_revise_cmd(cmd *cobra.Command, args []string) {
	flag := cmd.Flags()
	m, err := flag.GetString("mode")
	if err != nil {
		app.Warning(err)
		cmd.Help()
		os.Exit(1)
	}

	destroy, err := flag.GetBool("destroy")
	if err != nil {
		app.Warning(err)
		cmd.Help()
		os.Exit(1)
	}

	match_file_magic, err := flag.GetString("match-file-magic")
	if err != nil {
		app.Warning(err)
		cmd.Help()
		os.Exit(1)
	}

	match_tag, err := flag.GetString("match-tag")
	if err != nil {
		app.Warning(err)
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
		params.SetFileMode = true
		params.NewFileMode, err = revision.ParseFileMode(m)
		if err != nil {
			app.Fatal(err)
		}
	}

	if match_file_magic != "" {
		params.MatchFileMagic, err = hex.DecodeString(match_file_magic)
		if err != nil {
			app.Fatal(err)
		}
	}

	if match_tag != "" {
		params.MatchTag, err = regexp.CompilePOSIX(match_tag)
		if err != nil {
			app.Fatal(err)
		}
	}

	repository.MassRevise(&params)
}
