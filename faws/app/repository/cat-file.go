package repository

import (
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/cas"
)

// CatFileParams are the input parameters to the command "faws cat-file", [CatFile]
type CatFileParams struct {
	// The directory containing the Faws repository
	Directory   string
	Ref         string
	PrettyPrint bool
}

// CatFile implements the command "faws cat-file"
//
// It will load an object, and display its contents to stdout. If -p, --pretty-print is passed, it will be formatted and not spit out raw binary data.
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

	prefix, object, err := Repo.LoadObject(hash)
	if err != nil {
		app.Fatal(err)
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
				app.Info(file_part)
				object = object[cas.ContentIDSize:]
			}
		case cas.Part:
			app.Info("run without -p, --pretty-print to output raw data")
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
			_, part_object, err := Repo.LoadObject(file_part)
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
