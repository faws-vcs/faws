package primary

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/identities"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/id"
	"github.com/spf13/cobra"
)

var PrimaryCmd = cobra.Command{
	Use:   "primary id",
	Short: helpinfo.Text["id primary"],
	Run:   run_primary_cmd,
}

func init() {
	id.IdentityCmd.AddCommand(&PrimaryCmd)
}

func run_primary_cmd(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		os.Exit(1)
	}

	app.OpenConfiguration()

	var set_primary_params identities.SetPrimaryParams
	set_primary_params.ID = args[0]
	identities.SetPrimary(&set_primary_params)

}
