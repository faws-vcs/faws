package repository

import (
	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/identities"
	"github.com/faws-vcs/faws/faws/repo"
)

var Repo repo.Repository

func Open(directory string) (err error) {
	err = Repo.Open(directory,
		repo.WithTrust(identities.NewRingTrust(app.Configuration.Ring())),
		repo.WithNotify(notify),
	)

	console.RenderFunc(render_activity_screen)
	return
}

func Close() (err error) {
	err = Repo.Close()
	return
}
