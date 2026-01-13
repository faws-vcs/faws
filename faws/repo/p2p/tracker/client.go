package tracker

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/faws-vcs/faws/faws/identity"
)

type Client struct {
	peer_identity identity.Pair
	base_url      string
	web           http.Client
	closed        atomic.Bool

	signal_handler ClientSignalHandlerFunc
	peer_handler   ClientPeerHandlerFunc

	connection              signaling_connection
	guard_subscriptions     sync.Mutex
	subscriptions           map[TopicHash]Topic
	pending_unsubscriptions chan Topic
	pending_subscriptions   chan Topic
	pending_commands        chan command_message
	shutdown                chan struct{}
}

// Init creates a tracker client for the server at tracker_url.
//
// If peer_identity != nil, it will continuously connect
// to and login in to the signaling server using this ID
func (client *Client) Init(tracker_url string, peer_identity *identity.Pair) (err error) {
	client.signal_handler = ignore_signal
	client.peer_handler = ignore_peer

	client.base_url = tracker_url
	client.base_url, _ = strings.CutSuffix(tracker_url, "/")

	if peer_identity != nil {
		client.subscriptions = make(map[TopicHash]Topic)
		client.pending_unsubscriptions = make(chan Topic)
		client.pending_subscriptions = make(chan Topic)
		client.pending_commands = make(chan command_message)
		client.shutdown = make(chan struct{})

		client.peer_identity = *peer_identity
		client.connection.init()
		go client.manage_signaling()
	}

	return
}

// Stop using the client.
func (client *Client) Shutdown() {
	client.shutdown <- struct{}{}
	close(client.shutdown)
}
