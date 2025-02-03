package repository

import (
	"fmt"

	"github.com/faws-vcs/faws/faws/repo"
)

func notify(ev repo.Ev, args ...any) {
	fmt.Print(ev, " ")
	fmt.Println(args...)
}
