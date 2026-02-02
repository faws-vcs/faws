package repository

import (
	"github.com/faws-vcs/faws/faws/app"
)

// RemoveFileParams are the input parameters to the command "faws rm", [RemoveFile]
type RemoveFileParams struct {
	// The directory of the repository
	Directory string
	// The pathspec for files to be removed from the index
	Pattern string
}

// RemoveFile is the implementation of the command "faws rm"
//
// It removes a file or directory at Path from the index
func RemoveFile(params *RemoveFileParams) {
	app.Open()
	defer app.Close()
	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := Repo.Remove(params.Pattern); err != nil {
		app.Fatal(err)
	}

	if err := Close(); err != nil {
		app.Fatal(err)
		return
	}
}
