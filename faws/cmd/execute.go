package cmd

import (
	"os"

	_ "github.com/faws-vcs/faws/faws/cmd/id/create"
	_ "github.com/faws-vcs/faws/faws/cmd/id/ls"
	_ "github.com/faws-vcs/faws/faws/cmd/id/primary"
	_ "github.com/faws-vcs/faws/faws/cmd/id/rm"
	_ "github.com/faws-vcs/faws/faws/cmd/id/set"

	_ "github.com/faws-vcs/faws/faws/cmd/add"
	_ "github.com/faws-vcs/faws/faws/cmd/cat-file"
	_ "github.com/faws-vcs/faws/faws/cmd/checkout"
	_ "github.com/faws-vcs/faws/faws/cmd/commit"
	_ "github.com/faws-vcs/faws/faws/cmd/commit-tree"
	_ "github.com/faws-vcs/faws/faws/cmd/init"
	_ "github.com/faws-vcs/faws/faws/cmd/log"
	_ "github.com/faws-vcs/faws/faws/cmd/ls-tag"
	_ "github.com/faws-vcs/faws/faws/cmd/ls-tree"
	_ "github.com/faws-vcs/faws/faws/cmd/mass-revise"
	_ "github.com/faws-vcs/faws/faws/cmd/pull"
	_ "github.com/faws-vcs/faws/faws/cmd/rm"
	_ "github.com/faws-vcs/faws/faws/cmd/shadow"
	_ "github.com/faws-vcs/faws/faws/cmd/status"
	_ "github.com/faws-vcs/faws/faws/cmd/write-tree"

	"github.com/faws-vcs/faws/faws/cmd/root"
)

func Execute() {
	err := root.RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}
