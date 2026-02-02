package repo

import (
	"errors"
	"io"
	"runtime"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/queue"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

type visitor_queue struct {
	object_queue queue.TaskQueue[cas.ContentID]
}

func (vq *visitor_queue) init() {
	vq.object_queue.Init()
}

func (vq *visitor_queue) destroy() {
	vq.object_queue.Destroy()
}

func (repo *Repository) visitor_worker(vq *visitor_queue) (err error) {
	for {
		var (
			object_hash cas.ContentID
			prefix      cas.Prefix
			object      []byte
		)
		object_hash, err = vq.object_queue.Pop()
		if errors.Is(err, io.EOF) {
			// end of queue; not a problem
			err = nil
			return
		}

		var notify_visit_object event.NotifyParams
		notify_visit_object.Object1 = object_hash
		repo.notify(event.NotifyVisitObject, &notify_visit_object)

		prefix, object, err = repo.objects.Load(object_hash)
		if err != nil {
			break
		}

		switch prefix {
		case cas.Commit:
			var (
				commit      revision.Commit
				commit_info revision.CommitInfo
			)

			if err = revision.UnmarshalCommit(object, &commit); err != nil {
				return
			}

			if err = revision.UnmarshalCommitInfo(commit.Info, &commit_info); err != nil {
				return
			}

			vq.object_queue.Push(commit_info.Tree)

			// get parents too
			if commit_info.Parent != cas.Nil {
				vq.object_queue.Push(commit_info.Parent)
			}
		case cas.Tree:
			var tree revision.Tree
			if err = revision.UnmarshalTree(object, &tree); err != nil {
				return
			}
			for _, entry := range tree.Entries {
				vq.object_queue.Push(entry.Content)
			}
		case cas.File:
			var part_id cas.ContentID
			file_data := object
			// visit all parts
			for len(file_data) > 0 {
				copy(part_id[:], file_data[:cas.ContentIDSize])
				vq.object_queue.Push(part_id)
				file_data = file_data[cas.ContentIDSize:]
			}
		case cas.Part:
			// raw data, nothing to do except complete task
		}

		vq.object_queue.Complete(object_hash)

		var notify_object_count event.NotifyParams
		notify_object_count.Count = int64(vq.object_queue.Len())
		repo.notify(event.NotifyVisitQueueCount, &notify_object_count)
	}

	return
}

func (repo *Repository) spawn_visitor_worker(error_channel chan<- error, vq *visitor_queue) {
	err := repo.visitor_worker(vq)
	if err != nil {
		vq.object_queue.Stop()
	}
	error_channel <- err
	close(error_channel)
}

func (repo *Repository) visit_all_objects(vq *visitor_queue) (err error) {
	// staged files should be visited
	for _, staged_file := range repo.index.entries {
		vq.object_queue.Push(staged_file.file)
	}

	// start by visiting each tagged commit
	var tags []revision.Tag
	tags, err = repo.Tags()
	if err != nil {
		return
	}

	for _, tag := range tags {
		vq.object_queue.Push(tag.CommitHash)
	}

	// notify the CLI/GUI/remote client/whatever that we're starting to pull objects
	var visiting_objects event.NotifyParams
	visiting_objects.Stage = event.StageVisitObjects
	repo.notify(event.NotifyBeginStage, &visiting_objects)

	num_workers := runtime.NumCPU()
	error_channels := make([]chan error, num_workers)

	for i := 0; i < num_workers; i++ {
		error_channels[i] = make(chan error)
		go repo.spawn_visitor_worker(error_channels[i], vq)
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
		visiting_objects.Success = true
	}
	repo.notify(event.NotifyCompleteStage, &visiting_objects)

	return
}

// Prune removes all unpacked objects in the repository that cannot be visited
func (repo *Repository) PruneCache() (err error) {
	var vq visitor_queue
	vq.init()
	err = repo.visit_all_objects(&vq)
	if err != nil {
		return
	}

	var unreachable_cache_objects queue.UnorderedSet[cas.ContentID]
	unreachable_cache_objects.Init()

	repo.objects.List(func(packed bool, id cas.ContentID) (err error) {
		if !packed && !vq.object_queue.Contains(id) {
			unreachable_cache_objects.Push(id)
		}

		return
	})
	vq.destroy()

	// remove lazy signatures for unreachable files
	for lazy_signature, file := range repo.index.lazy_signatures {
		if unreachable_cache_objects.Contains(file) {
			delete(repo.index.lazy_signatures, lazy_signature)
		}
	}

	// remove the unreachable objects
	for {
		var object_hash cas.ContentID
		object_hash, err = unreachable_cache_objects.Get()
		if err != nil {
			err = nil
			break
		}
		repo.objects.Remove(object_hash)
		unreachable_cache_objects.Remove(object_hash)
		// notify ui
		var pruned_object event.NotifyParams
		pruned_object.Object1 = object_hash
		repo.notify(event.NotifyPruneObject, &pruned_object)
	}

	return
}
