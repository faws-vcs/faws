package cache

import (
	"encoding/binary"
	"fmt"

	"github.com/faws-vcs/faws/faws/multipart"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

var (
	ErrCacheEntryCannotBeEmpty = fmt.Errorf("faws/repo/cache: cache entry cannot be empty")
)

// An IndexEntry associates a path string with an object hash and a filemode
type IndexEntry struct {
	// The path inside the repository
	Path string
	// The hash of the cached file
	File cas.ContentID
	// The mode of the file
	Mode revision.FileMode
}

// A CacheObject is a handle to a loose object, or an object that is not necessarily part of the repository yet, but exists inside the cache or [cas.Set]
// Therefore it may be freed from the repository if the number of references in the index reaches 0
type CacheObject struct {
	// FILE or PART hash
	Hash cas.ContentID
	// number of times this hash is referenced across all entries
	References uint32
}

// A LazySignature memoizes the intensive process of scanning a file by exploiting format specific features.
// For instance, if you hash only the file size + table of contents of a large archive file, you may be certain
// that the hash acts as a unique signature that will not overlap with any other archive of the same format
type LazySignature struct {
	// The lazy signature of the file
	// These can't repeat, they control how the list is sorted.
	Signature multipart.LazySignature
	// The previously chunked and catalogued file object
	// If file is ever uncached, then the lazy signature MUST also be deleted
	File cas.ContentID
}

// Index lists pending changes to be written by the next commit
type Index struct {
	CacheObjects []CacheObject
	Entries      []IndexEntry
	// Completely optional: use lazy signatures to accelerate caching large collections of archive files
	LazySignatures []LazySignature
}

// MarshalIndex serializes the Index to a slice of bytes
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

	var lazy_signatures_count [4]byte
	binary.LittleEndian.PutUint32(lazy_signatures_count[:], uint32(len(index.LazySignatures)))
	data = append(data, lazy_signatures_count[:]...)
	for _, lazy_signature := range index.LazySignatures {
		data = append(data, lazy_signature.Signature[:]...)
		data = append(data, lazy_signature.File[:]...)
	}

	return
}

// UnmarshalIndex deserializes the slice of bytes into the Index
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

	lazy_signatures_count := binary.LittleEndian.Uint32(field[:4])
	field = field[4:]
	index.LazySignatures = make([]LazySignature, lazy_signatures_count)
	for i := range index.LazySignatures {
		ls := &index.LazySignatures[i]
		copy(ls.Signature[:], field[:multipart.LazySignatureSize])
		field = field[multipart.LazySignatureSize:]
		copy(ls.File[:], field[:cas.ContentIDSize])
		field = field[cas.ContentIDSize:]
	}

	return
}
