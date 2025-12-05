package repository

import (
	"github.com/faws-vcs/faws/faws/app"
)

// ListTagsParams are the input parameters to the command "faws tag", [ListTags]
type ListTagsParams struct {
	Directory string
	Name      string
}

// ListTags is the implementation of the command "faws tag"
//
// It lists all the commit tags in the repository. If Tag != "", the commit hash associated with a commit is displayed.
func ListTags(params *ListTagsParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
		return
	}

	if params.Name != "" {
		commit_hash, err := Repo.Tag(params.Name)
		if err != nil {
			app.Fatal(err)
		}
		app.Info(commit_hash)
	} else {
		tags, err := Repo.Tags()
		if err != nil {
			app.Fatal(err)
		}

		for _, tag := range tags {
			app.Info(tag.Hash, tag.Name)
		}
	}

	Close()
}
