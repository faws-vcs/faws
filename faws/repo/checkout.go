package repo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/faws-vcs/faws/faws/validate"
)

type CheckoutMode uint8

const (
	CheckoutOverwrite = 1 << iota
	CheckoutCasefoldTree
)

var (
	ErrCheckoutBadPrefix = fmt.Errorf("faws/repo: bad prefix")
	ErrCheckoutOverwrite = fmt.Errorf("faws/repo: a file exists at the destination. pass -w, --overwrite to write anyway")
)

func (repo *Repository) checkout_file(file_hash cas.ContentID, mode revision.FileMode, dest string, checkout_mode CheckoutMode) (err error) {
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

	if checkout_mode&CheckoutOverwrite == 0 {
		if _, err = os.Stat(dest); err == nil {
			err = ErrCheckoutOverwrite
			return
		}
	}

	var file_size int64
	var part_size int64
	// compute file size
	part_hashes := file_data
	var part_hash cas.ContentID
	for len(part_hashes) > 0 {
		// reach each part hash
		copy(part_hash[:], part_hashes[:cas.ContentIDSize])
		part_hashes = part_hashes[cas.ContentIDSize:]
		part_size, err = repo.objects.Stat(part_hash)
		if err != nil {
			return
		}
		file_size += part_size
	}

	var notify_checkout event.NotifyParams
	notify_checkout.Name1 = dest
	notify_checkout.Count = file_size
	repo.notify(event.NotifyCheckoutFile, &notify_checkout)

	var file *os.File
	file, err = os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_RDWR, perm)
	if err != nil {
		return
	}
	// allocate file size
	if _, err = file.Seek(file_size, io.SeekStart); err != nil {
		return
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return
	}

	var (
		part_prefix cas.Prefix
		part_data   []byte
	)
	part_hashes = file_data
	for len(part_hashes) > 0 {
		copy(part_hash[:], part_hashes[:cas.ContentIDSize])
		part_hashes = part_hashes[cas.ContentIDSize:]
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

		var notify_checkout_file_part event.NotifyParams
		notify_checkout_file_part.Count = int64(len(part_data))
		repo.notify(event.NotifyCheckoutFilePart, &notify_checkout_file_part)
	}

	err = file.Close()

	return
}

func (repo *Repository) checkout_tree(tree_hash cas.ContentID, dest string, checkout_mode CheckoutMode) (err error) {
	var tree *revision.Tree
	tree, err = repo.load_tree(tree_hash)
	if err != nil {
		return
	}

	err = os.Mkdir(dest, fs.DefaultPublicDirPerm)
	if err != nil && !os.IsExist(err) {
		return
	}

	for _, tree_entry := range tree.Entries {
		if err = validate.FileName(tree_entry.Name); err != nil {
			return
		}

		tree_entry_name := tree_entry.Name
		if checkout_mode&CheckoutCasefoldTree != 0 {
			tree_entry_name = strings.ToLower(tree_entry_name)
		}
		tree_entry_dest := filepath.Join(dest, tree_entry_name)

		switch tree_entry.Prefix {
		case cas.Tree:
			if err = repo.checkout_tree(tree_entry.Content, tree_entry_dest, checkout_mode); err != nil {
				return
			}
		case cas.File:
			if err = repo.checkout_file(tree_entry.Content, tree_entry.Mode, tree_entry_dest, checkout_mode); err != nil {
				return
			}
		default:
			err = ErrCheckoutBadPrefix
			return
		}
	}

	return
}

func (repo *Repository) checkout_commit(commit_hash cas.ContentID, dest string, checkout_mode CheckoutMode) (err error) {
	var commit_info *revision.CommitInfo
	_, commit_info, err = repo.check_commit(commit_hash)
	if err != nil {
		return
	}

	return repo.checkout_tree(commit_info.Tree, dest, checkout_mode)
}

// Checkout exports an object (most commonly, a commit) to a destination on the host filesystem.
//
// If overwrite == true, existing files in the path are destroyed and no error is returned.
func (repo *Repository) Checkout(object_hash cas.ContentID, dest string, mode CheckoutMode) (err error) {
	var checkout_stage event.NotifyParams
	checkout_stage.Stage = event.StageCheckout
	repo.notify(event.NotifyBeginStage, &checkout_stage)

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
		err = repo.checkout_commit(object_hash, dest, mode)
	case cas.Tree:
		err = repo.checkout_tree(object_hash, dest, mode)
	case cas.File:
		err = repo.checkout_file(object_hash, 0, dest, mode)
	default:
		err = ErrCheckoutBadPrefix
	}

	checkout_stage.Success = err == nil
	repo.notify(event.NotifyCompleteStage, &checkout_stage)

	return
}
