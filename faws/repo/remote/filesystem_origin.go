package remote

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/config"
	"github.com/faws-vcs/faws/faws/validate"
	"github.com/google/uuid"
)

// filesystem_origin wraps a Fs to implement an Origin
type filesystem_origin struct {
	filesystem filesystem
}

func (fs_origin filesystem_origin) URI() (uri string) {
	uri = fs_origin.filesystem.URI()
	return
}

func (fs_origin filesystem_origin) UUID() (id uuid.UUID, err error) {
	var (
		repo_config config.Config
		config_file io.ReadCloser
	)
	config_file, err = fs_origin.filesystem.Pull("config")
	if err != nil {
		return
	}
	decoder := json.NewDecoder(config_file)
	err = decoder.Decode(&repo_config)
	if err != nil {
		return
	}
	config_file.Close()
	id = repo_config.UUID
	return
}

func (fs_origin filesystem_origin) Tags() (tags []string, err error) {
	var (
		entries []dir_entry
	)
	entries, err = fs_origin.filesystem.ReadDir("tags")
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir {
			if validate.CommitTag(entry.Name) == nil {
				tags = append(tags, entry.Name)
			}
		}
	}

	return
}

func (fs_origin filesystem_origin) ReadTag(name string) (commit_hash cas.ContentID, err error) {
	if err = validate.CommitTag(name); err != nil {
		return
	}
	var (
		file io.ReadCloser
	)
	file, err = fs_origin.filesystem.Pull("tags/" + name)
	if err != nil {
		return
	}
	if _, err = io.ReadFull(file, commit_hash[:]); err != nil {
		return
	}
	file.Close()

	return
}

func (fs_origin filesystem_origin) GetObject(object_hash cas.ContentID) (prefix cas.Prefix, content []byte, err error) {
	hex_name := object_hash.String()
	name := "objects/" + hex_name[0:2] + "/" + hex_name[2:4] + "/" + hex_name[4:]

	var (
		size int64
		file io.ReadCloser
	)
	size, err = fs_origin.filesystem.Stat(name)
	if err != nil {
		return
	}
	if size < 4 {
		err = cas.ErrObjectCorrupted
		return
	}
	content = make([]byte, size-4)

	file, err = fs_origin.filesystem.Pull(name)
	if err != nil {
		return
	}

	if _, err = io.ReadFull(file, prefix[:]); err != nil {
		return
	}
	switch prefix {
	case cas.Commit:
	case cas.Tree:
	case cas.File:
	case cas.Part:
	default:
		err = cas.ErrObjectCorrupted
		return
	}

	if _, err = io.ReadFull(file, content); err != nil {
		return
	}
	file.Close()
	return
}

func (fs_origin filesystem_origin) Deabbreviate(abbreviation string) (hash cas.ContentID, err error) {
	abbreviation = strings.ToLower(abbreviation)
	if len(abbreviation) == 0 {
		err = cas.ErrAbbreviationTooShort
		return
	}
	if !validate.Hex(abbreviation) {
		err = cas.ErrAbbreviationNotHex
		return
	}

	if len(abbreviation) == 1 {
		var candidate string
		var dir_entries []dir_entry
		dir_entries, err = fs_origin.filesystem.ReadDir("objects")
		if err != nil {
			return
		}
		for _, dir_entry := range dir_entries {
			if validate.Hex(dir_entry.Name) && len(dir_entry.Name) == 2 {
				if strings.HasPrefix(dir_entry.Name, abbreviation) {
					if candidate != "" {
						err = cas.ErrAbbreviationAmbiguous
						return
					}
					candidate = abbreviation
				}
			}
		}
		if candidate == "" {
			err = cas.ErrObjectNotFound
			return
		}
		abbreviation = candidate
	}
	if len(abbreviation) == 2 {
		bucket1_path := strings.Join([]string{"objects", abbreviation}, "/")
		var dir_entries []dir_entry
		dir_entries, err = fs_origin.filesystem.ReadDir(bucket1_path)
		if err != nil {
			return
		}
		var candidate string
		for _, dir_entry := range dir_entries {
			if validate.Hex(dir_entry.Name) && len(dir_entry.Name) == 2 {
				if candidate != "" {
					err = cas.ErrAbbreviationAmbiguous
					return
				}
				candidate = abbreviation + dir_entry.Name
			}
		}
		if candidate == "" {
			err = cas.ErrObjectNotFound
			return
		}
		abbreviation = candidate
	}
	if len(abbreviation) == 3 {
		bucket1_path := strings.Join([]string{"objects", abbreviation[0:2]}, "/")
		var dir_entries []dir_entry
		dir_entries, err = fs_origin.filesystem.ReadDir(bucket1_path)
		if err != nil {
			return
		}
		var candidate string
		for _, dir_entry := range dir_entries {
			if validate.Hex(dir_entry.Name) && len(dir_entry.Name) == 2 && dir_entry.Name[0] == abbreviation[2] {
				if candidate != "" {
					err = cas.ErrAbbreviationAmbiguous
					return
				}
				candidate = abbreviation + dir_entry.Name[1:]
			}
		}
		if candidate == "" {
			err = cas.ErrObjectNotFound
			return
		}
		abbreviation = candidate
	}
	if len(abbreviation) == 4 {
		bucket2_path := strings.Join([]string{"objects", abbreviation[0:2], abbreviation[2:4]}, "/")
		var dir_entries []dir_entry
		dir_entries, err = fs_origin.filesystem.ReadDir(bucket2_path)
		if err != nil {
			return
		}
		var candidate string
		for _, dir_entry := range dir_entries {
			if validate.Hex(dir_entry.Name) && len(dir_entry.Name) == (cas.ContentIDSize-2)*2 {
				if candidate != "" {
					err = cas.ErrAbbreviationAmbiguous
					return
				}
				candidate = abbreviation + dir_entry.Name
			}
		}
		abbreviation = candidate
	}

	if _, err = hex.Decode(hash[0:2], []byte(abbreviation)[:4]); err != nil {
		return
	}

	lookup_part := abbreviation[4:]

	bucket := strings.Join([]string{"objects", abbreviation[0:2], abbreviation[2:4]}, "/")

	var dir_entries []dir_entry
	dir_entries, err = fs_origin.filesystem.ReadDir(bucket)
	if err != nil {
		return
	}

	var found_name string

	for _, entry := range dir_entries {
		if strings.HasPrefix(entry.Name, lookup_part) {
			if found_name != "" {
				err = cas.ErrAbbreviationAmbiguous
				return
			}
			found_name = entry.Name
		}
	}

	if found_name == "" {
		err = fmt.Errorf("%w: could not deabbreviate '%s'", cas.ErrObjectNotFound, abbreviation)
		return
	}

	_, err = hex.Decode(hash[2:], []byte(found_name))
	return
}
