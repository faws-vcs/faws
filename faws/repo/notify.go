package repo

type Ev uint8

const (
	EvCacheFile = iota
	EvCacheFilePart
	EvCacheUsedLazySignature
	EvPullTag
	// ( object cas.ContentID, size int )
	EvPullObject
	// ( count int )
	EvPullQueueCount
	// ( prefix cas.Prefix, object cas.ContentID)
	EvCorruptedObject
	// ( prefix cas.Prefix, object cas.ContentID)
	EvRemovedCorruptedObject
)

type NotifyFunc func(ev Ev, args ...any)

func dont_care(ev Ev, args ...any) {
}

func WithNotify(fn NotifyFunc) Option {
	return func(r *Repository) {
		r.notify = fn
	}
}
