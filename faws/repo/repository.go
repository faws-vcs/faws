package repo

import (
	"path/filepath"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/config"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
	"github.com/google/uuid"
)

type Repository struct {
	// describes what mode the repository operates under
	config config.Config
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
	// the URL of the tracker server
	tracker_url string
}

type Option func(*Repository)

// Open opens an existing Faws repository
func (repo *Repository) Open(directory string, options ...Option) (err error) {
	if !Exists(directory) {
		err = ErrRepoNotExist
		return
	}

	// ignore notifications by default
	repo.notify = dont_care

	// use official tracker server by default
	repo.tracker_url = tracker.DefaultURL

	// set directory
	repo.directory = directory

	for _, o := range options {
		o(repo)
	}

	if err = repo.lock(); err != nil {
		return
	}

	if err = config.ReadConfig(filepath.Join(repo.directory, "config"), &repo.config); err != nil {
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

func (repo *Repository) UUID() uuid.UUID {
	return repo.config.UUID
}

// Close saves changes made to the repository and releases its resources
func (repo *Repository) Close() (err error) {
	if err = repo.write_index(); err != nil {
		return
	}

	err = repo.unlock()
	return
}
