package repository

import "github.com/faws-vcs/faws/faws/app"

type AddFileParams struct {
	// The directory of the repository
	Directory string
	// The path to store the file at
	Path string
	// The path
	Origin string
}

func AddFile(params *AddFileParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := Repo.Cache(params.Path, params.Origin); err != nil {
		app.Fatal(err)
	}
	if err := Close(); err != nil {
		app.Fatal(err)
		return
	}
}
