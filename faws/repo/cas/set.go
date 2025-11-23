package cas

import (
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/faws-vcs/faws/faws/fs"
)

type Set struct {
	// the location of the cas.Set
	directory string
}

func (set *Set) path(id ContentID) (path string, err error) {
	hex_id := hex.EncodeToString(id[:])
	prefix := filepath.Join(set.directory, hex_id[0:2], hex_id[2:4])
	if err != nil {
		return
	}

	path = filepath.Join(prefix, string(hex_id[4:]))
	return
}

func (set *Set) store_path(id ContentID) (path string, err error) {
	hex_id := hex.EncodeToString(id[:])
	prefix := filepath.Join(set.directory, hex_id[0:2], hex_id[2:4])
	if err != nil {
		return
	}
	err = os.MkdirAll(prefix, fs.DefaultPublicDirPerm)
	if err != nil {
		return
	}

	path = filepath.Join(prefix, string(hex_id[4:]))
	return
}
