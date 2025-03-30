package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/cas"
)

type CheckObjectsParams struct {
	Directory   string
	Ref         string
	Destructive bool
}

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
