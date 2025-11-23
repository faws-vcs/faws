package validate

import (
	"fmt"
	"strings"
	"unicode"
)

var (
	ErrNametagTooLong                    = fmt.Errorf("faws/validate: nametag is too long")
	ErrNametagContainsTrailingWhitespace = fmt.Errorf("faws/validate: nametag contains trailing whitespace")
	ErrNametagInvalidCharacters          = fmt.Errorf("faws/validate: nametag contains invalid characters")
)

func is_invalid_character(r rune) bool {
	return !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '.' || r == '_' || r == '-')
}

func Nametag(n string) (err error) {
	if n == "" {
		return
	}
	if len(n) > 63 {
		err = ErrNametagTooLong
		return
	}
	if strings.ContainsFunc(n, is_invalid_character) {
		err = ErrNametagInvalidCharacters
		return
	}
	return
}
