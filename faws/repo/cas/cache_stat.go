package cas

import "os"

func (cache *cache) Stat(id ContentID) (size int64, err error) {
	var fi os.FileInfo
	fi, err = os.Stat(cache.path(id))
	if err != nil {
		err = object_error{ErrObjectNotFound, id}
		return
	}
	size = fi.Size() - 4
	return
}
