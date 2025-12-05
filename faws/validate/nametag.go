package validate

import (
	"fmt"
	"strings"
)

var (
	ErrNametagTooLong                    = fmt.Errorf("faws/validate: nametag is too long")
	ErrNametagContainsTrailingWhitespace = fmt.Errorf("faws/validate: nametag contains trailing whitespace")
	ErrNametagInvalidCharacters          = fmt.Errorf("faws/validate: nametag contains invalid characters")
)

// Nametag returns an error if the nametag is invalid
func Nametag(n string) (err error) {
	if n == "" {
		return
	}
	if len(n) > 63 {
		err = ErrNametagTooLong
		return
	}
	if strings.ContainsFunc(n, is_invalid_tag_character) {
		err = ErrNametagInvalidCharacters
		return
	}
	return
}
