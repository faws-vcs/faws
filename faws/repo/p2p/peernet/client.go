package peernet

import (
	"fmt"
	"strings"
	"sync"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker/models"
	"github.com/pion/webrtc/v4"
)

type (
	MessageHandlerFunc       func(topic tracker.Topic, peer identity.ID, message_id MessageID, message []byte)
	ChannelUpdateHandlerFunc func(topic tracker.Topic, peer identity.ID, channel_state ChannelState)
)

// Client serves as the interface for a peer to communicate with other peers
// It is not responsible for fully implementing the behavior of a peer.
type Client struct {
	tracker_client tracker.Client

	options client_options

	webrtc_config         webrtc.Configuration
	webrtc_setting_engine webrtc.SettingEngine
	webrtc_api            *webrtc.API
	// communication across at least one topic means that a peer connection
	// will be established.
	// a connection can have multiple channels, each corresponding to a topic both peers are
	// interested in.
	guard_topic_channels sync.Mutex
	topic_channels       map[tracker.Topic]*topic_channel

	// handlers
	channel_update_handler  ChannelUpdateHandlerFunc
	channel_message_handler MessageHandlerFunc
}

func (client *Client) Init(options ...ClientOption) (err error) {
	// client.webrtc_setting_engine.DetachDataChannels()

	client.webrtc_api = webrtc.NewAPI(webrtc.WithSettingEngine(client.webrtc_setting_engine))

	client.options.set_default()
	for _, option := range options {
		option(&client.options)
	}
	if !client.options.set_peer_identity {
		client.options.peer_identity, err = identity.New()
		if err != nil {
			return
		}
	}

	if err = client.tracker_client.Init(client.options.tracker_url, &client.options.peer_identity); err != nil {
		return
	}

	// set webrtc config
	client.webrtc_config.ICEServers = append(client.webrtc_config.ICEServers, default_ice_servers...)
	client.webrtc_config.BundlePolicy = webrtc.BundlePolicyMaxCompat
	if client.options.use_turn {
		client.webrtc_config.ICETransportPolicy = webrtc.ICETransportPolicyRelay
		var ice_server_list models.ICEServerList
		ice_server_list, err = client.tracker_client.GetICEServers()
		if err != nil {
			return
		}
		client.webrtc_config.ICEServers = append(client.webrtc_config.ICEServers, ice_server_list.ICEServers...)
		// check for TURN servers
		turn_server_exists := false
	check_for_turn_server:
		for _, ice_server := range client.webrtc_config.ICEServers {
			for _, url := range ice_server.URLs {
				if strings.HasPrefix(url, "turn:") || strings.HasPrefix(url, "turns:") {
					turn_server_exists = true
					break check_for_turn_server
				}
			}
		}

		if !turn_server_exists {
			err = fmt.Errorf("faws/p2p/peernet: force-enable TURN was requested, however no TURN relays are available")
			return
		}
	}

	client.topic_channels = make(map[tracker.Topic]*topic_channel)

	// set default handlers
	client.channel_update_handler = func(topic tracker.Topic, peer identity.ID, channel_state ChannelState) {}
	client.channel_message_handler = func(topic tracker.Topic, peer identity.ID, message_id MessageID, message []byte) {}

	client.tracker_client.OnPeer(func(topic tracker.Topic, peer identity.ID) {
		topic_channel := client.get_topic_channel(topic)

		// ignore if channel already exists for this peer
		peer_connection, created := topic_channel.get_peer_connection(peer)
		if !created {
			return
		}
		// got peer, sending offer")
		peer_connection.sdp_offer(topic)
	})

	// handle signaling messages from the tracker server
	client.tracker_client.OnSignal(func(topic tracker.Topic, peer identity.ID, signal tracker.Signal, message []byte) {
		// got signal", signal, spew.Sdump(message))

		topic_channel := client.get_topic_channel(topic)

		peer_connection, new_connection_created := topic_channel.get_peer_connection(peer)
		// this may be the first time we've heard about this peer, which is fine!

		switch signal {
		case tracker.OfferSDP:
			if err := peer_connection.handle_sdp_offer(topic, message); err != nil {
				// sdp offer", err)
			}
		case tracker.AnswerSDP:
			if !new_connection_created {
				peer_connection.handle_sdp_answer(message)
			}
		case tracker.ICECandidate:
			if err := peer_connection.handle_ice_candidate(message); err != nil {
				// ice candidate", err)
			}
		}
	})

	return
}

func (client *Client) OnChannelUpdate(channel_update_handler ChannelUpdateHandlerFunc) {
	client.channel_update_handler = channel_update_handler
}

func (client *Client) OnMessage(message_handler MessageHandlerFunc) {
	client.channel_message_handler = message_handler
}

func (client *Client) Subscribe(topic tracker.Topic) {
	client.tracker_client.Subscribe(topic)
}

func (client *Client) Unsubscribe(topic tracker.Topic) {
	client.tracker_client.Unsubscribe(topic)
	client.guard_topic_channels.Lock()
	topic_channel, topic_channel_exists := client.topic_channels[topic]
	if topic_channel_exists {
		topic_channel.shutdown()
		delete(client.topic_channels, topic)
	}
	client.guard_topic_channels.Unlock()
}

func (client *Client) Tracker() *tracker.Client {
	return &client.tracker_client
}

// Sends a directed message. Returns a non-fatal error if the peer cannot be contacted.
func (client *Client) Send(topic tracker.Topic, peer identity.ID, message_id MessageID, message []byte) (err error) {
	channel := client.get_topic_channel(topic)
	err = channel.send(peer, message_id, message)
	return
}

func (client *Client) Broadcast(topic tracker.Topic, message_id MessageID, message []byte) {
	channel := client.get_topic_channel(topic)
	channel.broadcast(message_id, message)
}

func (client *Client) Shutdown() {
	client.guard_topic_channels.Lock()
	var topics []tracker.Topic
	for topic := range client.topic_channels {
		topics = append(topics, topic)
	}
	client.guard_topic_channels.Unlock()

	for _, topic := range topics {
		client.Unsubscribe(topic)
	}

	client.tracker_client.Shutdown()
}
