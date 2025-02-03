package rm

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/identities"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/id"
	"github.com/spf13/cobra"
)

var RemoveCmd = cobra.Command{
	Use:   "rm <id | nametag>",
	Short: helpinfo.Text["id rm"],
	Run:   run_remove_cmd,
}

func init() {
	id.IdentityCmd.AddCommand(&RemoveCmd)
}

func run_remove_cmd(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		os.Exit(1)
	}

	app.OpenConfiguration()

	var remove_identity_params identities.RemoveParams
	remove_identity_params.ID = args[0]

	identities.Remove(&remove_identity_params)
}
