package commit_tree

import (
	"os"
	"time"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/repository"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/faws-vcs/faws/faws/timestamp"
	"github.com/spf13/cobra"
)

var commit_tree_cmd = cobra.Command{
	Use:     "commit-tree tree-hash",
	Short:   helpinfo.Text["commit-tree"],
	GroupID: "repo",
	Run:     run_commit_tree_cmd,
}

func init() {
	flags := commit_tree_cmd.Flags()
	flags.StringP("tag", "t", "", "a tag is required to make a commit")
	flags.StringP("parent", "p", "", "optionally, you can base this commit off a parent commit")
	flags.StringP("sign", "s", "", "specify a signing identity other than your current primary")
	flags.StringP("tree-date", "d", "", "specify the date of the tree object either in UNIX or DD.MM.YYYY format")
	flags.StringP("commit-date", "c", "", "specify the date of the commit object either in UNIX or DD.MM.YYYY format")
	root.RootCmd.AddCommand(&commit_tree_cmd)
}

func run_commit_tree_cmd(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		cmd.Help()
		os.Exit(1)
	}

	now := time.Now().Unix()

	// use working directory as default repository location
	working_directory, err := os.Getwd()
	if err != nil {
		app.Fatal(err)
		return
	}

	flags := cmd.Flags()

	tag, err := flags.GetString("tag")
	if err != nil {
		app.Fatal(err)
	}
	parent_commit, err := flags.GetString("parent")
	if err != nil {
		app.Fatal(err)
	}

	signing_identity, err := flags.GetString("sign")
	if err != nil {
		app.Fatal(err)
	}

	tree_date, err := flags.GetString("tree-date")
	if err != nil {
		app.Fatal(err)
	}
	commit_date, err := flags.GetString("commit-date")
	if err != nil {
		app.Fatal(err)
	}

	var params repository.CommitTreeParams
	params.Directory = working_directory
	params.TreeDate = now
	params.CommitDate = now
	params.Tag = tag
	params.Parent = parent_commit
	params.Sign = signing_identity
	params.Tree = args[0]
	if tree_date != "" {
		params.TreeDate, err = timestamp.Parse(tree_date)
		if err != nil {
			app.Fatal(err)
		}
	}
	if commit_date != "" {
		params.CommitDate, err = timestamp.Parse(commit_date)
		if err != nil {
			app.Fatal(err)
		}
	}

	repository.CommitTree(&params)
}
