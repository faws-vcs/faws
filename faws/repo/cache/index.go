package cache

import (
	"encoding/binary"
	"fmt"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

var (
	ErrCacheEntryCannotBeEmpty = fmt.Errorf("faws/repo/cache: cache entry cannot be empty")
)

type IndexEntry struct {
	// The path inside the repository
	Path string
	// The hash of the cached file
	File cas.ContentID
	// The mode of the file
	Mode revision.FileMode
}

type CacheObject struct {
	// FILE or PART hash
	Hash cas.ContentID
	// number of times this hash is referenced across all entries
	References uint32
}

// Index lists pending changes
// to be written by the next commit
type Index struct {
	CacheObjects []CacheObject
	Entries      []IndexEntry
}

func MarshalIndex(index *Index) (data []byte, err error) {
	var cache_objects_count [4]byte
	binary.LittleEndian.PutUint32(cache_objects_count[:], uint32(len(index.CacheObjects)))
	data = append(data, cache_objects_count[:]...)
	for _, cache_object := range index.CacheObjects {
		var references [4]byte
		binary.LittleEndian.PutUint32(references[:], uint32(cache_object.References))

		data = append(data, cache_object.Hash[:]...)
		data = append(data, references[:]...)
	}

	var entries_count [4]byte
	binary.LittleEndian.PutUint32(entries_count[:], uint32(len(index.Entries)))
	data = append(data, entries_count[:]...)

	for _, entry := range index.Entries {
		var (
			path_size [2]byte
		)

		binary.LittleEndian.PutUint16(path_size[:], uint16(len(entry.Path)))

		data = append(data, path_size[:]...)
		data = append(data, []byte(entry.Path)...)

		data = append(data, entry.File[:]...)

		data = append(data, byte(entry.Mode))
	}

	return
}

func UnmarshalIndex(data []byte, index *Index) (err error) {
	field := data

	cache_objects_count := binary.LittleEndian.Uint32(field[:4])
	field = field[4:]
	index.CacheObjects = make([]CacheObject, cache_objects_count)

	for i := range index.CacheObjects {
		cache_object := &index.CacheObjects[i]
		copy(cache_object.Hash[:], field[:cas.ContentIDSize])
		field = field[cas.ContentIDSize:]
		cache_object.References = binary.LittleEndian.Uint32(field[:4])
		field = field[4:]
	}

	entries_count := binary.LittleEndian.Uint32(field[:4])
	field = field[4:]

	index.Entries = make([]IndexEntry, entries_count)
	for i := range index.Entries {
		entry := &index.Entries[i]

		path_size := binary.LittleEndian.Uint16(field[:2])
		field = field[2:]
		entry.Path = string(field[:path_size])
		field = field[path_size:]

		copy(entry.File[:], field[:cas.ContentIDSize])
		field = field[cas.ContentIDSize:]

		entry.Mode = revision.FileMode(field[0])
		field = field[1:]
	}

	return
}
