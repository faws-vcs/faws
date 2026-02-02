package repo

import (
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/queue"
)

type object_pack_job struct {
	id      cas.ContentID
	size    uint32
	include bool
}

func (o object_pack_job) Less(than object_pack_job) bool {
	if o.size > than.size {
		return true
	} else if o.size < than.size {
		return false
	}
	return o.id.Less(than.id)
}

// Pack bundles together all objects of the repository into a named pack.
func (repo *Repository) Pack(name string, max_archive_size int64) (err error) {
	// gather a list of unreachable objects. These won't be included in the newly packed version of the repository
	var vq visitor_queue
	vq.init()
	err = repo.visit_all_objects(&vq)
	if err != nil {
		return
	}

	var writer cas.PackWriter
	if err = writer.Open(name, max_archive_size); err != nil {
		return
	}

	// load all object ids and sort descending by size
	var object_list queue.OrderedSet[object_pack_job]
	object_list.Init()

	if err = repo.objects.List(func(packed bool, id cas.ContentID) (err error) {
		var (
			size int64
		)
		size, err = repo.objects.Stat(id)
		if err != nil {
			return
		}
		var job object_pack_job
		job.id = id
		job.size = uint32(size)
		if vq.object_queue.Contains(id) {
			// this object will end up the final pack,
			job.include = true
		} // or it will just be removed
		object_list.Push(job)
		return
	}); err != nil {
		return
	}

	vq.destroy()

	var pack_objects event.NotifyParams
	pack_objects.Stage = event.StagePackObjects
	repo.notify(event.NotifyBeginStage, &pack_objects)

	for {
		object, popped := object_list.Pop()
		if !popped {
			break
		}
		var (
			prefix  cas.Prefix
			content []byte
		)
		prefix, content, err = repo.objects.Load(object.id)
		if err != nil {
			break
		}
		if object.include {
			_, _, err = writer.Store(prefix, content)
			if err != nil {
				break
			}
		}
	}

	pack_objects.Success = err == nil
	repo.notify(event.NotifyCompleteStage, &pack_objects)
	if err != nil {
		return
	}

	err = writer.Close()
	if err != nil {
		return
	}
	return
}
