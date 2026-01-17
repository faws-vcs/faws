package cas

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidSet      = errors.New("faws/repo/cas: set is not a directory")
	ErrObjectCorrupted = errors.New("faws/cas: object content ID does not match content on disk")

	ErrAbbreviationAmbiguous = fmt.Errorf("faws/repo/cas: abbreviation is ambiguous")
	ErrAbbreviationTooShort  = fmt.Errorf("faws/repo/cas: abbreviation is too short to be expanded")
	ErrAbbreviationNotHex    = fmt.Errorf("faws/repo/cas: abbreviation is not hexadecimal")

	ErrObjectNotFound = fmt.Errorf("faws/repo/cas: object not found")
)

type object_error struct {
	err error
	id  ContentID
}

func (object_error object_error) Error() string {
	return object_error.err.Error() + ": " + object_error.id.String()
}

func (object_error object_error) Unwrap() error {
	return object_error.err
}
