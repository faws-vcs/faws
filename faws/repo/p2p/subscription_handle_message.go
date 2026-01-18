package p2p

import (
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
)

func (subscription *subscription) handle_message(peer identity.ID, message_id peernet.MessageID, message []byte) {
	// console.Println("received message from", peer, message_id)

	var peernet_message event.NotifyParams
	peernet_message.MessageID = message_id
	peernet_message.ID = peer

	switch message_id {
	case peernet.WantObject, peernet.HaveObject, peernet.RequestObject:
		var object_hash cas.ContentID
		// size of message is already validated
		copy(object_hash[:], message)
		subscription.dispatch_object_request(message_id, peer, object_hash)
	case peernet.Object:
		var (
			object_hash   cas.ContentID
			object_prefix cas.Prefix
		)
		copy(object_hash[:], message[:cas.ContentIDSize])
		message = message[cas.ContentIDSize:]
		copy(object_prefix[:], message[:cas.PrefixSize])
		message = message[cas.PrefixSize:]

		subscription.handle_peer_object(peer, object_hash, object_prefix, message)
	}

	subscription.agent.options.notify(event.NotifyPeerNetMessage, &peernet_message)
}
