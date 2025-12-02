package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/faws-vcs/faws/faws/validate"
)

type CommitTreeParams struct {
	CommitParams
	Tree string
}

func CommitTree(params *CommitTreeParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	var (
		signing_identity  identity.Pair
		author_attributes identity.Attributes
	)

	ring := app.Configuration.Ring()

	var err error
	if params.Sign == "" {
		err = ring.GetPrimaryPair(&signing_identity, &author_attributes)
		if err != nil {
			app.Warning("You don't seem to have a signing identity yet. use")
			app.Warning("  faws id create")
			app.Warning("to create one")
		}
	} else {
		err = ring.GetNametagPair(params.Sign, &signing_identity, &author_attributes)
	}
	if err != nil {
		app.Fatal(err)
	}

	var p identity.Pair
	if signing_identity == p {
		panic(signing_identity)
	}

	// build commit info
	var commit_info revision.CommitInfo
	commit_info.AuthorAttributes = author_attributes
	if params.Parent != "" {
		commit_info.Parent, err = Repo.ParseRef(params.Parent)
		if err != nil {
			app.Fatal(err)
		}
	}
	if params.Tree == "" {
		app.Fatal("you must provide a tree object")
	}
	commit_info.Tree, err = Repo.ParseRef(params.Tree)
	if err != nil {
		app.Fatal(err)
	}
	commit_info.CommitDate = params.CommitDate
	commit_info.TreeDate = params.TreeDate

	err = validate.CommitTag(params.Tag)
	if err != nil {
		app.Fatal(err)
	}
	commit_info.Tag = params.Tag

	content_id, err := Repo.CommitTree(&signing_identity, &commit_info)
	if err != nil {
		app.Fatal(err)
	}
	if err := Close(); err != nil {
		app.Fatal(err)
		return
	}
	app.Info(content_id)
}
