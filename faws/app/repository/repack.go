package repository

import "github.com/faws-vcs/faws/faws/app"

type RepackParams struct {
	// The directory where the repository is located
	Directory string
	// Sets the maximum file size of an archive (pack.XXXXXX)
	MaxArchiveSize int64
}

func Repack(params *RepackParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	scrn.summary_mode |= summarize_pruning

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := Repo.Repack(params.MaxArchiveSize); err != nil {
		app.Fatal(err)
	}

	Close()
}
