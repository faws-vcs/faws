package repo

import (
	"github.com/faws-vcs/faws/faws/repo/cas"
)

func (repo *Repository) LoadObject(id cas.ContentID) (prefix cas.Prefix, object []byte, err error) {
	prefix, object, err = repo.objects.Load(id)
	return
}

func (repo *Repository) StoreObject(prefix cas.Prefix, data []byte) (new bool, id cas.ContentID, err error) {
	new, id, err = repo.objects.Store(prefix, data)
	return
}

func (repo *Repository) RemoveObject(id cas.ContentID) (err error) {
	err = repo.objects.Remove(id)
	return
}
