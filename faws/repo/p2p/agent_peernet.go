package p2p

import (
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
)

func (agent *Agent) set_peernet_handlers() {
	agent.peernet_client.OnChannelUpdate(func(topic tracker.Topic, peer identity.ID, channel_state peernet.ChannelState) {
		switch channel_state {
		case peernet.ChannelActive:
			var notify_params event.NotifyParams
			notify_params.ID = peer
			agent.options.notify(event.NotifyPeerChannelActivated, &notify_params)
		case peernet.ChannelDisconnected:
		case peernet.ChannelClosed:
		}
	})

	agent.peernet_client.OnMessage(func(topic tracker.Topic, peer identity.ID, message_id peernet.MessageID, message []byte) {
		if subscription, subscription_exists := agent.get_subscription(topic); subscription_exists {
			subscription.handle_message(peer, message_id, message)
		}
	})
}
