package cas

import (
	"encoding/hex"
	"os"
	"strings"

	"github.com/faws-vcs/faws/faws/fs"
)

// Set is the set of all objects held by the repository and index
type Set struct {
	// the location of the cas.Set. this never contains a trailing slash
	directory string
}

func (set *Set) path(id ContentID) (path string, err error) {
	var encoded [ContentIDSize * 2]byte
	hex.Encode(encoded[:], id[:])
	var builder strings.Builder
	builder.WriteString(set.directory)
	builder.WriteString("/")
	builder.Write(encoded[0:2])
	builder.WriteString("/")
	builder.Write(encoded[2:4])
	builder.WriteString("/")
	builder.Write(encoded[4:])
	path = builder.String()
	return
}

func (set *Set) store_path(id ContentID) (path string, err error) {
	var encoded [ContentIDSize * 2]byte
	hex.Encode(encoded[:], id[:])
	var builder strings.Builder
	builder.WriteString(set.directory)
	builder.WriteString("/")
	builder.Write(encoded[0:2])
	builder.WriteString("/")
	builder.Write(encoded[2:4])
	prefix := builder.String()
	err = os.MkdirAll(prefix, fs.DefaultPublicDirPerm)
	if err != nil {
		return
	}
	builder.WriteString("/")
	builder.Write(encoded[4:])
	path = builder.String()
	return
}
