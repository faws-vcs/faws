package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo"
	"github.com/faws-vcs/faws/faws/repo/remote"
)

type PullParams struct {
	Directory string
	Remote    string
}

func Pull(params *PullParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

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

	if err := Repo.Pull(fs, true); err != nil {
		app.Fatal(err)
	}
}
