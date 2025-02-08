package validate

import (
	"fmt"
	"strings"
)

const illegal_commit_tag_characters = ` */:<>?\\|`

var (
	ErrCommitTagTooBig            = fmt.Errorf("faws/validate: commit tag is too long")
	ErrCommitTagInvalidCharacters = fmt.Errorf("faws/validate: commit contains illegal characters '%s'", illegal_commit_tag_characters)
	ErrCommitTagCannotBeEmpty     = fmt.Errorf("faws/validate: commit tag cannot be empty")
)

func CommitTag(tag string) (err error) {
	if tag == "" {
		err = ErrCommitTagCannotBeEmpty
		return
	}

	if len(tag) > 120 {
		err = ErrCommitTagTooBig
		return
	}
	if strings.ContainsAny(tag, illegal_commit_tag_characters) {
		err = ErrCommitTagInvalidCharacters
		return
	}

	return
}
