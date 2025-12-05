package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/cas"
)

// CheckObjectsParams are the input parameters to the command "faws fsck", [CheckObjects]
type CheckObjectsParams struct {
	Directory   string
	Ref         string
	Destructive bool
}

// CheckObjects is the implementation of the command "faws fsck"
//
// It either checks all the objects in the repository, or checks a specific tree of objects using Ref as the root
func CheckObjects(params *CheckObjectsParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	var id cas.ContentID
	var err error
	if params.Ref != "" {
		id, err = Repo.ParseRef(params.Ref)
		if err != nil {
			app.Fatal(err)
		}
	}

	if err = Repo.CheckObjects(id, params.Destructive); err != nil {
		return
	}

	Close()
}
