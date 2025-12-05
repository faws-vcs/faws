package repository

import (
	"github.com/faws-vcs/faws/faws/app"
)

// PullParams are the input parameters to the command "faws pull", [Pull]
type PullParams struct {
	Directory string
	Ref       string
	Tags      bool
	Force     bool
	Verbose   bool
}

// Pull is the implementation of the command "faws pull"
//
// It attempts to pull information from the remote origin repository. If Tags == true, only tags are pulled. Otherwise, a specific tree of objects with Ref at the root is pulled.
func Pull(params *PullParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	scrn.verbose = params.Verbose

	if params.Tags {
		if err := Repo.PullTags(params.Force); err != nil {
			app.Fatal(err)
		}
	} else {
		if err := Repo.Pull(params.Ref, params.Force); err != nil {
			app.Fatal(err)
		}
	}

	Close()
}
