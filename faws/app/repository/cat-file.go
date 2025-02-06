package repository

import (
	"fmt"
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/cas"
)

type CatFileParams struct {
	Directory   string
	Prefix      string
	Ref         string
	PrettyPrint bool
}

func CatFile(params *CatFileParams) {
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

	prefix, object, err := Repo.Object(hash)
	if err != nil {
		app.Fatal(err)
	}

	if params.Prefix != "" {
		var proper_prefix cas.Prefix
		switch params.Prefix {
		case "tree":
			proper_prefix = cas.Tree
		case "commit":
			proper_prefix = cas.Commit
		case "file":
			proper_prefix = cas.File
		case "part":
			proper_prefix = cas.Part
		default:
			app.Fatal("unknown prefix")
		}
		if proper_prefix != prefix {
			app.Fatal("bad file")
		}
	}

	if params.PrettyPrint {
		switch prefix {
		case cas.Tree:
			tree, err := Repo.Tree(hash)
			if err != nil {
				app.Fatal(err)
			}
			list_tree_object(false, tree, "")
		case cas.Commit:
			display_commit(hash)
		case cas.File:
			var file_part cas.ContentID
			for len(object) > 0 {
				copy(file_part[:], object[:cas.ContentIDSize])
				fmt.Println(file_part)
				object = object[cas.ContentIDSize:]
			}
		case cas.Part:
			fmt.Println("run without -p, --pretty-print to output raw data")
		default:
			panic(prefix)
		}
		return
	}

	switch prefix {
	case cas.File:
		var file_part cas.ContentID
		for len(object) > 0 {
			copy(file_part[:], object[:cas.ContentIDSize])
			object = object[cas.ContentIDSize:]
			_, part_object, err := Repo.Object(file_part)
			if err != nil {
				app.Fatal(err)
			}
			if _, err = os.Stdout.Write(part_object); err != nil {
				app.Fatal(err)
			}
		}
	default:
		if _, err = os.Stdout.Write(object); err != nil {
			app.Fatal(err)
		}
	}

}
