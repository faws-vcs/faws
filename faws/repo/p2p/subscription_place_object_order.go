package p2p

import (
	"math/rand"
	"time"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
)

func (subscription *subscription) gather_candidates_for_object(object_hash cas.ContentID) (candidates []identity.ID, err error) {
	subscription.guard_peers.Lock()
	for _, peer := range subscription.peers {
		peer.guard.Lock()
		if peer.objects.Contains(object_hash) {
			candidates = append(candidates, peer.peer_identity)
		}
		peer.guard.Unlock()
	}
	subscription.guard_peers.Unlock()
	return
}

func (subscription *subscription) broadcast_want_object(object_hash cas.ContentID) {
	subscription.guard_peers.Lock()
	for _, peer := range subscription.peers {
		peer.want_object(object_hash)
	}
	subscription.guard_peers.Unlock()
}

// try to request an object from the network
// this should be called repeatedly on the same object until success
func (subscription *subscription) place_object_order(object_hash cas.ContentID) (err error) {
	// gather candidate peers, i.e. peers that have sent us HaveObject
	var candidates []identity.ID
	candidates, err = subscription.gather_candidates_for_object(object_hash)
	if err != nil {
		return
	}
	// shuffle list of peers. This doesn't have to be perfect, we just have to occasionally send requests to new peers
	rand.New(rand.NewSource(time.Now().UnixNano())).Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	// console.Println("place order", object_hash)

	// try to send a [novel] object request to at least one candidate
	// no biggie if none are available. this function will be called many, many times.
	for _, candidate_id := range candidates {
		candidate, candidate_err := subscription.get_peer(candidate_id)
		if candidate_err == nil {
			if candidate.request_object(object_hash) {
				break
			}
		}
	}

	// this is an object we don't have. broadcast that we want it.
	subscription.broadcast_want_object(object_hash)
	return
}
