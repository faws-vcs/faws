package remote

import (
	"fmt"
	"io"
	"net/url"
)

type DirEntry struct {
	Name  string
	IsDir bool
}

type Fs interface {
	URL() string
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
	case "http", "https":
		fs, err = open_web_server(u)
	default:
		err = fmt.Errorf("faws/remote: unknown scheme")
	}
	return
}
