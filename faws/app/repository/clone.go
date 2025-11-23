package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo"
)

type CloneParams struct {
	Directory string
	Remote    string
}

func Clone(params *CloneParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if !repo.Exists(params.Directory) {
		if err := repo.Initialize(params.Directory, params.Remote, false); err != nil {
			app.Fatal(err)
		}
	}

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := Repo.Clone(true); err != nil {
		app.Fatal(err)
	}
}
