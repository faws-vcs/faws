package validate

import (
	"fmt"
	"strings"
)

var (
	ErrFileNameTooLong           = fmt.Errorf("faws/validate: filename is too long")
	ErrFileNameEmpty             = fmt.Errorf("faws/validate: filename is empty")
	ErrFileNameSpecialCharacters = fmt.Errorf("faws/validate: filename contains special characters")
	ErrFileNamePathSpecifiers    = fmt.Errorf("faws/validate: filename cannot be current or parent directory specifier")
	ErrFilePathEmpty             = fmt.Errorf("faws/validate: filepath is empty")
)

// Validate a single file
func FileName(filename string) (err error) {
	if filename == "" {
		err = ErrFileNameEmpty
		return
	}

	if len(filename) > 256 {
		err = ErrFileNameTooLong
		return
	}

	if strings.ContainsAny(filename, "*/:<>?\\|") {
		err = ErrFileNameSpecialCharacters
		return
	}

	if filename == "." || filename == ".." || filename == "..." {
		err = ErrFileNamePathSpecifiers
		return
	}

	return
}

func FilePath(filepath string) (err error) {
	if filepath == "" {
		err = ErrFilePathEmpty
		return
	}

	for _, path_component := range strings.Split(filepath, "/") {
		if err = FileName(path_component); err != nil {
			return
		}
	}

	return
}
