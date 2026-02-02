package cas

import (
	"encoding/hex"
	"os"
	"strings"

	"github.com/faws-vcs/faws/faws/fs"
)

// returns the cache path for an object
func (cache *cache) path(id ContentID) (path string) {
	var encoded [ContentIDSize * 2]byte
	hex.Encode(encoded[:], id[:])
	var builder strings.Builder
	builder.WriteString(cache.directory)
	builder.WriteString("/")
	builder.Write(encoded[0:2])
	builder.WriteString("/")
	builder.Write(encoded[2:4])
	builder.WriteString("/")
	builder.Write(encoded[4:])
	path = builder.String()
	return
}

func (cache *cache) make_path(id ContentID) (path string, err error) {
	var encoded [ContentIDSize * 2]byte
	hex.Encode(encoded[:], id[:])
	var builder strings.Builder
	builder.WriteString(cache.directory)
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
