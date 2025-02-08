package repo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/faws-vcs/faws/faws/validate"
)

var (
	ErrTagInvalidCommit = fmt.Errorf("faws/repo: tag points to invalid commit")
)

func (repo *Repository) read_tag(tag string) (commit_hash cas.ContentID, err error) {
	if err = validate.CommitTag(tag); err != nil {
		return
	}

	var tag_file *os.File
	path := filepath.Join(repo.directory, "tags", tag)
	tag_file, err = os.Open(path)
	if err != nil {
		err = fmt.Errorf("%w: tag does not exist", ErrRefNotFound)
		return
	}
	defer tag_file.Close()
	if _, err = io.ReadFull(tag_file, commit_hash[:]); err != nil {
		return
	}

	var commit_info *revision.CommitInfo
	_, commit_info, err = repo.check_commit(commit_hash)
	if err != nil {
		err = ErrTagInvalidCommit
		repo.remove_tag(tag)
	} else if commit_info.Tag != tag {
		repo.remove_tag(tag)
		err = ErrTagInvalidCommit
		return
	}

	return
}

func (repo *Repository) remove_tag(tag string) (err error) {
	err = os.Remove(filepath.Join(repo.directory, "tags", tag))
	return
}

func (repo *Repository) write_tag(tag string, commit_hash cas.ContentID) (err error) {
	if err = validate.CommitTag(tag); err != nil {
		return
	}

	path := filepath.Join(repo.directory, "tags", tag)

	err = os.WriteFile(path, commit_hash[:], fs.DefaultPerm)

	return
}

type Tag struct {
	Name string
	Hash cas.ContentID
}

func (repo *Repository) Tags() (tags []Tag, err error) {
	path := filepath.Join(repo.directory, "tags")
	var items []os.DirEntry
	items, err = os.ReadDir(path)
	if err != nil {
		return
	}
	tags = make([]Tag, 0, len(items))
	for _, item := range items {
		info, info_err := item.Info()
		if info_err == nil {
			if !item.IsDir() && !strings.HasPrefix(item.Name(), ".") && info.Size() == cas.ContentIDSize {
				commit_hash, tag_err := repo.read_tag(item.Name())
				if tag_err == nil {
					tags = append(tags, Tag{
						Name: item.Name(),
						Hash: commit_hash,
					})
				}
			}
		}
	}
	return
}
