package identities

import (
	"strconv"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/timestamp"
)

type ListMode uint8

const (
	ListPublic ListMode = 1 << iota
	ListSecret
	ListAll = ListPublic | ListSecret
)

type ListParams struct {
	Verbose bool
	Mode    ListMode
}

func list_identity_entry(params *ListParams, entry *identity.RingEntry) {
	if entry.Secret {
		app.Log("secret", entry.ID)
	} else {
		app.Log("trusted", entry.ID)
	}

	if entry.Primary {
		app.Log("  (primary)")
	}

	if entry.Attributes.Nametag != "" {
		app.Log("  nametag:", entry.Attributes.Nametag)
	} else {
		app.Log("  no name")
	}

	if entry.Attributes.Email != "" {
		app.Log("  email:", "mailto:"+entry.Attributes.Email)
	}

	if entry.Attributes.Description != "" {
		app.Log("  description:", strconv.Quote(entry.Attributes.Description))
	}

	if params.Verbose {
		app.Log("  date:", timestamp.Format(entry.Attributes.Date), entry.Attributes.Date)
	} else {
		app.Log("  date:", timestamp.Format(entry.Attributes.Date))
	}

	app.Log("\n")
}

func List(params *ListParams) {
	ring_path := app.Configuration.Ring().Path()
	app.Header(ring_path)

	for entry := range app.Configuration.Ring().Entries() {
		if !entry.Secret && params.Mode&ListPublic != 0 {
			list_identity_entry(params, entry)
		}
		if entry.Secret && params.Mode&ListSecret != 0 {
			list_identity_entry(params, entry)
		}
	}
}
