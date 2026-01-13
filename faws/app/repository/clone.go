package repository

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/repo"
)

// CloneParams are the input parameters to the command "faws clone", [Clone]
type CloneParams struct {
	TrackerURL string
	Directory  string
	Remote     string
	Force      bool
}

// Clone is the implementation of the command "faws clone"
//
// It duplicates a remote repository into the current directory, or a named external directory.
// If Force == true, it will clone even if the directory is non-empty or the repository already exists.
func Clone(params *CloneParams) {
	if params.TrackerURL != "" {
		TrackerURL = params.TrackerURL
	}

	app.Open()
	defer func() {
		app.Close()
	}()

	directory_exists := false
	if fi, err := os.Stat(params.Directory); err == nil {
		directory_exists = fi.IsDir()
		if !directory_exists {
			app.Fatal("cannot clone into non-directory")
		}
	} else {
		err = os.Mkdir(params.Directory, fs.DefaultPublicDirPerm)
		if err != nil {
			app.Fatal(err)
		}
		directory_exists = true
	}

	if !repo.Exists(params.Directory) {
		if err := repo.Initialize(params.Directory, params.Remote, false, params.Force); err != nil {
			app.Fatal(err)
		}
	}

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := Repo.Clone(); err != nil {
		app.Fatal(err)
	}

	Close()
}
