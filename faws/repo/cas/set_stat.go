package cas

import "os"

func (set *Set) Stat(id ContentID) (size int64, err error) {
	var (
		path string
		fi   os.FileInfo
	)
	path, err = set.path(id)
	if err != nil {
		return
	}

	fi, err = os.Stat(path)
	if err != nil {
		err = ErrObjectNotFound
		return
	}
	size = fi.Size()
	return
}
