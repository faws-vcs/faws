package repo

func WithTracker(tracker_url string) Option {
	return func(r *Repository) {
		r.tracker_url = tracker_url
	}
}
