package repo

type Ev uint8

const (
	EvPullTag = iota
	EvPullObject
	// Number of download objects
	EvPullQueueCount
)

type NotifyFunc func(ev Ev, args ...any)

func dont_care(ev Ev, args ...any) {
}

func WithNotify(fn NotifyFunc) Option {
	return func(r *Repository) {
		r.notify = fn
	}
}
