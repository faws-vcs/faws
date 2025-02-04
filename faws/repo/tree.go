package repo

import (
	"github.com/faws-vcs/faws/faws/repo/cache"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

// Returns the Tree for a tree or commit object
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

// create a tree object from the current index
func (repo *Repository) WriteTree() (tree_hash cas.ContentID, err error) {
	var cache_tree cache.Tree
	if err = cache_tree.Build(repo.CacheIndex()); err != nil {
		return
	}
	if tree_hash, err = cache_tree.Store(&repo.objects); err != nil {
		return
	}
	return
}
