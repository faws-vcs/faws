package cas

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/faws-vcs/faws/faws/validate"
)

// Deabbreviate expands a hash abbreviation string. It will attempt to disambiguate even the shortest possible hexadecimal string.
// If the abbreviation is not hexadecimal, err will be [ErrAbbreviationNotHex]
// If multiple candidates exist for an abbreviation, err will be [ErrAbbreviationAmbiguous]
// If the abbreviation does not fit with any object in the Set, err will be [ErrObjectNotFound]
func (set *Set) Deabbreviate(abbreviation string) (hash ContentID, err error) {
	abbreviation = strings.ToLower(abbreviation)
	if len(abbreviation) == 0 {
		err = ErrAbbreviationTooShort
		return
	}
	if !validate.Hex(abbreviation) {
		err = ErrAbbreviationNotHex
		return
	}

	if len(abbreviation) == 1 {
		var candidate string
		var dir_entries []os.DirEntry
		dir_entries, err = os.ReadDir(set.directory)
		if err != nil {
			return
		}
		for _, dir_entry := range dir_entries {
			if validate.Hex(dir_entry.Name()) && len(dir_entry.Name()) == 2 {
				if strings.HasPrefix(dir_entry.Name(), abbreviation) {
					if candidate != "" {
						err = ErrAbbreviationAmbiguous
						return
					}
					candidate = abbreviation
				}
			}
		}
		if candidate == "" {
			err = ErrObjectNotFound
			return
		}
		abbreviation = candidate
	}
	if len(abbreviation) == 2 {
		bucket1_path := filepath.Join(set.directory, abbreviation)
		var dir_entries []os.DirEntry
		dir_entries, err = os.ReadDir(bucket1_path)
		if err != nil {
			return
		}
		var candidate string
		for _, dir_entry := range dir_entries {
			if validate.Hex(dir_entry.Name()) && len(dir_entry.Name()) == 2 {
				if candidate != "" {
					err = ErrAbbreviationAmbiguous
					return
				}
				candidate = abbreviation + dir_entry.Name()
			}
		}
		if candidate == "" {
			err = ErrObjectNotFound
			return
		}
		abbreviation = candidate
	}
	if len(abbreviation) == 3 {
		bucket1_path := filepath.Join(set.directory, abbreviation[0:2])
		var dir_entries []os.DirEntry
		dir_entries, err = os.ReadDir(bucket1_path)
		if err != nil {
			return
		}
		var candidate string
		for _, dir_entry := range dir_entries {
			if validate.Hex(dir_entry.Name()) && len(dir_entry.Name()) == 2 && dir_entry.Name()[0] == abbreviation[2] {
				if candidate != "" {
					err = ErrAbbreviationAmbiguous
					return
				}
				candidate = abbreviation + dir_entry.Name()[1:]
			}
		}
		if candidate == "" {
			err = ErrObjectNotFound
			return
		}
		abbreviation = candidate
	}
	if len(abbreviation) == 4 {
		bucket2_path := filepath.Join(set.directory, abbreviation[0:2], abbreviation[2:4])
		var dir_entries []os.DirEntry
		dir_entries, err = os.ReadDir(bucket2_path)
		if err != nil {
			return
		}
		var candidate string
		for _, dir_entry := range dir_entries {
			if validate.Hex(dir_entry.Name()) && len(dir_entry.Name()) == (ContentIDSize-2)*2 {
				if candidate != "" {
					err = ErrAbbreviationAmbiguous
					return
				}
				candidate = abbreviation + dir_entry.Name()
			}
		}
		abbreviation = candidate
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
		err = fmt.Errorf("%w: could not deabbreviate '%s'", ErrObjectNotFound, abbreviation)
		return
	}

	_, err = hex.Decode(hash[2:], []byte(found_name))
	return
}
