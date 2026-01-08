package remote

import (
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type filesystem_local struct {
	name string
}

func (filesystem_local *filesystem_local) path(name string) (path string, err error) {
	entities := strings.Split(name, "/")
	path = filepath.Join(append([]string{filesystem_local.name}, entities...)...)
	return
}

func (filesystem_local *filesystem_local) ReadDir(name string) (entries []dir_entry, err error) {
	var (
		path       string
		os_entries []os.DirEntry
	)
	path, err = filesystem_local.path(name)
	if err != nil {
		return
	}
	os_entries, err = os.ReadDir(path)
	if err != nil {
		return
	}

	entries = make([]dir_entry, len(os_entries))
	for i, e := range os_entries {
		entries[i].Name = e.Name()
		entries[i].IsDir = e.IsDir()
	}
	return
}

func (filesystem_local *filesystem_local) Stat(name string) (size int64, err error) {
	var (
		fi os.FileInfo
	)
	fi, err = os.Stat(name)
	if err != nil {
		err = os.ErrNotExist
		return
	}

	if fi.IsDir() {
		err = os.ErrNotExist
		return
	}

	size = fi.Size()
	return
}

func (filesystem_local *filesystem_local) Pull(name string) (file io.ReadCloser, err error) {
	var path string
	path, err = filesystem_local.path(name)
	if err != nil {
		return
	}

	file, err = os.Open(path)
	return
}

func (filesystem_local *filesystem_local) URI() (s string) {
	var u url.URL
	u.Scheme = "file"
	u.Path = filesystem_local.name
	s = u.String()
	return
}

func open_filesystem_local(name string) (origin Origin, err error) {
	if !filepath.IsAbs(name) {
		var abs_path string
		abs_path, err = filepath.Abs(name)
		if err != nil {
			err = nil
		} else {
			name = abs_path
		}
	}
	filesystem_local_ := new(filesystem_local)
	filesystem_local_.name = name

	origin = filesystem_origin{filesystem_local_}
	return
}
