package cas

import (
	"os"
)

// Attempt to store a segment of data, returning the associated ContentID.
func (set *Set) Store(prefix Prefix, data []byte) (new bool, id ContentID, err error) {
	id = hash_content(prefix, data)
	var (
		path string
	)
	path, err = set.store_path(id)
	if err != nil {
		return
	}

	// return true if file is already written (completely)
	fi, stat_err := os.Stat(path)
	if stat_err == nil {
		if fi.Size() == int64(len(data))+4 {
			return
		}
	}

	new = true

	// open file
	var file *os.File
	file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return
	}
	_, err = file.Write(prefix[:])
	if err != nil {
		return
	}
	if _, err = file.Write(data[:]); err != nil {
		return
	}
	err = file.Close()
	if err != nil {
		return
	}

	return
}
