package identities

import (
	"github.com/faws-vcs/faws/faws/app"
)

type RemoveParams struct {
	ID string
}

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
