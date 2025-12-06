package repo

import (
	"github.com/faws-vcs/faws/faws/repo/cache"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

// Tree returns the tree identified by a tree hash, or if object_hash refers to a commit, returns the tree associated with that commit
func (repo *Repository) Tree(object_hash cas.ContentID) (tree *revision.Tree, err error) {
	var (
		object []byte
		prefix cas.Prefix
	)
	prefix, object, err = repo.objects.Load(object_hash)
	if err != nil {
		return
	}
	switch prefix {
	case cas.Commit:
		var commit_info *revision.CommitInfo
		_, commit_info, err = repo.check_commit(object_hash)
		if err != nil {
			return
		}
		tree, err = repo.load_tree(commit_info.Tree)
		return
	case cas.Tree:
		tree = new(revision.Tree)
		err = revision.UnmarshalTree(object, tree)
		return
	default:
		err = ErrTreeInvalidPrefix
		return
	}
}

func (repo *Repository) load_tree(tree_hash cas.ContentID) (tree *revision.Tree, err error) {
	var (
		tree_data []byte
		prefix    cas.Prefix
	)
	prefix, tree_data, err = repo.objects.Load(tree_hash)
	if err != nil {
		return
	}
	if prefix != cas.Tree {
		err = ErrTreeInvalidPrefix
		return
	}

	tree = new(revision.Tree)
	err = revision.UnmarshalTree(tree_data, tree)
	return
}

// Write creates a tree object using the current index
func (repo *Repository) WriteTree() (tree_hash cas.ContentID, err error) {
	var notify_params event.NotifyParams
	notify_params.Stage = event.StageWriteTree
	repo.notify(event.NotifyBeginStage, &notify_params)

	var cache_tree cache.Tree
	if err = cache_tree.Build(repo.CacheIndex()); err != nil {
		repo.notify(event.NotifyCompleteStage, &notify_params)
		return
	}

	tree_hash, err = cache_tree.Store(&repo.objects)
	notify_params.Success = err == nil
	repo.notify(event.NotifyCompleteStage, &notify_params)

	// clear cached objects
	repo.index.cache_objects = make(map[cas.ContentID]uint32)

	return
}
