package repository

import (
	"github.com/faws-vcs/faws/faws/app"
)

type PullParams struct {
	Directory string
	Ref       string
	Tags      bool
	Force     bool
	Verbose   bool
}

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
