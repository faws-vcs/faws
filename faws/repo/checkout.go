package repo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/faws-vcs/faws/faws/validate"
)

var (
	ErrCheckoutBadPrefix = fmt.Errorf("faws/repo: bad prefix")
	ErrCheckoutOverwrite = fmt.Errorf("faws/repo: a file exists at the destination. pass -w, --overwrite to write anyway")
)

func (repo *Repository) checkout_file(file_hash cas.ContentID, mode revision.FileMode, dest string, overwrite bool) (err error) {
	var (
		file_data   []byte
		file_prefix cas.Prefix
	)
	file_prefix, file_data, err = repo.objects.Load(file_hash)
	if err != nil {
		return
	}
	if file_prefix != cas.File {
		err = ErrCheckoutBadPrefix
		return
	}

	var perm os.FileMode = 0700
	if mode&revision.FileModeExecutable != 0 {
		perm |= 0111
	}

	if !overwrite {
		if _, err = os.Stat(dest); err == nil {
			err = ErrCheckoutOverwrite
			return
		}
	}

	var file *os.File
	file, err = os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_RDWR, perm)
	if err != nil {
		return
	}

	var (
		part_hash   cas.ContentID
		part_prefix cas.Prefix
		part_data   []byte
	)
	for len(file_data) > 0 {
		copy(part_hash[:], file_data[:cas.ContentIDSize])
		part_prefix, part_data, err = repo.objects.Load(part_hash)
		if err != nil {
			return
		}
		if part_prefix != cas.Part {
			err = ErrCheckoutBadPrefix
			return
		}

		_, err = file.Write(part_data)
		if err != nil {
			return
		}

		file_data = file_data[cas.ContentIDSize:]
	}

	err = file.Close()
	return
}

func (repo *Repository) checkout_tree(tree_hash cas.ContentID, dest string, overwrite bool) (err error) {
	var tree *revision.Tree
	tree, err = repo.load_tree(tree_hash)
	if err != nil {
		return
	}

	err = os.Mkdir(dest, fs.DefaultPerm)
	if err != nil && !os.IsExist(err) {
		return
	}

	for _, tree_entry := range tree.Entries {
		if err = validate.FileName(tree_entry.Name); err != nil {
			return
		}

		switch tree_entry.Prefix {
		case cas.Tree:
			if err = repo.checkout_tree(tree_entry.Content, filepath.Join(dest, tree_entry.Name), overwrite); err != nil {
				return
			}
		case cas.File:
			if err = repo.checkout_file(tree_entry.Content, tree_entry.Mode, filepath.Join(dest, tree_entry.Name), overwrite); err != nil {
				return
			}
		default:
			err = ErrCheckoutBadPrefix
			return
		}
	}

	return
}

func (repo *Repository) checkout_commit(commit_hash cas.ContentID, dest string, overwrite bool) (err error) {
	var commit_info *revision.CommitInfo
	_, commit_info, err = repo.check_commit(commit_hash)
	if err != nil {
		return
	}

	return repo.checkout_tree(commit_info.Tree, dest, overwrite)
}

func (repo *Repository) Checkout(object_hash cas.ContentID, dest string, overwrite bool) (err error) {
	var (
		prefix cas.Prefix
		// data   []byte
	)
	prefix, _, err = repo.objects.Load(object_hash)
	if err != nil {
		return
	}

	switch prefix {
	case cas.Commit:
		err = repo.checkout_commit(object_hash, dest, overwrite)
	case cas.Tree:
		err = repo.checkout_tree(object_hash, dest, overwrite)
	case cas.File:
		err = repo.checkout_file(object_hash, 0, dest, overwrite)
	default:
		err = ErrCheckoutBadPrefix
	}
	return
}
