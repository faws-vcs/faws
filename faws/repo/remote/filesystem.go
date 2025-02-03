package remote

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type DirEntry struct {
	Name  string
	IsDir bool
	Size  int64
}

type Fs interface {
	ReadDir(name string) (entries []DirEntry, err error)
	Pull(name string) (file io.ReadCloser, err error)
	// Push(name string, source io.Reader) (err error)
}

func Open(path string) (fs Fs, err error) {
	var u *url.URL
	u, err = url.Parse(path)
	if err == nil && u.Scheme != "" {
		fs, err = open_url(u)
		return
	}

	fs, err = open_directory(path)
	return
}

func open_url(u *url.URL) (fs Fs, err error) {
	switch u.Scheme {
	case "file":
		fs, err = open_directory(u.Path)
	default:
		err = fmt.Errorf("faws/remote: unknown scheme")
	}
	return
}

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
		if !e.IsDir() {
			info, info_err := e.Info()
			if info_err == nil {
				entries[i].Size = info.Size()
			}
		}
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
