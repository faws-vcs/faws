package repository

import (
	"fmt"

	"github.com/faws-vcs/faws/faws/app"
)

type WriteTreeParams struct {
	Directory string
}

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
	fmt.Println(content_id)
}
