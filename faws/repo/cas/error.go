package cas

import (
	"fmt"
)

const package_id = "faws/repo/cas"

var (
	ErrInvalidSet      = fmt.Errorf("%s: set is not a directory", package_id)
	ErrObjectCorrupted = fmt.Errorf("%s: object content ID does not match content on disk", package_id)

	ErrAbbreviationAmbiguous = fmt.Errorf("%s: abbreviation is ambiguous", package_id)
	ErrAbbreviationTooShort  = fmt.Errorf("%s: abbreviation is too short to be expanded", package_id)
	ErrAbbreviationNotHex    = fmt.Errorf("%s: abbreviation is not hexadecimal", package_id)

	ErrObjectNotFound = fmt.Errorf("%s: object not found", package_id)

	ErrInvalidPackIndexFile = fmt.Errorf("%s: invalid index file in pack", package_id)
	ErrPackIndexNotExist    = fmt.Errorf("%s: index file in pack does not exist", package_id)

	ErrPackArchiveNotExist      = fmt.Errorf("%s: the archive does not exist")
	ErrPackArchiveEntryNotExist = fmt.Errorf("%s: the archive entry does not exist", package_id)
	ErrPackArchiveBadEntry      = fmt.Errorf("%s: the archive entry is badly formed", package_id)
	ErrPackMissingArchive       = fmt.Errorf("%s: the pack index points to an archive that is missing", package_id)
	ErrPackFileCannotRemove     = fmt.Errorf("%s: packed files may not be removed this way. Running 'faws gc' will take care of unused files")
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
