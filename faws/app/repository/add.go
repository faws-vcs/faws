package repository

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

// AddFileParams are the input parameters to the command "faws add", [AddFile]
type AddFileParams struct {
	// The directory of the repository
	Directory string
	// The destination of the file in the index
	Destination string
	// The source file on host filesystem
	Source string
	//
	SourceIsRef bool
	// If true, set file mode to Mode
	SetMode bool
	Mode    revision.FileMode
	// Scan the file based on a small subset of available information which is assumed to be unchanging
	AddLazy bool
	// Display all files that are cached
	Verbose bool
}

// AddFile is the implementation of the command "faws add"
//
// It will add a file or directory to the index, in preparation for a tree write or commit.
func AddFile(params *AddFileParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	if err := Open(params.Directory); err != nil {
		app.Fatal(err)
	}

	// // Graceful shutdown in event of Ctrl-C is crucial to not corrupt the index
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// go func() {
	// 	for {
	// 		<-c
	// 	}
	// 	// app.Warning("'faws add' was rudely interrupted.")
	// 	// app.Warning("you may want to 'faws reset' as continuing may have undesirable effects")
	// 	// Close()
	// 	// app.Close()
	// }()

	scrn.verbose = params.Verbose

	var o []repo.StagingOption
	if params.SetMode {
		o = append(o, repo.WithFileMode(params.Mode))
	}
	if params.AddLazy {
		o = append(o, repo.WithLazy(true))
	}

	// Users can specify already-existing objects to add to the index
	if params.SourceIsRef {
		existing_object, err := Repo.ParseRef(params.Source)
		if err != nil {
			app.Fatal(err)
		}

		if err := Repo.AddObject(params.Destination, existing_object); err != nil {
			app.Fatal(err)
		}
	} else {
		// scan a file into the repository normally
		if err := Repo.Add(params.Destination, params.Source, o...); err != nil {
			// panic(err)
			app.Fatal(err)
		}
	}

	if err := Close(); err != nil {
		app.Fatal(err)
		return
	}
}
