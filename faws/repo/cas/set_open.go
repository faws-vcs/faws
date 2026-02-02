package cas

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	fawsfs "github.com/faws-vcs/faws/faws/fs"
)

// Open will start using a directory to contain the [Set]. If a directory does not exist at path, one will be created.
// [Close] must be called when the [Set] is no longer in use.
func (set *Set) Open(path string) (err error) {
	path = strings.TrimRight(path, "\\/")

	set_fi, stat_err := os.Stat(path)
	if stat_err != nil {
		if !errors.Is(stat_err, fs.ErrNotExist) {
			err = stat_err
			return
		}

		if err = os.Mkdir(path, fawsfs.DefaultPublicDirPerm); err != nil {
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

	if err = set.cache.Open(set.directory); err != nil {
		return
	}
	if err = set.pack.Open(filepath.Join(set.directory, "/pack"), -1); err != nil {
		return
	}

	return
}
