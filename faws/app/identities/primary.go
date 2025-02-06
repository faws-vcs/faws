package identities

import (
	"github.com/faws-vcs/faws/faws/app"
)

type SetPrimaryParams struct {
	ID string
}

func SetPrimary(params *SetPrimaryParams) {
	app.Open()
	defer func() {
		app.Close()
	}()
	ring := app.Configuration.Ring()
	id, err := ring.Deabbreviate(params.ID)
	if err != nil {
		app.Fatal(err)
	}
	err = ring.SetPrimary(id)
	if err != nil {
		app.Fatal(err)
	}
	app.Log("set primary identity to", id)
}
