package ls_tag

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var tag_cmd = cobra.Command{
	Use:     "tag",
	Short:   helpinfo.Text["tag"],
	GroupID: "repo",
	Run:     run_tag_cmd,
}

func init() {
	root.RootCmd.AddCommand(&tag_cmd)
}

func run_tag_cmd(cmd *cobra.Command, args []string) {
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
