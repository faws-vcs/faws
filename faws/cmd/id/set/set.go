package set

import (
	"os"
	"time"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/identities"
	"github.com/faws-vcs/faws/faws/cmd/helpinfo"
	"github.com/faws-vcs/faws/faws/cmd/id"
	"github.com/faws-vcs/faws/faws/timestamp"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	SetCmd = cobra.Command{
		Use:   "set <id | nametag>",
		Short: helpinfo.Text["id set"],
		Run:   run_set_cmd,
	}
)

func init() {
	flags := SetCmd.Flags()

	flags.String("nametag", "", "a short string without spaces used to identify the commit author")
	flags.String("email", "", "a public email associated with the commit author")
	flags.String("description", "", "a longer text field for describing yourself. A good place to put a URL")
	flags.String("date", "", "the date these details were updated on. (DD/MM/YYYY) note that to successfully get imported into other's rings, you need to make this a later date than they already have")

	id.IdentityCmd.AddCommand(&SetCmd)
}

func run_set_cmd(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Help()
		os.Exit(1)
	}

	app.OpenConfiguration()

	flags := cmd.Flags()

	var set_identity_attributes_params identities.SetAttributesParams
	set_identity_attributes_params.ID = args[0]

	flags.Visit(func(f *pflag.Flag) {
		switch f.Name {
		case "nametag":
			set_identity_attributes_params.SetNametag = true
		case "email":
			set_identity_attributes_params.SetEmail = true
		case "description":
			set_identity_attributes_params.SetDescription = true
		case "date":
			set_identity_attributes_params.SetDate = true
		}
	})

	var err error
	if set_identity_attributes_params.SetNametag {
		set_identity_attributes_params.Attributes.Nametag, err = flags.GetString("nametag")
		if err != nil {
			app.Fatal(err)
		}
	}

	if set_identity_attributes_params.SetEmail {
		set_identity_attributes_params.Attributes.Email, err = flags.GetString("email")
		if err != nil {
			app.Fatal(err)
		}
	}

	if set_identity_attributes_params.SetDescription {
		set_identity_attributes_params.Attributes.Description, err = flags.GetString("description")
		if err != nil {
			app.Fatal(err)
		}
	}

	if set_identity_attributes_params.SetDate {
		date_string, date_string_err := flags.GetString("date")
		if date_string_err != nil {
			app.Fatal(date_string_err)
		}

		set_identity_attributes_params.Attributes.Date, err = timestamp.Parse(date_string)
		if err != nil {
			app.Fatal(err)
		}
	} else {
		// default to now
		set_identity_attributes_params.Attributes.Date = time.Now().Unix()
	}

	identities.SetAttributes(&set_identity_attributes_params)
}
