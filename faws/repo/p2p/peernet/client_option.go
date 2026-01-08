package peernet

import (
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

type client_options struct {
	set_peer_identity bool
	// How you as a peer identify yourself to the network
	peer_identity identity.Pair
	// URL of the tracker server
	tracker_url string
	// If true, forces the use of a TURN server
	use_turn bool
}

func (client_options *client_options) set_default() {
	client_options.tracker_url = tracker.DefaultURL
}

type ClientOption func(*client_options)

func WithIdentity(peer_identity identity.Pair) ClientOption {
	return func(client_options *client_options) {
		client_options.set_peer_identity = true
		client_options.peer_identity = peer_identity
	}
}

func WithForceTURN(use_turn_servers bool) ClientOption {
	return func(client_options *client_options) {
		client_options.use_turn = use_turn_servers
	}
}

func WithTrackerURL(tracker_url string) ClientOption {
	return func(client_options *client_options) {
		client_options.tracker_url = tracker_url
	}
}
