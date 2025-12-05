package repository

import (
	"github.com/faws-vcs/faws/faws/app"
)

// ResetParams are the input parameters to the command "faws reset", [Reset]
type ResetParams struct {
	Directory string
}

// Reset is the implementation of the command "faws reset"
//
// It resets the index, clearing out all references to files
func Reset(params *ResetParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := Repo.ResetCache(); err != nil {
		app.Fatal(err)
	}

	Close()
}
