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
		dir_entries, err := os.ReadDir(params.Directory)
		if err != nil {
			return
		}
		if len(dir_entries) != 0 && !params.Force {
			app.Fatal("refusing to clone into a non-empty directory (use -f, --force if this is what you really want to do)")
		}

		if err := repo.Initialize(params.Directory, params.Remote, false); err != nil {
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
