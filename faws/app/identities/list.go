package identities

import (
	"strconv"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/timestamp"
)

// ListMode represents the option of displaying only public, or only secret identities, or both
type ListMode uint8

const (
	// If ListPublic is set, public identities (identities from other users) are displayed
	ListPublic ListMode = 1 << iota
	// If ListSecret is set, secret (identities that you possess the secret key for) are displayed
	ListSecret
	// ListAll, all identities in the Ring are displayed
	ListAll = ListPublic | ListSecret
)

// ListParams are the input parameters to the command "faws id ls", [List]
type ListParams struct {
	Verbose bool
	// Controls which identities are displayed to the user
	Mode ListMode
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

// List is the implementation of the command "faws id ls"
//
// It displays the contents of the user's identity ring, filtered by the [ListMode] in params.
func List(params *ListParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	ring_path := app.Configuration.RingPath()
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
