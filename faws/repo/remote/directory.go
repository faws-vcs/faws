package remote

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

type directory struct {
	name string
}

func (d *directory) path(name string) (path string, err error) {
	entities := strings.Split(name, "/")
	path = filepath.Join(append([]string{d.name}, entities...)...)
	return
}

func (d *directory) ReadDir(name string) (entries []DirEntry, err error) {
	var (
		path       string
		os_entries []os.DirEntry
	)
	path, err = d.path(name)
	if err != nil {
		return
	}
	os_entries, err = os.ReadDir(path)
	if err != nil {
		return
	}

	entries = make([]DirEntry, len(os_entries))
	for i, e := range os_entries {
		entries[i].Name = e.Name()
		entries[i].IsDir = e.IsDir()
	}
	return
}

func (d *directory) Pull(name string) (file io.ReadCloser, err error) {
	var path string
	path, err = d.path(name)
	if err != nil {
		return
	}

	file, err = os.Open(path)
	return
}

func open_directory(name string) (fs Fs, err error) {
	d := new(directory)
	d.name = name
	fs = d
	return
}
