package identities

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/validate"
)

// SetAttributesParams are the input parameters to to the command "faws id set", [SetAttributes]
type SetAttributesParams struct {
	ID string

	SetPrimary bool

	SetNametag     bool
	SetEmail       bool
	SetDescription bool
	SetDate        bool

	Attributes identity.Attributes
}

// SetAttributes is the implementation of the "faws id set" command
//
// For each "Set" boolean in [SetAttributesParams], it will set the corresponding attribute for the named secret identity.
func SetAttributes(params *SetAttributesParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	ring := app.Configuration.Ring()

	id, err := ring.Deabbreviate(params.ID)
	if err != nil {
		app.Fatal(err)
	}

	var attributes identity.Attributes
	if err := ring.GetAttributesSecret(id, &attributes); err != nil {
		app.Fatal(err)
		return
	}

	if params.SetPrimary {
		app.Log("setting primary")
		err = ring.SetPrimary(id)
		if err != nil {
			app.Fatal(err)
		}
		app.Log("set primary identity to", id)
	}

	if params.SetNametag {
		app.Log("setting nametag")
		// nametag can't already be in use
		in_use_id, err := ring.Lookup(params.Attributes.Nametag)
		if err == nil && in_use_id != id {
			app.Fatal("'"+params.Attributes.Nametag+"'", "already in use", "("+in_use_id.String()+")")
		}

		if err := validate.Nametag(params.Attributes.Nametag); err != nil {
			app.Fatal(err)
		}
		attributes.Nametag = params.Attributes.Nametag
	}

	if params.SetEmail {
		app.Log("setting email")
		if err := validate.Email(params.Attributes.Email); err != nil {
			app.Fatal(err)
		}
		attributes.Email = params.Attributes.Email
	}

	if params.SetDescription {
		app.Log("setting description")
		if err := validate.Description(params.Attributes.Description); err != nil {
			app.Fatal(err)
		}
		attributes.Description = params.Attributes.Description
	}

	if params.SetDate {
		attributes.Date = params.Attributes.Date
	}

	if err := ring.SetAttributesSecret(id, &attributes); err != nil {
		return
	}
}
