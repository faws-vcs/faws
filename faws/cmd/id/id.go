package id

import (
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

var IdentityCmd = cobra.Command{
	Use:     "id",
	GroupID: "id",
}

func init() {
	root.RootCmd.AddCommand(&IdentityCmd)
}
