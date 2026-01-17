package p2p

import (
	"bytes"
	"crypto/sha256"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
)

func (subscription *subscription) handle_peer_want_object(peer identity.ID, object_hash cas.ContentID) {
	if _, err := subscription.repository.StatObject(object_hash); err == nil {
		subscription.agent.peernet_client.Send(subscription.topic, peer, peernet.HaveObject, object_hash[:])
	}
}

func (subscription *subscription) handle_peer_have_object(peer_identity identity.ID, object_hash cas.ContentID) {
	peer, err := subscription.get_peer(peer_identity)
	if err == nil {
		// to avoid memory leak from malicious users, we only confirm their ownership of objects in our own wishlist
		if subscription.object_wishlist.Contains(object_hash) {
			peer.guard.Lock()
			peer.objects.Push(object_hash)
			peer.guard.Unlock()
		} else {
			console.Println(peer_identity, "sent us a have object for an zero-interest object", object_hash)
		}
	}
}

func (subscription *subscription) handle_peer_object(peer_identity identity.ID, object_hash cas.ContentID, object_prefix cas.Prefix, object_data []byte) {
	peer, err := subscription.get_peer(peer_identity)
	if err != nil {
		return
	}

	// danger zone:
	// evil peers may attempt to flood us with irrelevant objects to exhaust us.
	// TODO: come up with a plan to systematically block annoying peers
	// application of penalties to naughty IP addresses or TURN relays

	if !subscription.object_wishlist.Contains(object_hash) {
		// the peer sent an object we were never interested in.
		// this is an unmistakable violation of protocol.
		// the peer should be permanently banned, and probably their IP too.
		// subscription.apply_peer_penalty(peer, time.Hour, penalize_ip_address=true)
		return
	}

	if subscription.object_wishlist.IsCompleted(object_hash) {
		return
	}

	peer.guard.RLock()
	requested := peer.outgoing_requested_objects.Contains(object_hash)
	peer.guard.RUnlock()

	if !requested {
		// this may also be a violation of protocol, or a simple data race.
		// peers may be providing objects we requested, but simply too late (busy hard drive, slow internet connection)
		return
	}

	// the peer might have sent us a bad object.
	// if the object doesn't match its checksum, this is a protocol violation and the peer must be blocked.
	var actual_object_hash cas.ContentID
	h := sha256.New()
	h.Write(object_prefix[:])
	h.Write(object_data[:])
	copy(actual_object_hash[:], h.Sum(nil))

	if actual_object_hash != object_hash {
		console.Println("hash mismatched", actual_object_hash, object_hash)
		return
	}

	// so this object is actually something we're interested in,
	// and we should (TODO) be reasonably certain that this peer is not part of a botnet trying to crash us.

	subscription.dispatch_object(true, object_hash, object_prefix, object_data)
}

func (subscription *subscription) handle_peer_request_object(peer identity.ID, object_hash cas.ContentID) {
	// TODO: screen out peers/IPs that repeatedly request the same object

	object_prefix, object_data, err := subscription.repository.LoadObject(object_hash)
	if err != nil {
		return
	}

	var buffer bytes.Buffer
	buffer.Write(object_hash[:])
	buffer.Write(object_prefix[:])
	buffer.Write(object_data)

	subscription.agent.peernet_client.Send(subscription.topic, peer, peernet.Object, buffer.Bytes())
}
