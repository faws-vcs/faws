package repository

import (
	"github.com/faws-vcs/faws/faws/app"
)

type StatParams struct {
	Directory string
}

func Stat(params *StatParams) {
	err := Open(params.Directory)
	if err != nil {
		app.Fatal(err)
		return
	}

	index := Repo.CacheIndex()
	if index != nil {
		if len(index.Entries) == 0 {
			app.Info("Nothing to commit yet")
		} else {
			app.Info(len(index.Entries), "files to be committed:")
			for _, index_entry := range index.Entries {
				app.Info(index_entry.Mode, index_entry.File, index_entry.Path)
			}
		}
	}
}
