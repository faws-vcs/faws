package version

import (
	"github.com/faws-vcs/faws/faws/app/about"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/root"
	"github.com/spf13/cobra"
)

func init() {
	root.RootCmd.AddCommand(&version_cmd)
}

var version_cmd = cobra.Command{
	Use:     "version",
	Short:   helpinfo.Text["version"],
	GroupID: "about",
	Run:     run_version_cmd,
}

func run_version_cmd(cmd *cobra.Command, args []string) {
	var params about.VersionParams
	about.Version(&params)
}
