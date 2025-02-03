package ls_tag

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var ls_tags_cmd = cobra.Command{
	Use:     "ls-tag",
	Short:   helpinfo.Text["ls-tag"],
	GroupID: "repo",
	Run:     run_ls_tags_cmd,
}

func init() {
	root.RootCmd.AddCommand(&ls_tags_cmd)
}

func run_ls_tags_cmd(cmd *cobra.Command, args []string) {
	app.OpenConfiguration()

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	var params = repository.ListTagsParams{
		Directory: working_directory,
	}
	repository.ListTags(&params)
}
