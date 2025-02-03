package repo

import "github.com/faws-vcs/faws/faws/repo/cas"

func (repo *Repository) Object(id cas.ContentID) (prefix cas.Prefix, object []byte, err error) {
	prefix, object, err = repo.objects.Load(id)
	if err != nil {
		return
	}
	return
}
