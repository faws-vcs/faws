package staging

import (
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

var (
	ErrTreeInvalidPrefix        = fmt.Errorf("faws/cache: hierarchy is malformed, tree object has invalid prefix")
	ErrTreeLacksContent         = fmt.Errorf("faws/cache: tree does not have any content hash")
	ErrTreePathNotFound         = fmt.Errorf("faws/cache: tree does not contain path")
	ErrTreeCannotRemoveNonempty = fmt.Errorf("faws/cache: cannot remove non-empty directory")
	ErrTreeEmpty                = fmt.Errorf("faws/cache: tree has no entries")
)

// File is a temporary structure used when converting the index into trees
type File struct {
	Prefix cas.Prefix
	Name   string
	Mode   revision.FileMode
	Tree   *Tree
	File   cas.ContentID
}

// Tree is a temporary structure used when converting the index into trees
// It is useful since the index format is flat and non-hierarchical, which repository trees must be
type Tree struct {
	Files []File
}

// Store recursively converts the temporary [Tree] structure into actual [revision.Tree] objects
func (cache_tree *Tree) Store(objects *cas.Set) (root cas.ContentID, err error) {
	var revision_tree revision.Tree
	revision_tree.Entries = make([]revision.TreeEntry, 0, len(cache_tree.Files))

	for i := range cache_tree.Files {
		cache_file := &cache_tree.Files[i]

		var revision_file revision.TreeEntry
		revision_file.Prefix = cache_file.Prefix
		revision_file.Name = cache_file.Name
		revision_file.Mode = cache_file.Mode

		if cache_file.Prefix == cas.Tree {
			// store subdirectory as a tree
			revision_file_tree, revision_file_err := cache_file.Tree.Store(objects)
			if revision_file_err != nil {
				if errors.Is(revision_file_err, ErrTreeEmpty) {
					continue
				}
				err = revision_file_err
				return
			}
			revision_file.Content = revision_file_tree
			revision_tree.Entries = append(revision_tree.Entries, revision_file)
		} else if cache_file.Prefix == cas.File {
			// TODO: write content
			revision_file.Content = cache_file.File
			revision_tree.Entries = append(revision_tree.Entries, revision_file)
		}
	}

	if len(revision_tree.Entries) == 0 {
		err = ErrTreeEmpty
		return
	}

	tree_data, tree_marshal_err := revision.MarshalTree(&revision_tree)
	if tree_marshal_err != nil {
		err = tree_marshal_err
		return
	}
	_, root, err = objects.Store(cas.Tree, tree_data)
	if err != nil {
		return
	}

	return
}

// converts flat index entries into a hierarchical temporary structure
func (tree *Tree) build_index_entry(linked_file *IndexEntry) (err error) {
	var (
		current_tree *Tree = tree
	)

	path_components := strings.Split(linked_file.Path, "/")
	containing_path_components := path_components[:len(path_components)-1]
	filename := path_components[len(path_components)-1]

	for _, path_component := range containing_path_components {
		// find the file in the current directory hierarchy
		entry_index := sort.Search(len(current_tree.Files), func(i int) bool {
			return current_tree.Files[i].Name >= path_component
		})
		if entry_index < len(current_tree.Files) && current_tree.Files[entry_index].Name == path_component {
			// found
			file := &current_tree.Files[entry_index]
			if file.Tree != nil {
				current_tree = file.Tree
			} else {
				// file in directory path
				err = ErrTreePathNotFound
				return
			}
		} else {
			// directory not found, create now
			var directory_file File
			directory_file.Prefix = cas.Tree
			directory_file.Name = path_component
			directory_file.Mode = 0
			directory_file.Tree = new(Tree)
			current_tree.Files = slices.Insert(current_tree.Files, entry_index, directory_file)
			current_tree = directory_file.Tree
		}
	}

	entry_index := sort.Search(len(current_tree.Files), func(i int) bool {
		return current_tree.Files[i].Name >= filename
	})
	if entry_index < len(current_tree.Files) && current_tree.Files[entry_index].Name == filename {
		// found
		file := &current_tree.Files[entry_index]
		file.Prefix = cas.File
		file.Name = filename
		file.Mode = linked_file.Mode
		file.File = linked_file.File

		// err = ErrTreeFileAlreadyLinked
	} else {
		var file File
		file.Prefix = cas.File
		file.Name = filename
		file.Mode = linked_file.Mode
		file.File = linked_file.File
		current_tree.Files = slices.Insert(current_tree.Files, entry_index, file)
	}
	return
}

// Build organizes the flat [Index] entries into a hierarchical [Tree] structure
func (tree *Tree) Build(index *Index) (err error) {
	for i := range index.Entries {
		if err = tree.build_index_entry(&index.Entries[i]); err != nil {
			return
		}
	}

	return
}
