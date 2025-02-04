package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo"
	"github.com/faws-vcs/faws/faws/repo/remote"
)

type ShadowParams struct {
	Directory string
	Remote    string
	Ref       string
}

func Shadow(params *ShadowParams) {
	if !repo.Exists(params.Directory) {
		if err := repo.Initialize(params.Directory, false); err != nil {
			app.Fatal(err)
		}
	}

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	fs, err := remote.Open(params.Remote)
	if err != nil {
		app.Fatal(err)
	}

	if err := Repo.Shadow(fs, params.Ref, true); err != nil {
		app.Fatal(err)
	}
}
