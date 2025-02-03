package cas

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
)

func (set *Set) Deabbreviate(abbreviation string) (hash ContentID, err error) {
	if len(abbreviation) < 5 {
		err = ErrAbbreviationTooShort
		return
	}

	if _, err = hex.Decode(hash[0:2], []byte(abbreviation)[:4]); err != nil {
		return
	}

	lookup_part := abbreviation[4:]

	bucket := filepath.Join(set.directory, abbreviation[0:2], abbreviation[2:4])

	var dir_entries []os.DirEntry
	dir_entries, err = os.ReadDir(bucket)
	if err != nil {
		return
	}

	var found_name string

	for _, entry := range dir_entries {
		if strings.HasPrefix(entry.Name(), lookup_part) {
			if found_name != "" {
				err = ErrAbbreviationAmbiguous
				return
			}
			found_name = entry.Name()
		}
	}

	if found_name == "" {
		err = ErrObjectNotFound
		return
	}

	_, err = hex.Decode(hash[2:], []byte(found_name))
	return
}
