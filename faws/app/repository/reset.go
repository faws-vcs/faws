package repository

import (
	"github.com/faws-vcs/faws/faws/app"
)

type ResetParams struct {
	Directory string
}

func Reset(params *ResetParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := Repo.ResetCache(); err != nil {
		app.Fatal(err)
	}

	Close()
}
