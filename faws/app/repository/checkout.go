package repository

import "github.com/faws-vcs/faws/faws/app"

// CheckoutParams are the input parameters to the command "faws checkout", [Checkout]
type CheckoutParams struct {
	Directory   string
	Ref         string
	Destination string
	Overwrite   bool
}

// Checkout is the implementation of the command "faws checkout"
//
// It takes in a ref and exports its content to a given destination file path.
func Checkout(params *CheckoutParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

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
