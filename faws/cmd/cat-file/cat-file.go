package cat_file

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var cat_file_cmd = cobra.Command{
	Use:     "cat-file [type] object",
	Short:   helpinfo.Text["cat-file"],
	GroupID: "repo",
	Run:     run_cat_file_cmd,
}

func init() {
	flags := cat_file_cmd.Flags()
	flags.BoolP("pretty-print", "p", false, "pretty-print the contents of object based on its type.")
	root.RootCmd.AddCommand(&cat_file_cmd)
}

func run_cat_file_cmd(cmd *cobra.Command, args []string) {
	if len(args) != 2 && len(args) != 1 {
		cmd.Help()
		return
	}

	app.OpenConfiguration()

	object := args[len(args)-1]

	object_type := ""

	if len(args) == 2 {
		object_type = args[0]
	}

	pretty_print, err := cmd.Flags().GetBool("pretty-print")
	if err != nil {
		panic(err)
	}

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.CatFileParams{
		Directory:   working_directory,
		Prefix:      object_type,
		Ref:         object,
		PrettyPrint: pretty_print,
	}
	repository.CatFile(&params)
}
