package validate

import (
	"fmt"
	"strings"
)

var (
	ErrCommitTagTooBig            = fmt.Errorf("faws/validate: commit tag is too long")
	ErrCommitTagInvalidCharacters = fmt.Errorf("faws/validate: commit contains illegal characters")
	ErrCommitTagCannotBeEmpty     = fmt.Errorf("faws/validate: commit tag cannot be empty")
)

// CommitTag returns an error if tag is invalid
//
// If the tag is empty, err = [ErrCommitTagCannotBeEmpty]
// If the tag is too long, err = [ErrCommitTagTooBig]
// If the tag contains illegal characters, err = [ErrCommitTagInvalidCharacters]
func CommitTag(tag string) (err error) {
	if tag == "" {
		err = ErrCommitTagCannotBeEmpty
		return
	}

	if len(tag) > 120 {
		err = ErrCommitTagTooBig
		return
	}

	if strings.ContainsFunc(tag, is_invalid_tag_character) {
		err = ErrCommitTagInvalidCharacters
		return
	}

	return
}
