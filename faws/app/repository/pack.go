package repository

import "github.com/faws-vcs/faws/faws/app"

type PackParams struct {
	// The directory where the repository is located
	Directory string

	// The name of the pack
	Name string

	// Sets the maximum file size of an archive (pack.XXXXXX)
	MaxArchiveSize int64
}

func Pack(params *PackParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := Repo.Pack(params.Name, params.MaxArchiveSize); err != nil {
		app.Fatal(err)
	}

	Close()
}
