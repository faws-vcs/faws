package repo

import (
	"errors"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

func (repo *Repository) CheckObjects(id cas.ContentID, purge bool) (err error) {
	if id == cas.Nil {
		repo.objects.List(func(id cas.ContentID) (err error) {
			var (
				prefix cas.Prefix
			)
			prefix, _, err = repo.LoadObject(id)
			if err != nil {
				if errors.Is(err, cas.ErrObjectCorrupted) {
					if purge {
						err = repo.RemoveObject(id)
						if err != nil {
							return
						}
						repo.notify(EvRemovedCorruptedObject, prefix, id)
					} else {
						repo.notify(EvCorruptedObject, prefix, id)
					}
					err = nil
					return
				}

				err = nil
				return
			}
			return
		})

	}

	var (
		prefix      cas.Prefix
		object_data []byte
	)
	prefix, object_data, err = repo.LoadObject(id)
	if err != nil {
		if errors.Is(err, cas.ErrObjectCorrupted) {
			if purge {
				err = repo.RemoveObject(id)
				if err != nil {
					return
				}
				repo.notify(EvRemovedCorruptedObject, prefix, id)
			} else {
				repo.notify(EvCorruptedObject, prefix, id)
			}
			err = nil
			return
		}

		err = nil
		return
	}

	if prefix == cas.Commit {
		var commit revision.Commit
		err = revision.UnmarshalCommit(object_data, &commit)
		if err != nil {
			err = nil
			if purge {
				repo.notify(EvRemovedCorruptedObject, prefix, id)
				err = repo.RemoveObject(id)
			} else {
				repo.notify(EvCorruptedObject, prefix, id)
			}
			return
		}
		if !identity.Verify(commit.Author, &commit.Signature, commit.Info) {
			err = nil
			if purge {
				repo.notify(EvRemovedCorruptedObject, prefix, id)
				err = repo.RemoveObject(id)
			} else {
				repo.notify(EvCorruptedObject, prefix, id)
			}
			return
		}
		var commit_info revision.CommitInfo
		err = revision.UnmarshalCommitInfo(commit.Info, &commit_info)
		if err != nil {
			err = nil
			if purge {
				repo.notify(EvRemovedCorruptedObject, prefix, id)
				err = repo.RemoveObject(id)
			} else {
				repo.notify(EvCorruptedObject, prefix, id)
			}
			return
		}

		return repo.CheckObjects(commit_info.Tree, purge)
	} else if prefix == cas.Tree {
		var tree revision.Tree
		err = revision.UnmarshalTree(object_data, &tree)
		if err != nil {
			if purge {
				if err = repo.RemoveObject(id); err != nil {
					return
				}

				repo.notify(EvRemovedCorruptedObject, prefix, id)
			} else {
				repo.notify(EvCorruptedObject, prefix, id)
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
