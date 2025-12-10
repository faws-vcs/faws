package cas

import (
	"fmt"
	"os"
)

// Stat tests the existence of an object named by the [ContentID], and returns its size if it does exist.
// If it does not exist, err will be [ErrObjectNotFound].
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
		err = fmt.Errorf("%w: %s", ErrObjectNotFound, id)
		return
	}
	size = fi.Size() - 4
	return
}
