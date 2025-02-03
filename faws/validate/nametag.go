package validate

import (
	"fmt"
	"regexp"
)

var (
	ErrNametagTooLong           = fmt.Errorf("faws/validate: nametag is too long")
	ErrNametagInvalidCharacters = fmt.Errorf("faws/validate: nametag contains invalid characters")
)

var nametag_regex = regexp.MustCompilePOSIX("^[a-z0-9.]+$")

func Nametag(n string) (err error) {
	if n == "" {
		return
	}
	if len(n) > 63 {
		err = ErrNametagTooLong
		return
	}
	if !nametag_regex.MatchString(n) {
		err = ErrNametagInvalidCharacters
		return
	}
	return
}
