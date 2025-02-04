package repo

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/remote"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/faws-vcs/faws/faws/validate"
	"github.com/google/btree"
)

// 1. retrieve list of tags from server
// 2. download all commits and their parents going back from current server tags
// 3. go through all commits, looking for missing objects, put them in a queue
// 4. download the missing objects, updating user on progress as we go

type pull_queue struct {
	guard sync.Mutex
	items *btree.BTreeG[cas.ContentID]
}

func (pq *pull_queue) Len() int {
	return pq.items.Len()
}

func (pq *pull_queue) Push(object_hash cas.ContentID) {
	pq.items.ReplaceOrInsert(object_hash)
}

// when there are no more objects, ok is false.
func (pq *pull_queue) Pop() (object_hash cas.ContentID, ok bool) {
	pq.guard.Lock()
	object_hash, ok = pq.items.DeleteMin()
	pq.guard.Unlock()
	return
}

func new_pull_queue() (pq *pull_queue) {
	pq = new(pull_queue)
	pq.items = btree.NewG(2, cas.ContentID.Less)
	return
}

// type pull_list[T any] struct {
// 	Count uint32
// }

// type pull_queue_entry struct {
// 	Object cas.ContentID
// 	Prev   *pull_queue_entry
// }

// type pull_queue struct {
// 	trees   pull_list[]
// 	commits pull_queue_entry
// }

func (repo *Repository) remote_cache_path(object_hash cas.ContentID) string {
	s := object_hash.String()
	return "objects/" + s[0:2] + "/" + s[2:4] + "/" + s[4:]
}

// read an object from disk, or if missing, download it from the remote filesystem
func (repo *Repository) fetch_object(fs remote.Fs, object_hash cas.ContentID) (prefix cas.Prefix, data []byte, err error) {
	prefix, data, err = repo.objects.Load(object_hash)
	if err != nil && errors.Is(err, cas.ErrObjectNotFound) {
		var (
			object_file          io.ReadCloser
			object_data          []byte
			received_object_hash cas.ContentID
		)

		object_file, err = fs.Pull(repo.remote_cache_path(object_hash))
		if err != nil {
			return
		}

		object_data, err = io.ReadAll(object_file)
		if err != nil {
			return
		}

		object_file.Close()

		copy(prefix[:], object_data[:4])
		data = object_data[4:]

		// attempt to store data
		// also get the hash
		_, received_object_hash, err = repo.objects.Store(prefix, object_data[4:])
		if err != nil {
			return
		}

		if received_object_hash != object_hash {
			// if the hash isn't correct just delete it
			repo.objects.Delete(received_object_hash)
			err = ErrBadObject
			return
		}

		repo.notify(EvPullObject, prefix, object_hash, len(object_data)-4)
	} else if err != nil {
		return
	}

	return
}

func (repo *Repository) pull_remote_file(pq *pull_queue, fs remote.Fs, file_hash cas.ContentID) (err error) {
	var (
		prefix    cas.Prefix
		file_data []byte
	)
	prefix, file_data, err = repo.fetch_object(fs, file_hash)
	if err != nil {
		return
	}
	if prefix != cas.File {
		err = ErrBadObject
		return
	}

	var part_id cas.ContentID
	for len(file_data) > 0 {
		copy(part_id[:], file_data[:cas.ContentIDSize])

		if _, err = repo.objects.Stat(part_id); err != nil {
			pq.Push(part_id)
			err = nil
		}

		file_data = file_data[cas.ContentIDSize:]
	}
	return
}

func (repo *Repository) pull_remote_tree(pq *pull_queue, fs remote.Fs, tree_hash cas.ContentID) (err error) {
	var (
		prefix    cas.Prefix
		tree_data []byte
		tree      revision.Tree
	)
	prefix, tree_data, err = repo.fetch_object(fs, tree_hash)
	if err != nil {
		return
	}
	if prefix != cas.Tree {
		err = ErrBadObject
		return
	}

	if err = revision.UnmarshalTree(tree_data, &tree); err != nil {
		return
	}

	for _, entry := range tree.Entries {
		switch entry.Prefix {
		case cas.File:
			if err = repo.pull_remote_file(pq, fs, entry.Content); err != nil {
				return
			}
		case cas.Tree:
			if err = repo.pull_remote_tree(pq, fs, entry.Content); err != nil {
				return
			}
		default:
			err = ErrBadObject
			return
		}
	}

	return
}

// pull information for a commits and all parents, trees, and files
func (repo *Repository) pull_remote_commits(pq *pull_queue, fs remote.Fs, commit_hash cas.ContentID) (err error) {
	for commit_hash != cas.Nil {
		var (
			prefix      cas.Prefix
			commit_data []byte
		)

		prefix, commit_data, err = repo.fetch_object(fs, commit_hash)
		if err != nil {
			return
		}

		if prefix != cas.Commit {
			err = ErrBadCommit
			return
		}

		var (
			commit      revision.Commit
			commit_info revision.CommitInfo
		)

		if err = revision.UnmarshalCommit(commit_data, &commit); err != nil {
			return
		}

		if err = revision.UnmarshalCommitInfo(commit.Info, &commit_info); err != nil {
			return
		}

		if err = repo.pull_remote_tree(pq, fs, commit_info.Tree); err != nil {
			return
		}

		commit_hash = commit_info.Parent
	}

	// if commit_hash already is in the repository,

	// if _, err = repo.objects.Stat(commit_hash); err != nil && errors.Is(err, cas.ErrObjectNotFound) {

	return
}

