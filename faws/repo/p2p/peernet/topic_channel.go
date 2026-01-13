package peernet

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
	"github.com/pion/webrtc/v4"
)

// a topic channel symbolizes a series of connections to
// other peers on the basis of interest in a shared topic
// connections are isolated based on both topic and identity
// which is important in case the same peer identity is used for different topics.
type topic_channel struct {
	stopped                atomic.Bool
	client                 *Client
	topic                  tracker.Topic
	guard_peer_connections sync.Mutex
	peer_connections       map[identity.ID]*peer_connection
}

// just creates a new topic channel structure
func (client *Client) new_topic_channel(topic tracker.Topic) (channel *topic_channel) {
	channel = new(topic_channel)
	channel.client = client
	channel.topic = topic
	channel.peer_connections = make(map[identity.ID]*peer_connection)
	return
}

// gets (or creates) a topic channel for that topic.
func (client *Client) get_topic_channel(topic tracker.Topic) (channel *topic_channel) {
	client.guard_topic_channels.Lock()
	var (
		topic_channel_found bool
	)
	channel, topic_channel_found = client.topic_channels[topic]
	if !topic_channel_found {
		channel = client.new_topic_channel(topic)
		client.topic_channels[topic] = channel
	}
	client.guard_topic_channels.Unlock()
	return
}

func (topic_channel *topic_channel) new_peer_connection(peer identity.ID) (peer_connection_instance *peer_connection) {
	var err error
	peer_connection_instance = new(peer_connection)
	peer_connection_instance.topic_channel = topic_channel
	peer_connection_instance.peer = peer
	peer_connection_instance.connection, err = topic_channel.client.webrtc_api.NewPeerConnection(topic_channel.client.webrtc_config)
	if err != nil {
		panic(err)
	}
	peer_connection_instance.create_data_channel()
	// signaling
	peer_connection_instance.connection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			return
		}
		ice_candidate_data, err := encode_ice_candidate(candidate.ToJSON())
		if err != nil {
			panic(err)
		}
		peer_connection_instance.topic_channel.client.tracker_client.Signal(topic_channel.topic, peer_connection_instance.peer, tracker.ICECandidate, ice_candidate_data)
	})
	peer_connection_instance.connection.OnConnectionStateChange(peer_connection_instance.handle_connection_state_change)
	return
}

func (topic_channel *topic_channel) get_peer_connection(id identity.ID) (peer_connection_ *peer_connection, created bool) {
	topic_channel.guard_peer_connections.Lock()
	var found bool
	peer_connection_, found = topic_channel.peer_connections[id]
	if !found {
		peer_connection_ = topic_channel.new_peer_connection(id)
		topic_channel.peer_connections[id] = peer_connection_
		created = true
	}
	topic_channel.guard_peer_connections.Unlock()
	return
}

// simply sends a message to all peers interested in a topic
func (topic_channel *topic_channel) broadcast(message_id MessageID, message []byte) {
	topic_channel.guard_peer_connections.Lock()
	for _, peer_connection := range topic_channel.peer_connections {
		peer_connection.send(message_id, message)
	}
	topic_channel.guard_peer_connections.Unlock()
}

// sends a message directed at a particular peer
func (topic_channel *topic_channel) send(peer identity.ID, message_id MessageID, message []byte) (err error) {
	topic_channel.guard_peer_connections.Lock()
	peer_connection, peer_connection_found := topic_channel.peer_connections[peer]
	if peer_connection_found {
		err = peer_connection.send(message_id, message)
	} else {
		err = fmt.Errorf("%w: %s", ErrPeerNotFound, peer)
	}
	topic_channel.guard_peer_connections.Unlock()
	return
}

func (topic_channel *topic_channel) close(peer identity.ID) (err error) {
	topic_channel.guard_peer_connections.Lock()
	peer_connection, peer_connection_exists := topic_channel.peer_connections[peer]
	if !peer_connection_exists {
		topic_channel.guard_peer_connections.Unlock()
		err = fmt.Errorf("%w: %s", ErrPeerNotFound, peer)
		return
	}
	peer_connection.data_channel.Close()
	peer_connection.connection.Close()
	delete(topic_channel.peer_connections, peer)
	topic_channel.guard_peer_connections.Unlock()
	return
}

func (topic_channel *topic_channel) shutdown() {
	topic_channel.guard_peer_connections.Lock()
	topic_channel.stopped.Store(true)
	for _, peer_connection := range topic_channel.peer_connections {
		peer_connection.close()
	}
	topic_channel.peer_connections = map[identity.ID]*peer_connection{}
	topic_channel.guard_peer_connections.Unlock()
}
