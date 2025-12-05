package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

// ChmodParams are the input parameters to the command "faws chmod", [Chmod]
type ChmodParams struct {
	// The directory of the repository
	Directory string
	// The path to store the file at
	Path string
	Mode revision.FileMode
}

// Chmod is the implementation of the command "faws chmod"
//
// It changes the file mode (namely whether a file is executable or not) of a file in the index
func Chmod(params *ChmodParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := Repo.CacheSetFileMode(params.Path, params.Mode); err != nil {
		app.Fatal(err)
	}
	if err := Close(); err != nil {
		app.Fatal(err)
		return
	}
}
