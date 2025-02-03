package repository

import (
	"strings"

	"github.com/faws-vcs/faws/faws/app"
)

type RemoveFileParams struct {
	// The directory of the repository
	Directory string
	// If true, deletes any files contained in a directory
	Recurse bool
	// The path in the index
	Path string
}

func RemoveFile(params *RemoveFileParams) {
	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	index := Repo.CacheIndex()

	if params.Path == "" {
		if err := Repo.EmptyCache(); err != nil {
			app.Fatal(err)
		}
	} else {
		path := strings.TrimRight(params.Path, "/")

		if !params.Recurse {
			if err := Repo.Uncache(params.Path); err != nil {
				app.Fatal(err)
			}
		} else {
			directory := path + "/"
			var remove_list []string

			for _, entry := range index.Entries {
				if strings.HasPrefix(entry.Path, directory) {
					remove_list = append(remove_list, entry.Path)
				} else if entry.Path == path {
					remove_list = append(remove_list, path)
				}
			}

			for _, rm_path := range remove_list {
				if err := Repo.Uncache(rm_path); err != nil {
					app.Fatal(err)
				}
			}
		}
	}

	if err := Close(); err != nil {
		app.Fatal(err)
		return
	}
}
