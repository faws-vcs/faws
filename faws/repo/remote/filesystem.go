package remote

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

// A DirEntry is used to enumerate items in a remote filesystem
type dir_entry struct {
	Name  string
	IsDir bool
}

// filesystem describes an interface to a remote filesystem
type filesystem interface {
	// URL returns the URL string for the filesystem so it can be accessed at a later date
	URI() string
	// ReadDir enumerates items in a remote directory
	ReadDir(name string) (entries []dir_entry, err error)
	// Retrieves the size of the file, or an error if it is inaccessible
	Stat(name string) (size int64, err error)
	// Pull starts reading a file from the remote filesystem
	Pull(name string) (file io.ReadCloser, err error)
}

func is_uri(name string) (is_uri bool) {
	var (
		scheme  string
		was_cut bool
	)
	scheme, _, was_cut = strings.Cut(name, ":")
	if !was_cut {
		return
	}

	switch scheme {
	case "http", "https", "topic":
		is_uri = true
		// case "git+https://"
		// todo: implement git host free-riding
		// i.e. storing massive repositories inside of git trees,
		// download git trees and cache them in-memory,
		// while downloading individual Faws objects directly
	default:
		is_uri = false
	}
	return
}

// Open opens a remote Origin using a named local directory or URI
func Open(name string) (origin Origin, err error) {
	if is_uri(name) {
		origin, err = open_uri(name)
		return
	}

	origin, err = open_filesystem_local(name)
	return
}

func open_uri(uri string) (origin Origin, err error) {
	scheme, _, was_cut := strings.Cut(uri, ":")
	if !was_cut {
		return
	}

	switch scheme {
	case "file":
		var file_url *url.URL
		file_url, err = url.Parse(uri)
		if err != nil {
			return
		}
		origin, err = open_filesystem_local(file_url.Path)
	case "http", "https":
		var website_url *url.URL
		website_url, err = url.Parse(uri)
		if err != nil {
			return
		}
		origin, err = open_filesystem_website(website_url)
	default:
		err = fmt.Errorf("%w: %s", scheme)
	}
	return
}
