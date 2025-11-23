package repo

import "github.com/faws-vcs/faws/faws/repo/event"

func dont_care(ev event.Notification, params *event.NotifyParams) {
}

func WithNotify(fn event.NotifyFunc) Option {
	return func(r *Repository) {
		r.notify = fn
	}
}
