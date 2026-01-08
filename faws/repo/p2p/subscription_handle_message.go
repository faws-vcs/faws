package p2p

import (
	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
)

func (subscription *subscription) handle_message(peer identity.ID, message_id peernet.MessageID, message []byte) {
	console.Println("received message from", peer, message_id)

	switch message_id {
	case peernet.HaveObject:
		var object_hash cas.ContentID
		// size of message is already validated
		copy(object_hash[:], message)
		//
		subscription.handle_peer_have_object(peer, object_hash)
	case peernet.WantObject:
		var object_hash cas.ContentID
		// size of message is already validated
		copy(object_hash[:], message)
		// check for the existence of the wanted object in our repository,
		// if we have it, tell the peer that we have it
		subscription.handle_peer_want_object(peer, object_hash)
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
	case peernet.RequestObject:
		var object_hash cas.ContentID
		// size of message is already validated
		copy(object_hash[:], message)
		subscription.handle_peer_request_object(peer, object_hash)
	}
}
