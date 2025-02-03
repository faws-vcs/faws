package cas

import (
	"errors"
	"fmt"
)

var (
	ErrBadChecksum = errors.New("faws/cas: bad checksum")

	ErrAbbreviationAmbiguous = fmt.Errorf("faws/repo/cas: abbreviation is ambiguous")
	ErrAbbreviationTooShort  = fmt.Errorf("faws/repo/cas: abbreviation is too short to be expanded")

	ErrObjectNotFound = fmt.Errorf("faws/repo/cas: object not found")
)
