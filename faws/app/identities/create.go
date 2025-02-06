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
	app.Log("created new identity", id)
	if primary {
		app.Warning("it is now your primary ID")
	} else {
		app.Warning("however, it is not your primary ID")
	}
}
