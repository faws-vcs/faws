package validate

import (
	"fmt"
	"strings"
)

const (
	illegal_description_characters = "\x1B"
	max_description_length         = 1024
)

var (
	ErrDescriptionTooLong           = fmt.Errorf("faws/validate: description is too long")
	ErrDescriptionInvalidCharacters = fmt.Errorf("faws/validate: description contains invalid characters")
)

// Description returns an error if description is invalid
//
// If description is too long, err = ErrDescriptionTooLong
// If description contains illegal characters, err = ErrDescriptionInvalidCharacters
func Description(description string) (err error) {
	if len(description) > max_description_length {
		err = ErrDescriptionTooLong
		return
	}
	if strings.ContainsAny(description, illegal_description_characters) {
		err = ErrDescriptionInvalidCharacters
		return
	}
	return
}
