package identities

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
)

// CreateParams are the input parameters to the command "faws id create", [Create]
type CreateParams struct {
	Attributes identity.Attributes
}

// Create is the implementation of the command "faws id create"
//
// It generates a new signing identity using the user's provided [identity.Attributes].
//
// This new identity becomes the primary if there is not already a primary.
func Create(params *CreateParams) {
	app.Open()
	defer func() {
		app.Close()
	}()
	ring := app.Configuration.Ring()
	id, primary, err := ring.CreateIdentity(&params.Attributes)
	if err != nil {
		app.Fatal(err)
	}
	app.Log("A new identity (ID) was created: ")
	app.Quote(id.String())
	nametag := id.String()
	if params.Attributes.Nametag != "" {
		nametag = params.Attributes.Nametag
	}
	if primary {
		app.Warning("This is now your primary ID.\n")
		app.Warning(`Remember that you may update the attributes of this ID at any time:`)
		app.Quote("faws id set ", nametag, ` --nametag user.name --email user@example.com --description "info about you here"`)
	} else {
		app.Warning("This ID is not currently set as your primary ID. This means that it will not be used to author commits by default. In order to make it your primary:")
		app.Quote("faws id primary ", nametag)
		app.Warning("Alternatively, you may specify which ID you want to sign with before committing:")
		app.Quote("faws commit --sign ", nametag)
	}
}
