package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo"
)

// InitParams are the input parameters to the command "faws init", [Init]
type InitParams struct {
	Directory string
	Remote    string
	Force     bool
}

// Init is the implementation of the command "faws init"
//
// It initializes a new repository in an empty directory. If Force == true, the directory can also be non-empty.
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

	if err := repo.Initialize(p.Directory, p.Remote, reinitialize, p.Force); err != nil {
		app.Fatal(err)
	}
	if reinitialize {
		app.Warning("Reinitialized existing repository in", p.Directory)
	} else {
		app.Info("Initialized empty repository in", p.Directory)
	}
}
