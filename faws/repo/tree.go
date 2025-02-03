package repo

import (
	"sort"
	"strings"

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

// search a tree for a path.
func (repo *Repository) find_tree_file(root_tree_hash cas.ContentID, path string) (file *revision.TreeEntry, err error) {
	if path == "" {
		err = ErrBadFilename
		return
	}

	path_components := strings.Split(path, "/")

	root_tree, tree_load_err := repo.load_tree(root_tree_hash)
	if tree_load_err != nil {
		err = tree_load_err
		return
	}

	tree := root_tree

	for _, path_component := range path_components {
		index_of := sort.Search(len(tree.Entries), func(i int) bool {
			return tree.Entries[i].Name >= path_component
		})
		if index_of < len(tree.Entries) && tree.Entries[index_of].Name == path_component {
			file = &tree.Entries[index_of]
			//
			if file.Prefix == cas.Tree {
				tree, err = repo.load_tree(file.Content)
				if err != nil {
					return
				}
			}
		} else {
			err = ErrTreeFileNotFound
			return
		}
	}

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
