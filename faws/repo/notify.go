package repo

type Ev uint8

const (
	EvPullTag = iota
	// ( object cas.ContentID, size int )
	EvPullObject
	// ( count int )
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
