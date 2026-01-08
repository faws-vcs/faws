package p2p

import (
	"bytes"
	"crypto/sha256"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

func (subscription *subscription) stat_object(object_hash cas.ContentID) (size int64, err error) {
	size, err = subscription.repository.StatObject(object_hash)
	return
}

func (subscription *subscription) load_object(object_hash cas.ContentID) (prefix cas.Prefix, object []byte, err error) {
	prefix, object, err = subscription.repository.LoadObject(object_hash)
	return
}

func (subscription *subscription) store_object(object_prefix cas.Prefix, object_data []byte) (new bool, object_hash cas.ContentID, err error) {
	new, object_hash, err = subscription.repository.StoreObject(object_prefix, object_data)
	return
}

func (subscription *subscription) handle_peer_want_object(peer identity.ID, object_hash cas.ContentID) {
	if _, err := subscription.stat_object(object_hash); err == nil {
		subscription.agent.peernet_client.Send(subscription.topic, peer, peernet.HaveObject, object_hash[:])
	}
}

func (subscription *subscription) handle_peer_have_object(peer identity.ID, object_hash cas.ContentID) {
	if subscription.object_wishlist.IsAvailable(object_hash) {
		subscription.agent.peernet_client.Send(subscription.topic, peer, peernet.RequestObject, object_hash[:])
	}
}

func (subscription *subscription) receive_object(object_hash cas.ContentID, object_prefix cas.Prefix, object_data []byte) {
	switch object_prefix {
	case cas.Commit:
		var (
			commit      revision.Commit
			commit_info revision.CommitInfo
		)

		if err := revision.UnmarshalCommit(object_data, &commit); err != nil {
			return
		}

		if err := revision.UnmarshalCommitInfo(commit.Info, &commit_info); err != nil {
			return
		}

		// get parents too
		if commit_info.Parent != cas.Nil {
			subscription.object_wishlist.Push(commit_info.Parent)
		}

		subscription.object_wishlist.Push(commit_info.Tree)
	case cas.Tree:
		var tree revision.Tree
		if err := revision.UnmarshalTree(object_data, &tree); err != nil {
			return
		}
		for _, entry := range tree.Entries {
			subscription.object_wishlist.Push(entry.Content)
		}
	case cas.File:
		var part_id cas.ContentID
		file_data := object_data
		for len(file_data) > 0 {
			copy(part_id[:], file_data[:cas.ContentIDSize])

			// only download file parts we don't have
			if _, err := subscription.stat_object(part_id); err != nil {
				subscription.object_wishlist.Push(part_id)
				err = nil
			}

			file_data = file_data[cas.ContentIDSize:]
		}
	case cas.Part:
		// raw data, nothing to do except complete task
	}

	// mark object as complete
	subscription.object_wishlist.Complete(object_hash)

	var notify_queue_count event.NotifyParams
	notify_queue_count.Count = int64(subscription.object_wishlist.Len())
	subscription.agent.options.notify(event.NotifyPullQueueCount, &notify_queue_count)

	// notify ui that it was pulled
	var notify_pull event.NotifyParams
	notify_pull.Prefix = object_prefix
	notify_pull.Object1 = object_hash
	notify_pull.Count = int64(len(object_data))
	subscription.agent.options.notify(event.NotifyPullObject, &notify_pull)
}

func (subscription *subscription) handle_peer_object(peer identity.ID, object_hash cas.ContentID, object_prefix cas.Prefix, object_data []byte) {
	// danger zone:
	// evil peers may attempt to flood us with irrelevant objects to exhaust us.
	// TODO: come up with a plan to systematically block annoying peers
	// application of penalties to naughty IP addresses or TURN relays

	console.Println("peer sent us object", object_hash, object_prefix)

	if !subscription.object_wishlist.Contains(object_hash) {
		// the peer sent an object we were never interested in.
		// this is an unmistakable violation of protocol.
		// the peer should be permanently banned, and probably their IP too.
		// subscription.apply_peer_penalty(peer, time.Hour, penalize_ip_address=true)
		console.Println("I never wanted", object_hash)
		return
	}

	if !subscription.object_wishlist.IsAvailable(object_hash) {
		// this may also be a violation of protocol, or a simple data race.
		// peers may be providing objects we requested, but simply too late (busy hard drive, slow internet connection)

		// determine
		console.Println("I don't want", object_hash, "right now")
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
		// this peer sent us a bad object.
		// criterion for permanent IP block.
		console.Println("hash mismatched", actual_object_hash, object_hash)
		return
	}

	// so this object is actually something we're interested in,
	// and we should (TODO) be reasonably certain that this peer is not part of a botnet trying to crash us.

	// now let's see

	// save the object to the repository
	subscription.store_object(object_prefix, object_data)

	// add child objects to the wishlist
	// 1
	subscription.receive_object(object_hash, object_prefix, object_data)

	// notify all peers that we now have this object
	subscription.agent.peernet_client.Broadcast(subscription.topic, peernet.HaveObject, object_hash[:])
}

func (subscription *subscription) handle_peer_request_object(peer identity.ID, object_hash cas.ContentID) {
	// TODO: screen out peers/IPs that repeatedly request the same object

	object_prefix, object_data, err := subscription.load_object(object_hash)
	if err != nil {
		return
	}

	var buffer bytes.Buffer
	buffer.Write(object_hash[:])
	buffer.Write(object_prefix[:])
	buffer.Write(object_data)

	subscription.agent.peernet_client.Send(subscription.topic, peer, peernet.Object, buffer.Bytes())
}
