package repo

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

func Locked(directory string) (err error) {
	// check to see if repository is locked
	var pid int
	lock_file, lock_err := os.ReadFile(filepath.Join(directory, "lock"))
	if lock_err == nil {
		pid, err = strconv.Atoi(string(lock_file))
		if err != nil {
			panic(err)
		}
		err = fmt.Errorf("%w by process %d", ErrLocked, pid)
		return
	}
	return
}

func (repo *Repository) lock() (err error) {
	// check if repository is locked
	if err = Locked(repo.directory); err != nil {
		return
	}

	return
}

func (repo *Repository) unlock() (err error) {
	return
}
