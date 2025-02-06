package repository

import (
	"fmt"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

type ListTreeParams struct {
	Directory string
	Ref       string
	Recurse   bool
}

func list_tree_object(recurse bool, tree *revision.Tree, path string) {
	for _, entry := range tree.Entries {
		switch entry.Prefix {
		case cas.File:
			fmt.Println(entry.Mode, "file", entry.Content, "  ", path+entry.Name)
		case cas.Tree:
			if recurse {
				sub_tree, err := Repo.Tree(entry.Content)
				if err != nil {
					app.Fatal(err)
				}
				list_tree_object(recurse, sub_tree, path+entry.Name+"/")
			} else {
				fmt.Println(entry.Mode, "tree", entry.Content, "  ", path+entry.Name)
			}
		}
	}
}

func ListTree(params *ListTreeParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	hash, err := Repo.ParseRef(params.Ref)
	if err != nil {
		app.Fatal(err)
	}

	tree, err := Repo.Tree(hash)
	if err != nil {
		app.Fatal(err)
	}

	list_tree_object(params.Recurse, tree, "")

	if err := Close(); err != nil {
		app.Fatal(err)
		return
	}
}
