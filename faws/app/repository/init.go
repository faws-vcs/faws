package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo"
)

// Create a new repository in the current directory.
type InitParams struct {
	Directory string
	Remote    string
}

func Init(p *InitParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	initialized := repo.Exists(p.Directory)
	reinitialize := false

	if initialized {
		reinitialize = true
	}

	if err := repo.Initialize(p.Directory, p.Remote, reinitialize); err != nil {
		app.Fatal(err)
	}
	if reinitialize {
		app.Warning("Reinitialized existing repository in", p.Directory)
	} else {
		app.Info("Initialized empty repository in", p.Directory)
	}
}
