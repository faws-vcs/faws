package p2p

import (
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/event"
)

type agent_options struct {
	// if >= 0, determines how many bytes we are allowed to upload per second (per repository).
	upload_bytes_per_second int64
	// if >= 0, determines how many bytes we are allowed to download per second (per repository).
	download_bytes_per_second int64
	// if false,
	set_peer_identity bool
	// How you as a peer identify yourself to the network
	peer_identity identity.Pair
	// URL of the tracker server
	tracker_url string
	// If true, forces the use of a TURN server
	use_turn bool
	notify   event.NotifyFunc
}

type Option func(*agent_options)

func WithIdentity(peer_identity identity.Pair) Option {
	return func(a *agent_options) {
		a.peer_identity = peer_identity
		a.set_peer_identity = true
	}
}

func WithNotify(notify_func event.NotifyFunc) Option {
	return func(a *agent_options) {
		a.notify = notify_func
	}
}

func WithTrackerURL(tracker_url string) Option {
	return func(a *agent_options) {
		a.tracker_url = tracker_url
	}
}
