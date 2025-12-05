package cas

import (
	"bytes"
	"os"

	"github.com/faws-vcs/faws/faws/fs"
)

// Store attempts to store data, returning the associated ContentID.
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
			var (
				current_prefix Prefix
				current_data   []byte
			)
			current_data, err = os.ReadFile(path)
			if err != nil {
				return
			}
			copy(current_prefix[:], current_data[:4])
			current_data = current_data[4:]
			if current_prefix == prefix {
				if bytes.Equal(current_data, data) {
					return
				}
			}
		}
	}

	new = true

	// open file
	var file *os.File
	file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fs.DefaultPublicPerm)
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
