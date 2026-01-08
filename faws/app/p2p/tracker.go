package p2p

import (
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

type RunTrackerServerParams struct {
	Directory string
}

func RunTrackerServer(params *RunTrackerServerParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	var server tracker.Server
	if err := server.Init(params.Directory); err != nil {
		app.Fatal(err)
	}

	if err := server.Serve(); err != nil {
		app.Fatal()
	}
}
