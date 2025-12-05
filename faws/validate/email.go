package validate

import (
	"fmt"
	"net/mail"
)

var (
	ErrEmailTooLong = fmt.Errorf("faws/validate: email is too long")
)

// Email returns an error if email is invalid
func Email(n string) (err error) {
	if n == "" {
		return
	}
	_, err = mail.ParseAddress(n)
	return
}
