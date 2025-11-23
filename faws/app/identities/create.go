package identities

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
)

type CreateParams struct {
	Attributes identity.Attributes
}

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
