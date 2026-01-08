package repo

import (
	"errors"
	"fmt"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
	"github.com/faws-vcs/faws/faws/repo/remote"
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

// read an object from disk, or if missing, download it from the remote filesystem
func (repo *Repository) fetch_object(origin remote.Origin, object_hash cas.ContentID) (prefix cas.Prefix, data []byte, err error) {
	prefix, data, err = repo.objects.Load(object_hash)
	if err != nil && errors.Is(err, cas.ErrObjectNotFound) {
		var (
			received_object_hash cas.ContentID
		)

		prefix, data, err = origin.GetObject(object_hash)
		if err != nil {
			return
		}

		// attempt to store data
		// also get the hash
		_, received_object_hash, err = repo.objects.Store(prefix, data)
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
		notify_params.Count = int64(len(data))
		repo.notify(event.NotifyPullObject, &notify_params)
	} else if err != nil {
		return
	}

	return
}

// // attempt to replace a local tag (if it exists) with a remote one. returns an error if there is a conflict
// // that would result in the local tag becoming inaccessible (unless force == true)
// func (repo *Repository) record_remote_tag(origin remote.Origin, remote_tag revision.Tag, force bool) (err error) {
// 	remote_commit_hash := remote_tag.CommitHash

// 	var (
// 		local_commit_hash cas.ContentID
// 	)
// 	local_commit_hash, err = repo.read_tag(remote_tag.Name)
// 	if force || err != nil {
// 		var notify_params event.NotifyParams
// 		notify_params.Name1 = remote_tag.Name
// 		notify_params.Object1 = local_commit_hash
// 		notify_params.Object2 = remote_commit_hash

// 		repo.notify(event.NotifyPullTag, &notify_params)
// 		err = repo.write_tag(remote_tag.Name, remote_commit_hash)
// 	} else {
// 		var commit_info *revision.CommitInfo
// 		current_remote_commit_hash := remote_commit_hash
// 		should_overwrite_tag := false

// 		for current_remote_commit_hash != cas.Nil {
// 			_, _, err = repo.fetch_object(origin, current_remote_commit_hash)
// 			if err != nil {
// 				return
// 			}

// 			// look for commit info for this step in the tag's history
// 			_, commit_info, err = repo.check_commit(current_remote_commit_hash)
// 			if err != nil {
// 				return
// 			}

// 			// if this step in the tag's history is the current local commit hash
// 			if current_remote_commit_hash == local_commit_hash {
// 				// we should overwrite the tag
// 				should_overwrite_tag = true
// 				break
// 			}
// 			// move to the previous commit
// 			current_remote_commit_hash = commit_info.Parent
// 		}

// 		if should_overwrite_tag {
// 			var notify_params event.NotifyParams
// 			notify_params.Name1 = remote_tag.Name
// 			notify_params.Object1 = local_commit_hash
// 			notify_params.Object2 = remote_commit_hash
// 			repo.notify(event.NotifyPullTag, &notify_params)
// 			err = repo.write_tag(remote_tag.Name, remote_commit_hash)
// 		} else {
// 			err = ErrLocalTagNotInRemote
// 		}
// 	}

// 	return
// }

// Clone retrieves all information from the remote, saving it to the current repository.
func (repo *Repository) Clone() (err error) {
	if repo.config.Origin == "" {
		err = ErrPullNoOrigin
		return
	}

	if tracker.IsTopicURI(repo.config.Origin) {
		err = repo.clone_p2p()
		return
	}

	var origin remote.Origin
	origin, err = remote.Open(repo.config.Origin)
	if err != nil {
		return
	}

	var pull_tags_stage event.NotifyParams
	pull_tags_stage.Stage = event.StagePullTags
	repo.notify(event.NotifyBeginStage, &pull_tags_stage)

	var tags []string
	tags, err = origin.Tags()
	if err != nil {
		return
	}

	var tags_in_queue event.NotifyParams
	tags_in_queue.Count = int64(len(tags))
	repo.notify(event.NotifyTagQueueCount, &tags_in_queue)

	var tagged_commit_objects []cas.ContentID

	for _, tag := range tags {
		var current_tag_commit cas.ContentID
		current_tag_commit, _ = repo.read_tag(tag)

		var remote_tag_commit cas.ContentID
		remote_tag_commit, err = origin.ReadTag(tag)
		if err != nil {
			return
		}

		if err = repo.WriteTag(tag, remote_tag_commit); err != nil {
			return
		}

		var notify_pull_tag event.NotifyParams
		notify_pull_tag.Name1 = tag
		notify_pull_tag.Object1 = current_tag_commit
		notify_pull_tag.Object2 = remote_tag_commit
		repo.notify(event.NotifyPullTag, &notify_pull_tag)

		tagged_commit_objects = append(tagged_commit_objects, remote_tag_commit)
	}

	pull_tags_stage.Success = true
	repo.notify(event.NotifyCompleteStage, &pull_tags_stage)

	err = repo.pull_object_graph(origin, tagged_commit_objects...)

	return
}

// Pull only objects associated with a tag or an abbreviated object hash
func (repo *Repository) Pull(ref ...string) (err error) {
	if repo.config.Origin == "" {
		err = ErrPullNoOrigin
		return
	}

	if tracker.IsTopicURI(repo.config.Origin) {
		err = repo.pull_p2p(ref...)
		return
	}

	origin, err := remote.Open(repo.config.Origin)
	if err != nil {
		return
	}

	objects := make([]cas.ContentID, len(ref))
	for i := range ref {
		objects[i], err = repo.ParseRef(ref[i])
		if err != nil {
			if validate.Hex(ref[i]) {
				if expanded, expansion_err := origin.Deabbreviate(ref[i]); expansion_err == nil {
					objects[i] = expanded
					continue
				}
			}
			return
		}
	}

	err = repo.pull_object_graph(origin, objects...)

	return
}

// PullTags retrieves all tags from the remote origin.
// overwrites any tags already in the repository.
func (repo *Repository) PullTags() (err error) {
	if tracker.IsTopicURI(repo.config.Origin) {
		err = repo.pull_tags_p2p()
		return
	}

	var origin remote.Origin
	origin, err = remote.Open(repo.config.Origin)
	if err != nil {
		return
	}

	// begin to pull tags
	var notify_pull_tags event.NotifyParams
	notify_pull_tags.Stage = event.StagePullTags
	repo.notify(event.NotifyBeginStage, &notify_pull_tags)
	// notify success if err == nil
	defer func() {
		if err == nil {
			notify_pull_tags.Success = true
		}
		repo.notify(event.NotifyCompleteStage, &notify_pull_tags)
	}()

	var tags []string
	tags, err = origin.Tags()
	if err != nil {
		return
	}

	// notify tag count
	var tags_in_queue event.NotifyParams
	tags_in_queue.Count = int64(len(tags))
	repo.notify(event.NotifyTagQueueCount, &tags_in_queue)

	for _, tag := range tags {
		var current_tag_commit cas.ContentID
		current_tag_commit, _ = repo.read_tag(tag)

		var remote_tag_commit cas.ContentID
		remote_tag_commit, err = origin.ReadTag(tag)
		if err != nil {
			return
		}

		if err = repo.WriteTag(tag, remote_tag_commit); err != nil {
			return
		}

		var notify_pull_tag event.NotifyParams
		notify_pull_tag.Name1 = tag
		notify_pull_tag.Object1 = current_tag_commit
		notify_pull_tag.Object2 = remote_tag_commit
		repo.notify(event.NotifyPullTag, &notify_pull_tag)
	}

	return
}

// PullTag retrieves only certain tags from the remote origin
func (repo *Repository) PullTag(tags ...string) (err error) {
	if tracker.IsTopicURI(repo.config.Origin) {
		err = repo.pull_tag_p2p(tags...)
		return
	}

	var origin remote.Origin
	origin, err = remote.Open(repo.config.Origin)
	if err != nil {
		return
	}

	// begin to pull tags
	var notify_pull_tags event.NotifyParams
	notify_pull_tags.Stage = event.StagePullTags
	repo.notify(event.NotifyBeginStage, &notify_pull_tags)
	// notify success if err == nil
	defer func() {
		if err == nil {
			notify_pull_tags.Success = true
		}
		repo.notify(event.NotifyCompleteStage, &notify_pull_tags)
	}()

	// notify tag count
	var tags_in_queue event.NotifyParams
	tags_in_queue.Count = int64(len(tags))
	repo.notify(event.NotifyTagQueueCount, &tags_in_queue)

	for _, tag := range tags {
		var current_tag_commit cas.ContentID
		current_tag_commit, _ = repo.read_tag(tag)

		var remote_tag_commit cas.ContentID
		remote_tag_commit, err = origin.ReadTag(tag)
		if err != nil {
			return
		}

		if err = repo.WriteTag(tag, remote_tag_commit); err != nil {
			return
		}

		var notify_pull_tag event.NotifyParams
		notify_pull_tag.Name1 = tag
		notify_pull_tag.Object1 = current_tag_commit
		notify_pull_tag.Object2 = remote_tag_commit
		repo.notify(event.NotifyPullTag, &notify_pull_tag)
	}

	return
}
