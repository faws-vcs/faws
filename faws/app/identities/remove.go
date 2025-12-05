package identities

import (
	"github.com/faws-vcs/faws/faws/app"
)

// RemoveParams are the input parameters to the "faws id rm" command, [Remove]
type RemoveParams struct {
	// The abbreviated fingerprint or nametag of the identity the user wishes to remove from their ring
	ID string
}

// Remove is the implementation of the "faws id rm" command
//
// It removes the named ID from the user's ring.
func Remove(params *RemoveParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	ring := app.Configuration.Ring()

	id, err := app.Configuration.Ring().Deabbreviate(params.ID)
	if err != nil {
		app.Fatal(err)
	}

	err = ring.RemoveIdentity(id)
	if err != nil {
		app.Fatal(err)
	}
	app.Log("identity", id, "removed")
}
