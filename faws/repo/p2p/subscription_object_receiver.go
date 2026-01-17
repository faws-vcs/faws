package p2p

import (
	"math/big"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/queue"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

type named_object struct {
	novel  bool
	name   cas.ContentID
	prefix cas.Prefix
	data   []byte
}

func (subscription *subscription) process_object(object_hash cas.ContentID, object_prefix cas.Prefix, object_data []byte) {

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
			if _, err := subscription.repository.StatObject(part_id); err != nil {
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

	// remove requests and wants
	subscription.guard_peers.RLock()
	for _, peer := range subscription.peers {
		peer.guard.Lock()
		peer.outgoing_requested_objects.Remove(object_hash)
		delete(peer.outgoing_wanted_objects, object_hash)
		peer.guard.Unlock()
	}
	subscription.guard_peers.RUnlock()

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

func (subscription *subscription) object_receiver_worker(object_receiver_channel <-chan named_object) (err error) {
	var processed_list queue.UnorderedSet[cas.ContentID]
	processed_list.Init()

	for {
		named_object, ok := <-object_receiver_channel
		if !ok {
			return
		}

		// we already dealt with this, ignore.
		if processed_list.Contains(named_object.name) {
			continue
		}

		// if the object is novel, it has to be stored to disk
		if named_object.novel {
			if _, _, err = subscription.repository.StoreObject(named_object.prefix, named_object.data); err != nil {
				return
			}
		}

		// process the object, including all its children
		subscription.process_object(named_object.name, named_object.prefix, named_object.data)
		// in this worker, mark the object as completed.
		processed_list.Push(named_object.name)
	}
}

// from the hash and the number of receiver workers, determine which receiver this object hash is sent to
func compute_slot(num_slots int, object_hash cas.ContentID) int {
	var object_number big.Int
	object_number.SetBytes(object_hash[:])

	// domain = 2^160-1
	var domain big.Int
	domain.Exp(big.NewInt(2), big.NewInt(cas.ContentIDSize*8), nil)
	domain.Sub(&domain, big.NewInt(1))

	// fraction = domain / num_workers
	var fraction big.Int
	fraction.Div(&domain, big.NewInt(int64(num_slots)))

	// slot = object_number / fraction
	var slot big.Int
	slot.Div(&object_number, &fraction)

	// don't go out of bounds
	// slot = min(slot, num_receivers-1)
	return min(int(slot.Int64()), num_slots-1)
}

func (subscription *subscription) spawn_object_receiver(err chan<- error, object_receiver_channel <-chan named_object) {
	err <- subscription.object_receiver_worker(object_receiver_channel)
	close(err)
}

func (subscription *subscription) dispatch_object(novel bool, object_hash cas.ContentID, object_prefix cas.Prefix, object_data []byte) {
	// receiver_slot := subscription.compute_receiver_slot(object_hash)
	receiver_slot := compute_slot(len(subscription.object_receiver_channels), object_hash)
	subscription.object_receiver_channels[receiver_slot] <- named_object{
		novel:  novel,
		name:   object_hash,
		prefix: object_prefix,
		data:   object_data,
	}
}
