package p2p

import (
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
)

type object_request struct {
	// valid ids include:
	//   WantObject
	//   RequestObject
	message_id    peernet.MessageID
	peer_identity identity.ID
	object_hash   cas.ContentID
}

func (subscription *subscription) object_server_worker(request_channel <-chan object_request) (err error) {
	for {
		request, ok := <-request_channel
		if !ok {
			return
		}

		switch request.message_id {
		case peernet.WantObject:
			subscription.handle_peer_want_object(request.peer_identity, request.object_hash)
		case peernet.HaveObject:
			subscription.handle_peer_have_object(request.peer_identity, request.object_hash)
		case peernet.RequestObject:
			subscription.handle_peer_request_object(request.peer_identity, request.object_hash)
		}
	}
}

func (subscription *subscription) spawn_object_server(error_channel chan<- error, request_channel <-chan object_request) {
	error_channel <- subscription.object_server_worker(request_channel)
	close(error_channel)
}

func (subscription *subscription) dispatch_object_request(message_id peernet.MessageID, peer_identity identity.ID, object_hash cas.ContentID) {
	server_slot := compute_slot(len(subscription.object_receiver_channels), object_hash)
	subscription.object_server_channels[server_slot] <- object_request{
		message_id:    message_id,
		peer_identity: peer_identity,
		object_hash:   object_hash,
	}
}
