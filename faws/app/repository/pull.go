package repository

import (
	"github.com/faws-vcs/faws/faws/app"
)

// PullParams are the input parameters to the command "faws pull", [Pull]
type PullParams struct {
	Directory  string
	TrackerURL string
	Ref        []string
	// If true, refs are just tags to be downloaded.
	// If len(refs) == 0, download all tags from the origin
	Tags    bool
	Verbose bool
}

// Pull is the implementation of the command "faws pull"
//
// It attempts to pull information from the remote origin repository. If Tags == true, only tags are pulled. Otherwise, a specific tree of objects with Ref at the root is pulled.
func Pull(params *PullParams) {
	if params.TrackerURL != "" {
		TrackerURL = params.TrackerURL
	}

	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	scrn.verbose = params.Verbose

	if params.Tags {
		if len(params.Ref) == 0 {
			if err := Repo.PullTags(); err != nil {
				app.Fatal(err)
			}
		} else {
			if err := Repo.PullTag(params.Ref...); err != nil {
				app.Fatal(err)
			}
		}
	} else {
		if err := Repo.Pull(params.Ref...); err != nil {
			app.Fatal(err)
		}
	}

	Close()
}
