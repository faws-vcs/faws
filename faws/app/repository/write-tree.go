package repository

import (
	"github.com/faws-vcs/faws/faws/app"
)

// WriteTreeParams are the input parameters to the command "faws write-tree", [WriteTree]
type WriteTreeParams struct {
	Directory string
}

// WriteTree is the implementation of the command "faws write-tree"
//
// It creates a tree object (which may already exist) using files the user added index
func WriteTree(params *WriteTreeParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	content_id, err := Repo.WriteTree()
	if err != nil {
		app.Fatal(err)
	}
	if err := Close(); err != nil {
		app.Fatal(err)
		return
	}
	app.Info(content_id)
}
