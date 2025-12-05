package remote

import (
	"fmt"
	"io"
	"net/url"
)

// A DirEntry is used to enumerate items in a remote filesystem
type DirEntry struct {
	Name  string
	IsDir bool
}

// Fs describes an interface to a remote filesystem
type Fs interface {
	// URL returns the URL string for the filesystem so it can be accessed at a later date
	URL() string
	// ReadDir enumerates items in a remote directory
	ReadDir(name string) (entries []DirEntry, err error)
	// Pull starts reading a file from the remote filesystem
	Pull(name string) (file io.ReadCloser, err error)
}

// Open opens a remote filesystem using a local path or a URL string
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
