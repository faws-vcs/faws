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

			subscription, is_subscribed := agent.get_subscription(topic)
			if is_subscribed {
				subscription.remove_peer(peer)
			}
		}
	})

	agent.peernet_client.OnMessage(func(topic tracker.Topic, peer identity.ID, message_id peernet.MessageID, message []byte) {
		if subscription, subscription_exists := agent.get_subscription(topic); subscription_exists {
			subscription.handle_message(peer, message_id, message)
		}
	})
}
