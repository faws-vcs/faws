package repository

import "github.com/faws-vcs/faws/faws/app"

type CheckoutParams struct {
	Directory   string
	Ref         string
	Destination string
	Overwrite   bool
}

func Checkout(params *CheckoutParams) {
	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	ref, err := Repo.ParseRef(params.Ref)
	if err != nil {
		app.Fatal(err)
	}

	if err := Repo.Checkout(ref, params.Destination, params.Overwrite); err != nil {
		app.Fatal(err)
	}
}
