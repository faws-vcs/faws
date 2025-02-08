package cas

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	fawsfs "github.com/faws-vcs/faws/faws/fs"
)

func (set *Set) Open(path string) (err error) {
	set_fi, stat_err := os.Stat(path)
	if stat_err != nil {
		if !errors.Is(stat_err, fs.ErrNotExist) {
			err = stat_err
			return
		}

		if err = os.Mkdir(path, fawsfs.DefaultPerm); err != nil {
			return
		}
		set_fi, err = os.Stat(path)
		if err != nil {
			panic(err)
		}
	}

	if !set_fi.IsDir() {
		err = fmt.Errorf("faws/cas: is not a directory")
		return
	}

	test_path := filepath.Join(path, "testwrite")
	if err = os.WriteFile(test_path, nil, os.ModePerm); err != nil {
		return
	}
	os.Remove(test_path)

	set.directory = path
	return
}
