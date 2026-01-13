package repo

import (
	"errors"
	"io"
	"sync"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/queue"
	"github.com/faws-vcs/faws/faws/repo/remote"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

type pull_queue struct {
	object_lock  sync.Mutex
	object_locks map[[8]byte]*sync.Mutex
	object_queue queue.TaskQueue[cas.ContentID]
}

func (pq *pull_queue) init() {
	pq.object_locks = make(map[[8]byte]*sync.Mutex)
	pq.object_queue.Init()
}

func (repo *Repository) pull_object_graph_worker(error_channel chan<- error, origin remote.Origin, pq *pull_queue) {
	var err error

loop:
	for {
		var (
			object_hash cas.ContentID
			prefix      cas.Prefix
			object      []byte
		)
		object_hash, err = pq.object_queue.Pop()
		if errors.Is(err, io.EOF) {
			// end of queue; not a problem
			err = nil
			break
		}

		prefix, object, err = repo.fetch_object(origin, object_hash)
		if err != nil {
			break
		}

		// var pulled_object event.NotifyParams
		// pulled_object.Count = int64(len(object))
		// pulled_object.Prefix = prefix
		// pulled_object.Object1 = object_hash
		// repo.notify(event.NotifyPullObject, &pulled_object)

		switch prefix {
		case cas.Commit:
			var (
				commit      revision.Commit
				commit_info revision.CommitInfo
			)

			if err = revision.UnmarshalCommit(object, &commit); err != nil {
				break loop
			}

			if err = revision.UnmarshalCommitInfo(commit.Info, &commit_info); err != nil {
				break loop
			}

			pq.object_queue.Push(commit_info.Tree)

			// get parents too
			if commit_info.Parent != cas.Nil {
				pq.object_queue.Push(commit_info.Parent)
			}
		case cas.Tree:
			var tree revision.Tree
			if err = revision.UnmarshalTree(object, &tree); err != nil {
				break loop
			}
			for _, entry := range tree.Entries {
				pq.object_queue.Push(entry.Content)
			}
		case cas.File:
			var part_id cas.ContentID
			file_data := object
			for len(file_data) > 0 {
				copy(part_id[:], file_data[:cas.ContentIDSize])

				// only download file parts we don't have
				if _, err = repo.objects.Stat(part_id); err != nil {
					pq.object_queue.Push(part_id)
					err = nil
				}

				file_data = file_data[cas.ContentIDSize:]
			}
		case cas.Part:
			// raw data, nothing to do except complete task
		}

		pq.object_queue.Complete(object_hash)

		var notify_object_count event.NotifyParams
		notify_object_count.Count = int64(pq.object_queue.Len())
		repo.notify(event.NotifyPullQueueCount, &notify_object_count)
	}

	if err != nil {
		// tell all other workers to stop
		pq.object_queue.Stop()
	}

	error_channel <- err
	close(error_channel)
}

// peforms a (potentially quite enormous and slow! operation to download 1+ objects and all their children
func (repo *Repository) pull_object_graph(origin remote.Origin, objects ...cas.ContentID) (err error) {
	if len(objects) == 0 {
		return
	}

	var pq pull_queue
	pq.init()
	// starting with 1+ object tasks
	// means that the task counter will not achieve 0 until all necessary objects are downloaded
	for _, object := range objects {
		pq.object_queue.Push(object)
	}

	// notify the CLI/GUI/remote client/whatever that we're starting to pull objects
	var pulling_objects event.NotifyParams
	pulling_objects.Stage = event.StagePullObjects
	repo.notify(event.NotifyBeginStage, &pulling_objects)

	// spawn #num_workers goroutines to
	// continuously download
	// todo: reflect how many CPUs the host has
	const num_workers = 12
	error_channels := make([]chan error, 12)

	for i := 0; i < num_workers; i++ {
		error_channels[i] = make(chan error)
		go repo.pull_object_graph_worker(error_channels[i], origin, &pq)
	}

	// wait for workers to notice that the queue of tasks is complete
	for i := 0; i < num_workers; i++ {
		any_worker_error := <-error_channels[i]
		if any_worker_error != nil {
			err = any_worker_error
		}
	}

	// notify that we are done pulling objects
	if err == nil {
		pulling_objects.Success = true
	}
	repo.notify(event.NotifyCompleteStage, &pulling_objects)

	return
}
