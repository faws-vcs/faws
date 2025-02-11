package repository

import (
	"fmt"

	"github.com/faws-vcs/faws/faws/app"
)

type StatParams struct {
	Directory     string
	ShowLazyFiles bool
}

func Stat(params *StatParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	err := Open(params.Directory)
	if err != nil {
		app.Fatal(err)
		return
	}

	index := Repo.CacheIndex()
	if index != nil {
		if params.ShowLazyFiles {
			if len(index.LazySignatures) != 0 {
				app.Header("lazy files:")
				for _, lazy_signature := range index.LazySignatures {
					app.Info(lazy_signature.Signature, lazy_signature.File)
				}
			}
		}

		if len(index.Entries) == 0 {
			app.Info("nothing to commit yet")
		} else {
			app.Header(fmt.Sprintf("%d files to be committed:", len(index.Entries)))
			for _, index_entry := range index.Entries {
				app.Info(index_entry.Mode, index_entry.File, index_entry.Path)
			}
		}
	}
}
