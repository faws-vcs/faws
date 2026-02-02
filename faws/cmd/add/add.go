package add

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/spf13/cobra"
)

var add_cmd = cobra.Command{
	Use:     "add [--object] destination source",
	Short:   helpinfo.Text["add"],
	GroupID: "repo",
	Run:     run_add_cmd,
}

func init() {
	flag := add_cmd.Flags()
	flag.StringP("mode", "m", "", "file mode override")
	// TODO: I want to add a flag that matches each filename added as a pathspec
	// flag.StringP("pathspec", "p", "", "pathspec")
	flag.BoolP("lazy", "l", false, "refrain from chunking large files which share essential details with previously added files. Use this carefully, as it can introduce inconsistent information into your repository")
	flag.BoolP("verbose", "v", false, "display each file that gets cached")
	flag.BoolP("ref", "n", false, "if true, the source is a ref to a repository object, rather than a file name")
	root.RootCmd.AddCommand(&add_cmd)
}

func run_add_cmd(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Help()
		os.Exit(1)
	}

	flag := cmd.Flags()

	mode, err := flag.GetString("mode")
	if err != nil {
		app.Fatal(err)
	}

	lazy, err := flag.GetBool("lazy")
	if err != nil {
		app.Fatal(err)
	}

	verbose, err := flag.GetBool("verbose")
	if err != nil {
		app.Fatal(err)
	}

	source_is_ref, err := flag.GetBool("ref")
	if err != nil {
		app.Fatal(err)
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
	}

	var params = repository.AddFileParams{
		Directory:   working_directory,
		Destination: args[0],
		Source:      args[1],
		SourceIsRef: source_is_ref,
		AddLazy:     lazy,
		Verbose:     verbose,
	}

	if mode != "" {
		params.SetMode = true
		params.Mode, err = revision.ParseFileMode(mode)
		if err != nil {
			app.Fatal(err)
		}
	}

	repository.AddFile(&params)
}
