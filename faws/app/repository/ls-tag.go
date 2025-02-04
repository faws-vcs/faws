package repository

import (
	"fmt"

	"github.com/faws-vcs/faws/faws/app"
)

type ListTagsParams struct {
	Directory string
}

func ListTags(params *ListTagsParams) {
	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
		return
	}

	tags, err := Repo.Tags()
	if err != nil {
		app.Fatal(err)
	}

	for _, tag := range tags {
		fmt.Println(tag.Hash, tag.Name)
	}

	Close()
}
