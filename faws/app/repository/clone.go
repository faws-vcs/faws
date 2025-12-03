package repository

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/repo"
)

type CloneParams struct {
	Directory string
	Remote    string
	Force     bool
}

func Clone(params *CloneParams) {
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

	if err := Repo.Clone(true); err != nil {
		app.Fatal(err)
	}
}
