package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

type AddFileParams struct {
	// The directory of the repository
	Directory string
	// The path to store the file at
	Path string
	// The path
	Origin string
	// If true, set file mode to Mode
	SetMode bool
	Mode    revision.FileMode
	//
	AddLazy bool
}

func AddFile(params *AddFileParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	var o []repo.CacheOption
	if params.SetMode {
		o = append(o, repo.WithFileMode(params.Mode))
	}
	if params.AddLazy {
		o = append(o, repo.WithLazy(true))
	}

	if err := Repo.Cache(params.Path, params.Origin, o...); err != nil {
		app.Fatal(err)
	}
	if err := Close(); err != nil {
		app.Fatal(err)
		return
	}
}
