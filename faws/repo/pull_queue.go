package repo

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/remote"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

type pull_queue struct {
	// a pull task starts with 1 count, representing the root object (which is typically a commit)
	// once all child objects (in the form of pull tasks) are added, the task counter will subtract 1 and the task will be removed
	task_counter atomic.Int64
	push_cond    sync.Cond
	guard_tasks  sync.Mutex
	// contains tasks that are being processed or were already processed
	popped_tasks object_hash_set
	// contains tasks that are available for workers to process
	available_tasks object_hash_set
}

func new_pull_queue() (pq *pull_queue) {
	pq = new(pull_queue)
	pq.push_cond.L = new(sync.Mutex)
	pq.popped_tasks.Init()
	pq.available_tasks.Init()
	return
}

func (pq *pull_queue) PushTask(object cas.ContentID) {
	pq.guard_tasks.Lock()
	if !pq.popped_tasks.Contains(object) && pq.available_tasks.Push(object) {
		pq.task_counter.Add(1)
		pq.push_cond.Signal()
	}
	pq.guard_tasks.Unlock()
}

// complete a task
func (pq *pull_queue) CompleteTask(object cas.ContentID) {
	pq.guard_tasks.Lock()
	new_counter := pq.task_counter.Add(-1)
	if new_counter == 0 {
		// tells all workers to stop
		pq.push_cond.Broadcast()
	}
	pq.guard_tasks.Unlock()
}

func (pq *pull_queue) Stop() {
	pq.guard_tasks.Lock()
	pq.task_counter.Store(0)
	pq.available_tasks.Clear()
	pq.popped_tasks.Clear()
	pq.push_cond.Broadcast()
	pq.guard_tasks.Unlock()
}

// remove and return a task from the set of available tasks
// if the available task set is empty, this will block if future tasks are expected
// once empty and no future tasks are expected, will return with io.EOF
func (pq *pull_queue) PopTask() (object cas.ContentID, err error) {
	for {
		pq.guard_tasks.Lock()

		if pq.available_tasks.Len() == 0 {
			pq.guard_tasks.Unlock()
			if pq.task_counter.Load() == 0 {
				err = io.EOF
				return
			}

			// wait for more objects to be pushed to the queue
			// TODO: ensure that broadcast is sent when the task counter reaches zero
			pq.push_cond.L.Lock()
			pq.push_cond.Wait()
			pq.push_cond.L.Unlock()

			continue
		}

		var exists bool
		object, exists = pq.available_tasks.Pop()
		if !exists {
			panic("cannot remove task from set, though it is non-empty")
		}

		if !pq.popped_tasks.Push(object) {
			panic("task was already popped")
		}

		pq.guard_tasks.Unlock()

		return
	}
}

func (repo *Repository) pull_object_graph_worker(error_channel chan<- error, fs remote.Fs, pq *pull_queue) {
	var err error

loop:
	for {
		var (
			object_hash cas.ContentID
			prefix      cas.Prefix
			object      []byte
		)
		object_hash, err = pq.PopTask()
		if errors.Is(err, io.EOF) {
			// end of queue; not a problem
			err = nil
			break
		}

		prefix, object, err = repo.fetch_object(fs, object_hash)
		if err != nil {
			break
		}

		var pulled_object event.NotifyParams
		pulled_object.Count = int64(len(object))
		pulled_object.Prefix = prefix
		pulled_object.Object1 = object_hash
		repo.notify(event.NotifyPullObject, &pulled_object)

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

			pq.PushTask(commit_info.Tree)
		case cas.Tree:
			var tree revision.Tree
			if err = revision.UnmarshalTree(object, &tree); err != nil {
				break loop
			}
			for _, entry := range tree.Entries {
				pq.PushTask(entry.Content)
			}
		case cas.File:
			var part_id cas.ContentID
			file_data := object
			for len(file_data) > 0 {
				copy(part_id[:], file_data[:cas.ContentIDSize])

				// only download file parts we don't have
				if _, err = repo.objects.Stat(part_id); err != nil {
					pq.PushTask(part_id)
					err = nil
				}

				file_data = file_data[cas.ContentIDSize:]
			}
		case cas.Part:
			// raw data, nothing to do except complete task
		}

		pq.CompleteTask(object_hash)

		var notify_object_count event.NotifyParams
		notify_object_count.Count = int64(pq.available_tasks.Len() + pq.popped_tasks.Len())
		repo.notify(event.NotifyPullQueueCount, &notify_object_count)
	}

	if err != nil {
		// tell all other workers to stop
		pq.Stop()
	}

	error_channel <- err
	close(error_channel)
}

// peforms a (potentially quite enormous and slow! operation to download 1+ objects and all their children
func (repo *Repository) pull_object_graph(fs remote.Fs, objects ...cas.ContentID) (err error) {
	if len(objects) == 0 {
		return
	}

	pq := new_pull_queue()
	// starting with 1+ object tasks
	// means that the task counter will not achieve 0 until all necessary objects are downloaded
	for _, object := range objects {
		pq.PushTask(object)
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
		go repo.pull_object_graph_worker(error_channels[i], fs, pq)
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
