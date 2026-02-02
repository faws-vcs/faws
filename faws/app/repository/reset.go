package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/cas"
)

// ResetParams are the input parameters to the command "faws reset", [Reset]
type ResetParams struct {
	Directory string
	// tag or commit hash to reset the index to
	Ref string
}

// Reset is the implementation of the command "faws reset"
//
// It resets the index, clearing out all references to files
func Reset(params *ResetParams) {
	app.Open()
	defer func() {
		app.Close()
	}()
	var (
		err         error
		commit_hash cas.ContentID
	)

	if err = Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if params.Ref != "" {
		commit_hash, err = Repo.ParseRef(params.Ref)
		if err != nil {
			app.Fatal(err)
		}
	}

	if err := Repo.Reset(commit_hash); err != nil {
		app.Fatal(err)
	}

	Close()
}
