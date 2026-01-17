package p2p

import (
	"sync"
	"time"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
	"github.com/faws-vcs/faws/faws/repo/queue"
)

// how long to wait before sending another want request
var want_ttl = time.Minute

type peer struct {
	subscription *subscription

	peer_identity identity.ID

	guard sync.RWMutex

	// objects this peer contains
	objects queue.UnorderedSet[cas.ContentID]

	// objects we wanted from this peer
	outgoing_wanted_objects map[cas.ContentID]time.Time

	// objects we requested
	outgoing_requested_objects queue.UnorderedSet[cas.ContentID]
}

func (subscription *subscription) add_peer(peer_identity identity.ID) (peer_ *peer) {
	subscription.guard_peers.Lock()
	var exists bool
	peer_, exists = subscription.peers[peer_identity]
	if !exists {
		peer_ = new(peer)
		peer_.peer_identity = peer_identity
		peer_.subscription = subscription
		peer_.outgoing_wanted_objects = make(map[cas.ContentID]time.Time)
		peer_.outgoing_requested_objects.Init()
		peer_.objects.Init()
		subscription.peers[peer_identity] = peer_
	}
	subscription.guard_peers.Unlock()
	return
}

func (subscription *subscription) get_peer(peer_identity identity.ID) (peer_ *peer, err error) {
	subscription.guard_peers.RLock()
	var exists bool
	peer_, exists = subscription.peers[peer_identity]
	if !exists {
		err = ErrSubscriptonPeerNotFound
		return
	}
	subscription.guard_peers.RUnlock()
	return
}

func (subscription *subscription) remove_peer(peer_identity identity.ID) (removed bool) {
	subscription.guard_peers.Lock()
	_, removed = subscription.peers[peer_identity]
	if removed {
		delete(subscription.peers, peer_identity)
	}
	subscription.guard_peers.Unlock()
	return
}

func (peer *peer) want_object(object_hash cas.ContentID) {
	peer.guard.Lock()
	defer peer.guard.Unlock()
	subscription := peer.subscription

	now := time.Now()

	// if peer is already a candidate for this object, we don't need to want it - they've got it
	if !peer.objects.Contains(object_hash) {
		// if enough time has elapsed since the last want message, we can send another
		time_, already_wanted := peer.outgoing_wanted_objects[object_hash]
		if already_wanted && now.Sub(time_) > want_ttl || !already_wanted {
			peer.outgoing_wanted_objects[object_hash] = now
			subscription.agent.peernet_client.Send(subscription.topic, peer.peer_identity, peernet.WantObject, object_hash[:])
		}
	}
}

func (peer *peer) request_object(object_hash cas.ContentID) (requested bool) {
	peer.guard.Lock()
	defer peer.guard.Unlock()

	if !peer.outgoing_requested_objects.Contains(object_hash) {
		err := peer.subscription.agent.peernet_client.Send(peer.subscription.topic, peer.peer_identity, peernet.RequestObject, object_hash[:])
		if err == nil {
			peer.outgoing_requested_objects.Push(object_hash)
			requested = true
		}
	}

	return
}