// check if the
func (repo *Repository) pull_remote_tag(pq *pull_queue, fs remote.Fs, tag string, force bool) (err error) {
	var (
		file               io.ReadCloser
		remote_commit_hash cas.ContentID
	)
	file, err = fs.Pull(fmt.Sprintf("tags/%s", tag))
	if err != nil {
		return
	}
	if _, err = io.ReadFull(file, remote_commit_hash[:]); err != nil {
		return
	}
	file.Close()

	if err = repo.pull_remote_commits(pq, fs, remote_commit_hash); err != nil {
		return
	}

	var (
		local_commit_hash cas.ContentID
	)
	local_commit_hash, err = repo.read_tag(tag)
	if force || err != nil {
		err = repo.write_tag(tag, remote_commit_hash)
	} else {

		var commit_info *revision.CommitInfo
		current_commit_hash := remote_commit_hash
		// tag already exists
		// overwrite if remote is based on local
		overwrite := false
		for current_commit_hash != cas.Nil {
			_, commit_info, err = repo.check_commit(current_commit_hash)
			if err != nil {
				return
			}

			current_commit_hash = commit_info.Parent

			if current_commit_hash == local_commit_hash {
				overwrite = true
				break
			}
		}

		if overwrite {
			repo.notify(EvPullTag, tag, local_commit_hash, remote_commit_hash)
			err = repo.write_tag(tag, remote_commit_hash)
		}
	}

	return
}

func (repo *Repository) list_remote_tags(fs remote.Fs) (tags []string, err error) {
	entries, fs_err := fs.ReadDir("tags")
	if fs_err != nil {
		err = fs_err
		return
	}

	for _, entry := range entries {
		if !entry.IsDir {
			if validate.CommitTag(entry.Name) == nil {
				tags = append(tags, entry.Name)
			}
		}
	}

	return
}

func (repo *Repository) pull_queue(fs remote.Fs, pq *pull_queue) (err error) {
	if pq.Len() > 0 {
		repo.notify(EvPullQueueCount, int(pq.Len()))

		num_workers := 12
		var wg sync.WaitGroup
		wg.Add(num_workers)

		for i := 0; i < num_workers; i++ {
			go func() {
				for {
					next_object_hash, ok := pq.Pop()
					if !ok {
						break
					}

					if _, _, err := repo.fetch_object(fs, next_object_hash); err != nil {
						panic(err)
						// return
					}
				}

				wg.Done()
			}()
		}

		wg.Wait()
	}
	return
}

func (repo *Repository) Pull(fs remote.Fs, force bool) (err error) {
	pq := new_pull_queue()

	var tags []string
	tags, err = repo.list_remote_tags(fs)
	if err != nil {
		return
	}

	for _, tag := range tags {
		if err = repo.pull_remote_tag(pq, fs, tag, force); err != nil {
			return
		}
	}

	err = repo.pull_queue(fs, pq)

	return
}

func (repo *Repository) deabbreviate_remote_hash(fs remote.Fs, ref string) (object_hash cas.ContentID, err error) {
	if !validate.Hex(ref) {
		err = ErrBadRef
		return
	}

	// skip
	if len(ref) == cas.ContentIDSize*2 {
		_, err = hex.Decode(object_hash[:], []byte(ref))
		return
	}

	if len(ref) < 5 {
		err = ErrBadRef
		return
	}

	bucket := "objects/" + ref[0:2] + "/" + ref[2:4]

	unknown_part := ref[4:]

	var bucket_items []remote.DirEntry
	bucket_items, err = fs.ReadDir(bucket)
	if err != nil {
		return
	}

	for _, item := range bucket_items {
		if !item.IsDir {
			if strings.HasPrefix(item.Name, unknown_part) {
				_, err = hex.Decode(object_hash[:], []byte(ref[0:4]+item.Name))
				return
			}
		}
	}

	err = ErrBadRef
	return
}

// Pull only objects associated with a tag or an abbreviated object hash
func (repo *Repository) Shadow(fs remote.Fs, ref string, force bool) (err error) {
	pq := new_pull_queue()

	var tags []string
	tags, err = repo.list_remote_tags(fs)
	if err != nil {
		return
	}

	for _, tag := range tags {
		if tag == ref {
			err = repo.pull_remote_tag(pq, fs, ref, force)
			if err != nil {
				return
			}
			break
		}
	}

	var (
		object_hash cas.ContentID
		prefix      cas.Prefix
	)
	object_hash, err = repo.deabbreviate_remote_hash(fs, ref)
	if err != nil {
		return
	}

	// find what kind of object this is, so its dependencies can be fetched
	prefix, _, err = repo.fetch_object(fs, object_hash)
	if err != nil {
		return
	}

	switch prefix {
	case cas.Commit:
		err = repo.pull_remote_commits(pq, fs, object_hash)
	case cas.Tree:
		err = repo.pull_remote_tree(pq, fs, object_hash)
	case cas.File:
		err = repo.pull_remote_file(pq, fs, object_hash)
	}

	if err != nil {
		return
	}

	err = repo.pull_queue(fs, pq)

	return
}
