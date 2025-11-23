package repo

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/remote"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/faws-vcs/faws/faws/validate"
)

// 1. retrieve list of tags from server
// 2. download all commits and their parents going back from current server tags
// 3. go through all commits, looking for missing objects, put them in a queue
// 4. download the missing objects, updating user on progress as we go

// NEW:
// 1. retrieve list of tags from server
// 2. read only one tag (fast; pull) or read all of them at once (slow; clone)
// 3. (Commits stage) download all commits associated with each tag, then recursively each
//     previous commit each commit has as its parent, putting each root tree into a queue
// 4. (Trees stage) put all commit's root trees into a queue. Meanwhile, the queue
//    is being drained in the background, and all higher trees are being added back to the queue
//    (increasing the job count by 1 for each sub-tree, then decreasing for the already processed root tree (order is important so job count does not reach 0 prematurely), and so on up the chain)
//    The background will only stop draining once the supply of objects is exhausted (as long as the job count is > 0)
// 5. (Files stage) look

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
			repo.objects.Remove(received_object_hash)
			err = fmt.Errorf("%w: remote object %s does not match its hash", ErrBadObject, object_hash)
			return
		}

		var notify_params event.NotifyParams
		notify_params.Prefix = prefix
		notify_params.Object1 = object_hash
		notify_params.Count = len(object_data) - 4
		repo.notify(event.NotifyPullObject, &notify_params)
	} else if err != nil {
		return
	}

	return
}

// download a tag from the remote, but do not download all its contained information at once
func (repo *Repository) pull_remote_tag(fs remote.Fs, tag string, force bool) (remote_commit_hash cas.ContentID, err error) {
	var (
		file io.ReadCloser
	)
	file, err = fs.Pull(fmt.Sprintf("tags/%s", tag))
	if err != nil {
		return
	}
	if _, err = io.ReadFull(file, remote_commit_hash[:]); err != nil {
		return
	}
	file.Close()

	var (
		local_commit_hash cas.ContentID
	)
	local_commit_hash, err = repo.read_tag(tag)
	if force || err != nil {
		var notify_params event.NotifyParams
		notify_params.Name1 = tag
		notify_params.Object1 = local_commit_hash
		notify_params.Object2 = remote_commit_hash

		repo.notify(event.NotifyPullTag, &notify_params)
		err = repo.write_tag(tag, remote_commit_hash)
	} else {
		var commit_info *revision.CommitInfo
		current_remote_commit_hash := remote_commit_hash
		should_overwrite_tag := false

		for current_remote_commit_hash != cas.Nil {
			_, _, err = repo.fetch_object(fs, current_remote_commit_hash)
			if err != nil {
				return
			}

			// look for commit info for this step in the tag's history
			_, commit_info, err = repo.check_commit(current_remote_commit_hash)
			if err != nil {
				return
			}

			// if this step in the tag's history is the current local commit hash
			if current_remote_commit_hash == local_commit_hash {
				// we should overwrite the tag
				should_overwrite_tag = true
				break
			}
			// move to the previous commit
			current_remote_commit_hash = commit_info.Parent
		}

		if should_overwrite_tag {
			var notify_params event.NotifyParams
			notify_params.Name1 = tag
			notify_params.Object1 = local_commit_hash
			notify_params.Object2 = remote_commit_hash
			repo.notify(event.NotifyPullTag, &notify_params)
			err = repo.write_tag(tag, remote_commit_hash)
		} else {
			err = ErrLocalTagNotInRemote
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

func (repo *Repository) Clone(force bool) (err error) {
	if repo.config.Remote == "" {
		err = fmt.Errorf("faws/repo: cannot clone with no remote")
		return
	}

	var fs remote.Fs
	fs, err = remote.Open(repo.config.Remote)
	if err != nil {
		return
	}

	var pull_tags_stage event.NotifyParams
	pull_tags_stage.Stage = event.StagePullTags
	repo.notify(event.NotifyBeginStage, &pull_tags_stage)

	var tags []string
	tags, err = repo.list_remote_tags(fs)
	if err != nil {
		return
	}

	tag_commit_hashes := make([]cas.ContentID, len(tags))

	for i, tag := range tags {
		if tag_commit_hashes[i], err = repo.pull_remote_tag(fs, tag, force); err != nil {
			return
		}
	}

	repo.notify(event.NotifyCompleteStage, &pull_tags_stage)

	err = repo.pull_object_graph(fs, tag_commit_hashes...)

	return
}

func (repo *Repository) deabbreviate_remote_hash(fs remote.Fs, ref string) (object_hash cas.ContentID, err error) {
	if !validate.Hex(ref) {
		err = fmt.Errorf("%w: ref abbreviation must be hexadecimal", ErrBadRef)
		return
	}

	// skip
	if len(ref) == cas.ContentIDSize*2 {
		_, err = hex.Decode(object_hash[:], []byte(ref))
		return
	}

	if len(ref) < 5 {
		err = fmt.Errorf("%w: ref abbreviation can't be less than 5 characters", ErrBadRef)
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

	err = fmt.Errorf("%w: remote hash not found", ErrBadRef)
	return
}

// Pull only objects associated with a tag or an abbreviated object hash
func (repo *Repository) Pull(ref string, force bool) (err error) {
	if repo.config.Remote == "" {
		err = fmt.Errorf("faws/repo: cannot pull into a local repository")
		return
	}

	fs, err := remote.Open(repo.config.Remote)
	if err != nil {
		return
	}

	var tags []string
	tags, err = repo.list_remote_tags(fs)
	if err != nil {
		return
	}

	var object_hash cas.ContentID
	var found_tag bool

	for _, tag := range tags {
		if tag == ref {
			found_tag = true
			object_hash, err = repo.pull_remote_tag(fs, ref, force)
			if err != nil {
				return
			}
			break
		}
	}

	if !found_tag {
		object_hash, err = repo.deabbreviate_remote_hash(fs, ref)
		if err != nil {
			return
		}
	}

	err = repo.pull_object_graph(fs, object_hash)

	return
}

func (repo *Repository) PullTags(force bool) (err error) {
	var fs remote.Fs
	fs, err = remote.Open(repo.config.Remote)
	if err != nil {
		return
	}

	var tags []string
	tags, err = repo.list_remote_tags(fs)
	if err != nil {
		return
	}

	for _, tag := range tags {
		if _, err = repo.pull_remote_tag(fs, tag, force); err != nil {
			return
		}
	}

	return
}
