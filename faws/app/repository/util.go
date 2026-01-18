package repository

import (
	"time"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/app/identities"
	"github.com/faws-vcs/faws/faws/repo"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

// The repository the user is accessing
var Repo repo.Repository

var TrackerURL = tracker.DefaultURL

var quiet bool

// Open opens the repository located at directory
func Open(directory string) (err error) {
	notify_func := notify

	if quiet {
		notify_func = func(ev event.Notification, params *event.NotifyParams) {}
	}

	err = Repo.Open(directory,
		repo.WithTrust(identities.NewRingTrust(app.Configuration.Ring())),
		repo.WithNotify(notify_func),
		repo.WithTracker(TrackerURL),
	)

	if !quiet {
		console.RenderFunc(render_activity_screen)
		console.SwapInterval(time.Second / 3)
	}
	return
}

// Close closes the repository
func Close() (err error) {
	err = Repo.Close()
	return
}
