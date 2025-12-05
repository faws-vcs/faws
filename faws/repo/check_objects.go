package repo

import (
	"errors"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

// CheckObjects checks a list of objects in the repo for consistency
//
// If id != cas.Nil, the ID is used as the root in a tree of objects, and all children are recursively checked for consistency.
// If id == nil, each object is checked for consistency, including orphaned objects.
// If purge == false, [event.NotifyCorruptedObject] is generated upon encountering an inconsistent or corrupt object.
// If purge == true,  [event.NotifyRemovedCorruptedObject] is generated upon encountering an inconsistent or corrupt object, and the object is deleted.
func (repo *Repository) CheckObjects(id cas.ContentID, purge bool) (err error) {
	if id == cas.Nil {
		err = repo.objects.List(func(id cas.ContentID) (err error) {
			var (
				prefix cas.Prefix
			)
			prefix, _, err = repo.LoadObject(id)
			if err != nil {
				if errors.Is(err, cas.ErrObjectCorrupted) {
					var notify_params event.NotifyParams
					notify_params.Prefix = prefix
					notify_params.Object1 = id
					if purge {
						err = repo.RemoveObject(id)
						if err != nil {
							return
						}

						repo.notify(event.NotifyRemovedCorruptedObject, &notify_params)
					} else {
						repo.notify(event.NotifyCorruptedObject, &notify_params)
					}
					err = nil
					return
				}

				err = nil
				return
			}
			return
		})
		return
	}

	var (
		prefix      cas.Prefix
		object_data []byte
	)
	prefix, object_data, err = repo.LoadObject(id)
	if err != nil {
		if errors.Is(err, cas.ErrObjectCorrupted) {
			var notify_params event.NotifyParams
			notify_params.Prefix = prefix
			notify_params.Object1 = id
			if purge {
				err = repo.RemoveObject(id)
				if err != nil {
					return
				}
				repo.notify(event.NotifyRemovedCorruptedObject, &notify_params)
			} else {
				repo.notify(event.NotifyCorruptedObject, &notify_params)
			}
			err = nil
			return
		}

		err = nil
	}

	if prefix == cas.Commit {
		var commit revision.Commit
		err = revision.UnmarshalCommit(object_data, &commit)
		if err != nil {
			err = nil
			var notify_params event.NotifyParams
			notify_params.Prefix = prefix
			notify_params.Object1 = id
			if purge {
				repo.notify(event.NotifyRemovedCorruptedObject, &notify_params)
				err = repo.RemoveObject(id)
			} else {
				repo.notify(event.NotifyCorruptedObject, &notify_params)
			}
			return
		}
		if !identity.Verify(commit.Author, &commit.Signature, commit.Info) {
			err = nil
			var notify_params event.NotifyParams
			notify_params.Prefix = prefix
			notify_params.Object1 = id
			if purge {
				repo.notify(event.NotifyRemovedCorruptedObject, &notify_params)
				err = repo.RemoveObject(id)
			} else {
				repo.notify(event.NotifyCorruptedObject, &notify_params)
			}
			return
		}
		var commit_info revision.CommitInfo
		err = revision.UnmarshalCommitInfo(commit.Info, &commit_info)
		if err != nil {
			err = nil
			var notify_params event.NotifyParams
			notify_params.Prefix = prefix
			notify_params.Object1 = id
			if purge {
				repo.notify(event.NotifyRemovedCorruptedObject, &notify_params)
				err = repo.RemoveObject(id)
			} else {
				repo.notify(event.NotifyCorruptedObject, &notify_params)
			}
			return
		}

		return repo.CheckObjects(commit_info.Tree, purge)
	} else if prefix == cas.Tree {
		var tree revision.Tree
		err = revision.UnmarshalTree(object_data, &tree)
		if err != nil {
			var notify_params event.NotifyParams
			notify_params.Prefix = prefix
			notify_params.Object1 = id
			if purge {
				if err = repo.RemoveObject(id); err != nil {
					return
				}

				repo.notify(event.NotifyRemovedCorruptedObject, &notify_params)
			} else {
				repo.notify(event.NotifyCorruptedObject, &notify_params)
			}

			err = nil
			return
		}

		for _, tree_entry := range tree.Entries {
			if err = repo.CheckObjects(tree_entry.Content, purge); err != nil {
				return
			}
		}
	} else if prefix == cas.File {
		var part_id cas.ContentID
		for len(object_data) > 0 {
			copy(part_id[:], object_data[:cas.ContentIDSize])
			object_data = object_data[cas.ContentIDSize:]
			if err = repo.CheckObjects(part_id, purge); err != nil {
				return
			}
		}
	}

	return
}
