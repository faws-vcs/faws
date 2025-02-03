package cas

import (
	"encoding/hex"
	"os"
	"path/filepath"
)

func (set *Set) Delete(id ContentID) (err error) {
	s := hex.EncodeToString(id[:])

	prefix1 := filepath.Join(set.directory, s[0:2])
	prefix2 := filepath.Join(prefix1, s[2:4])

	err = os.Remove(filepath.Join(prefix2, s[4:]))
	if err != nil {
		return
	}

	// remove empty directories
	var entries []os.DirEntry
	entries, err = os.ReadDir(prefix2)
	if err != nil {
		return
	}

	if len(entries) == 0 {
		if err = os.Remove(prefix2); err != nil {
			return
		}

		entries, err = os.ReadDir(prefix1)
		if err != nil {
			return
		}

		if len(entries) == 0 {
			err = os.Remove(prefix1)
		}
	}

	return
}
