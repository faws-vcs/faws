package repo

import (
	"github.com/faws-vcs/faws/faws/repo/cas"
)

// LoadObject loads an object from the cache
func (repo *Repository) LoadObject(id cas.ContentID) (prefix cas.Prefix, object []byte, err error) {
	prefix, object, err = repo.objects.Load(id)
	return
}

// LoadObject adds an object to the cache
func (repo *Repository) StoreObject(prefix cas.Prefix, data []byte) (new bool, id cas.ContentID, err error) {
	new, id, err = repo.objects.Store(prefix, data)
	return
}

// RemoveObject removes an object from the cache
func (repo *Repository) RemoveObject(id cas.ContentID) (err error) {
	err = repo.objects.Remove(id)
	return
}
