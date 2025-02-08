package add

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

var add_cmd = cobra.Command{
	Use:     "add path origin",
	Short:   helpinfo.Text["add"],
	GroupID: "repo",
	Run:     run_add_cmd,
}

func init() {
	flag := add_cmd.Flags()
	flag.StringP("mode", "m", "", "file mode override")
	root.RootCmd.AddCommand(&add_cmd)
}

func run_add_cmd(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		cmd.Help()
		os.Exit(1)
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		app.Fatal(err)
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
	}

	var params = repository.AddFileParams{
		Directory: working_directory,
		Path:      args[0],
		Origin:    args[1],
	}

	if mode != "" {
		m, err := strconv.ParseUint(mode, 10, 8)
		if err != nil {
			app.Fatal(err)
		}
		params.SetMode = true
		params.Mode = revision.FileMode(m)
	}

	repository.AddFile(&params)
}
