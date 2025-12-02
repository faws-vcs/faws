package cas

import (
	"errors"
	"fmt"
)

var (
	ErrObjectCorrupted = errors.New("faws/cas: object content ID does not match content on disk")

	ErrAbbreviationAmbiguous = fmt.Errorf("faws/repo/cas: abbreviation is ambiguous")
	ErrAbbreviationTooShort  = fmt.Errorf("faws/repo/cas: abbreviation is too short to be expanded")
	ErrAbbreviationNotHex    = fmt.Errorf("faws/repo/cas: abbreviation is not hexadecimal")

	ErrObjectNotFound = fmt.Errorf("faws/repo/cas: object not found")
)
