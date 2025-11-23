package repo

import (
	"path/filepath"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
)

type Repository struct {
	// describes what mode the repository operates under
	config Config
	// the mechanism for detecting new authors and verifying their identities
	trust Trust
	// called to notify the user of certain events
	notify event.NotifyFunc
	// the directory where the repoistory is located
	directory string
	// cas objects that belong to the repository itself
	objects cas.Set
	// cached changes
	index cache_index
}

type Option func(*Repository)

// Open an existing Faws repository
func (repo *Repository) Open(directory string, options ...Option) (err error) {
	if !Exists(directory) {
		err = ErrRepoNotExist
		return
	}

	// ignore notifications by default
	repo.notify = dont_care

	// set directory
	repo.directory = directory

	for _, o := range options {
		o(repo)
	}

	if err = repo.lock(); err != nil {
		return
	}

	if err = ReadConfig(filepath.Join(repo.directory, "config"), &repo.config); err != nil {
		return
	}

	// open main cas
	if err = repo.objects.Open(filepath.Join(repo.directory, "objects")); err != nil {
		return
	}

	// open index
	if err = repo.read_index(); err != nil {
		return
	}

	return
}

func (repo *Repository) Close() (err error) {
	if err = repo.write_index(); err != nil {
		return
	}

	err = repo.unlock()
	return
}
