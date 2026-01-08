package p2p

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/peernet"
)

type pull_job struct {
	subscription *subscription
	cancel       atomic.Bool
	done         sync.WaitGroup
}

func (pull_job *pull_job) init(subscription *subscription, initial_objects []cas.ContentID) {

	pull_job.subscription = subscription
	pull_job.done.Add(1)

	for _, object := range initial_objects {
		pull_job.subscription.object_wishlist.Push(object)
	}

	var notify_queue_count event.NotifyParams
	notify_queue_count.Count = int64(pull_job.subscription.object_wishlist.Len())
	pull_job.subscription.agent.options.notify(event.NotifyPullQueueCount, &notify_queue_count)

	go pull_job.spawn()
}

// func (repo *Repository) pull_object_graph_worker(error_channel chan<- error, origin remote.Origin, pq *queue.TaskQueue[cas.ContentID]) {
// 	var err error

// loop:
// 	for {
// 		var (
// 			object_hash cas.ContentID
// 			prefix      cas.Prefix
// 			object      []byte
// 		)
// 		object_hash, err = pq.PopTask()
// 		if errors.Is(err, io.EOF) {
// 			// end of queue; not a problem
// 			err = nil
// 			break
// 		}

// 		prefix, object, err = repo.fetch_object(origin, object_hash)
// 		if err != nil {
// 			break
// 		}

// 		// var pulled_object event.NotifyParams
// 		// pulled_object.Count = int64(len(object))
// 		// pulled_object.Prefix = prefix
// 		// pulled_object.Object1 = object_hash
// 		// repo.notify(event.NotifyPullObject, &pulled_object)

// 		switch prefix {
// 		case cas.Commit:
// 			var (
// 				commit      revision.Commit
// 				commit_info revision.CommitInfo
// 			)

// 			if err = revision.UnmarshalCommit(object, &commit); err != nil {
// 				break loop
// 			}

// 			if err = revision.UnmarshalCommitInfo(commit.Info, &commit_info); err != nil {
// 				break loop
// 			}

// 			pq.PushTask(commit_info.Tree)
// 		case cas.Tree:
// 			var tree revision.Tree
// 			if err = revision.UnmarshalTree(object, &tree); err != nil {
// 				break loop
// 			}
// 			for _, entry := range tree.Entries {
// 				pq.PushTask(entry.Content)
// 			}
// 		case cas.File:
// 			var part_id cas.ContentID
// 			file_data := object
// 			for len(file_data) > 0 {
// 				copy(part_id[:], file_data[:cas.ContentIDSize])

// 				// only download file parts we don't have
// 				if _, err = repo.objects.Stat(part_id); err != nil {
// 					pq.PushTask(part_id)
// 					err = nil
// 				}

// 				file_data = file_data[cas.ContentIDSize:]
// 			}
// 		case cas.Part:
// 			// raw data, nothing to do except complete task
// 		}

// 		pq.CompleteTask(object_hash)

// 		var notify_object_count event.NotifyParams
// 		notify_object_count.Count = int64(pq.Len())
// 		repo.notify(event.NotifyPullQueueCount, &notify_object_count)
// 	}

// 	if err != nil {
// 		// tell all other workers to stop
// 		pq.Stop()
// 	}

// 	error_channel <- err
// 	close(error_channel)
// }

func (pull_job *pull_job) worker_broadcast_want_rate_limit() {
	// todo: proper rate limit
	time.Sleep(10 * time.Millisecond)
}

func (pull_job *pull_job) worker() (err error) {
	subscription := pull_job.subscription

	var (
		object_hash   cas.ContentID
		object_prefix cas.Prefix
		object_data   []byte
	)

	console.Println("worker")

	// in practice, this continously re-broadcasts peernet.WantObject
	// sometimes, on the same object repeatedly.
	// this may be understood if peers are continuously joining the swarm
	for {
		if pull_job.cancel.Load() {
			return
		}

		//
		object_hash, err = subscription.object_wishlist.Pick()
		if err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			return
		}

		// console.Println("picked", object_hash)

		// first, check that we have the object.
		// if we have the object already, there's no need to ask for it
		object_prefix, object_data, err = pull_job.subscription.repository.LoadObject(object_hash)
		if err == nil {
			// treat it like we received it from a peer
			subscription.receive_object(object_hash, object_prefix, object_data)
			continue
		}
		err = nil

		// limit the amount of wants we send to peers per second
		pull_job.worker_broadcast_want_rate_limit()

		// pull_job.subscription.object_want_list.Push(object_hash[:])

		// this is an object we don't have. broadcast that we want it.
		// console.Println("broadcast", object_hash)
		pull_job.subscription.agent.peernet_client.Broadcast(pull_job.subscription.topic, peernet.WantObject, object_hash[:])
	}
}

func (pull_job *pull_job) start_worker(error_channel chan<- error) {
	error_channel <- pull_job.worker()
}

func (pull_job *pull_job) spawn() {
	var pulling_objects event.NotifyParams
	pulling_objects.Stage = event.StagePullObjects
	pull_job.subscription.agent.options.notify(event.NotifyBeginStage, &pulling_objects)

	num_workers := 1
	error_channels := make([]chan error, num_workers)

	for i := 0; i < num_workers; i++ {
		error_channels[i] = make(chan error)
		go pull_job.start_worker(error_channels[i])
	}

	pulling_objects.Success = true

	// wait for workers to notice that the queue of tasks is complete
	for i := 0; i < num_workers; i++ {
		if err := <-error_channels[i]; err != nil {
			console.Println(err)
			pulling_objects.Success = false
		}
	}

	// cause Wait() to return
	pull_job.done.Done()

	pull_job.subscription.agent.options.notify(event.NotifyCompleteStage, &pulling_objects)
}

func (pull_job *pull_job) Wait() {
	pull_job.done.Wait()
}

func (pull_job *pull_job) Cancel() {
	pull_job.cancel.Store(true)
}
