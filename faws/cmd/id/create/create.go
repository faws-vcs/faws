package create

import (
	"time"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/identities"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/id"
	"github.com/faws-vcs/faws/faws/timestamp"
	"github.com/spf13/cobra"
)

var CreateCmd = cobra.Command{
	Use:     "create",
	Short:   helpinfo.Text["id create"],
	Example: `faws id create --nametag "john doe" --email "john.doe@mail.com"`,
	Run:     run_create_cmd,
}

func init() {
	flags := CreateCmd.Flags()
	flags.String("nametag", "", "short user-id for your commits")
	flags.String("email", "", "the email of your new identity")
	flags.String("date", "", "the date at which these details were created (DD/MM/YYYY)")
	id.IdentityCmd.AddCommand(&CreateCmd)
}

func run_create_cmd(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	nametag, err := flags.GetString("nametag")
	if err != nil {
		app.Fatal(err)
	}
	email, err := flags.GetString("email")
	if err != nil {
		app.Fatal(err)
	}
	date_string, err := flags.GetString("date")
	if err != nil {
		app.Fatal(err)
	}

	// build attributes
	var create_identity_params identities.CreateParams
	create_identity_params.Attributes.Nametag = nametag
	create_identity_params.Attributes.Email = email

	if date_string == "" {
		create_identity_params.Attributes.Date = int64(time.Now().Unix())
	} else {
		date, err := timestamp.Parse(date_string)
		if err != nil {
			app.Fatal(err)
		}
		create_identity_params.Attributes.Date = date
	}

	identities.Create(&create_identity_params)
}
