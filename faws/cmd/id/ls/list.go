package ls

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/identities"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/id"
	"github.com/spf13/cobra"
)

var ListCmd = cobra.Command{
	Use:   "ls",
	Short: helpinfo.Text["id ls"],
	Run:   run_list_cmd,
}

func init() {
	flags := ListCmd.Flags()
	flags.BoolP("trusted", "t", false, "list only my trusted IDs")
	flags.BoolP("secret", "s", false, "list only my secret identity keys")
	flags.BoolP("verbose", "v", false, "list hidden information")
	id.IdentityCmd.AddCommand(&ListCmd)
}

func run_list_cmd(cmd *cobra.Command, args []string) {
	// Ensure open configuration
	app.OpenConfiguration()

	//
	flags := cmd.Flags()
	public_only, err := flags.GetBool("trusted")
	if err != nil {
		app.Fatal(err)
	}
	secret_only, err := flags.GetBool("secret")
	if err != nil {
		app.Fatal(err)
	}
	verbose, err := flags.GetBool("verbose")
	if err != nil {
		app.Fatal(err)
	}

	var list_identities_params identities.ListParams
	list_identities_params.Verbose = verbose
	list_identities_params.Mode = identities.ListAll
	if secret_only {
		list_identities_params.Mode &= ^identities.ListPublic
	} else if public_only {
		list_identities_params.Mode &= ^identities.ListSecret
	}

	identities.List(&list_identities_params)
}
