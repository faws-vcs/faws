package cas

import (
	"encoding/hex"
	"os"
	"path/filepath"
)

// Remove removes the object by its [ContentID]
func (set *Set) Remove(id ContentID) (err error) {
	// id = AABBCCDDEEFF...
	s := hex.EncodeToString(id[:])

	// cas_set/AA
	prefix1 := filepath.Join(set.directory, s[0:2])
	// cas_set/AA/BB
	prefix2 := filepath.Join(prefix1, s[2:4])

	// rm cas_set/AA/BB/CCDDEEFF...
	err = os.Remove(filepath.Join(prefix2, s[4:]))
	if err != nil {
		return
	}

	// remove empty directories
	var entries []os.DirEntry
	// ls cas_set/AA/BB
	entries, err = os.ReadDir(prefix2)
	if err != nil {
		return
	}

	if len(entries) == 0 {
		// rmdir cas_set/AA/BB
		if err = os.Remove(prefix2); err != nil {
			return
		}

		// ls cas_set/AA
		entries, err = os.ReadDir(prefix1)
		if err != nil {
			return
		}

		if len(entries) == 0 {
			// rmdir cas_set/AA
			err = os.Remove(prefix1)
		}
	}

	return
}
