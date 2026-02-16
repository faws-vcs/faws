package p2p

import (
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

func (agent *Agent) set_peernet_handlers() {
	agent.peernet_client.OnPeerUpdate(func(topic tracker.Topic, peer identity.ID, peer_state peernet.PeerState) {
		switch peer_state {
		case peernet.PeerConnected:
			var notify_params event.NotifyParams
			notify_params.ID = peer
			agent.options.notify(event.NotifyPeerConnected, &notify_params)

			subscription, _ := agent.get_subscription(topic)
			subscription.add_peer(peer)
		case peernet.PeerDisconnected:
			var notify_params event.NotifyParams
			notify_params.ID = peer
			agent.options.notify(event.NotifyPeerDisconnected, &notify_params)
		case peernet.PeerClosed:
			subscription, is_subscribed := agent.get_subscription(topic)
			if is_subscribed {
				subscription.remove_peer(peer)
			}
		}
	})

	agent.peernet_client.OnMessage(func(topic tracker.Topic, outgoing bool, peer identity.ID, message_guid peernet.MessageGUID, message []byte) {
		if outgoing {
			var notify_outgoing event.NotifyParams
			notify_outgoing.MessageGUID = message_guid
			notify_outgoing.ID = peer
			notify_outgoing.Outbound = true
			agent.options.notify(event.NotifyPeerNetMessage, &notify_outgoing)
			return
		}
		if subscription, subscription_exists := agent.get_subscription(topic); subscription_exists {
			subscription.handle_incoming_message(peer, message_guid, message)
		}
	})

	agent.peernet_client.OnMessageDrop(func(topic tracker.Topic, peer identity.ID, message_guid peernet.MessageGUID) {
		var notify_dropped event.NotifyParams
		notify_dropped.MessageGUID = message_guid
		notify_dropped.ID = peer
		agent.options.notify(event.NotifyPeerNetMessageDrop, &notify_dropped)
	})
}
