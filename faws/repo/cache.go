package repo

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/faws-vcs/faws/faws/multipart"
	"github.com/faws-vcs/faws/faws/repo/cache"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

type cache_index struct {
	cache_objects map[cas.ContentID]uint32
	entries       []cache.IndexEntry
}

func (repo *Repository) CacheIndex() (cache_index *cache.Index) {
	cache_index = new(cache.Index)
	for object_hash, references := range repo.index.cache_objects {
		cache_index.CacheObjects = append(cache_index.CacheObjects, cache.CacheObject{
			Hash:       object_hash,
			References: references,
		})
	}
	sort.Slice(cache_index.CacheObjects, func(i, j int) bool {
		return cache_index.CacheObjects[i].Hash.Less(cache_index.CacheObjects[j].Hash)
	})
	cache_index.Entries = repo.index.entries
	return
}

func (repo *Repository) CachedFiles() []cache.IndexEntry {
	return repo.index.entries
}

func (repo *Repository) EmptyCache() (err error) {
	entries := repo.index.entries
	for _, entry := range entries {
		if err = repo.Uncache(entry.Path); err != nil {
			return
		}
	}
	return
}

func (repo *Repository) find_cache_index_entry(path string) (i int) {
	i = sort.Search(len(repo.index.entries), func(i int) bool {
		return repo.index.entries[i].Path >= path
	})
	return
}

func (repo *Repository) insert_cache_index_entry(entry cache.IndexEntry) (err error) {
	i := repo.find_cache_index_entry(entry.Path)
	if i < len(repo.index.entries) && repo.index.entries[i].Path == entry.Path {
		// replace
		repo.index.entries[i] = entry
	} else {
		repo.index.entries = slices.Insert(repo.index.entries, i, entry)
	}
	return
}

func (repo *Repository) index_object_is_cache(object_hash cas.ContentID) (is_cache bool) {
	_, is_cache = repo.index.cache_objects[object_hash]
	return
}

func (repo *Repository) reference_index_cache_object(object_hash cas.ContentID) {
	repo.index.cache_objects[object_hash] = repo.index.cache_objects[object_hash] + 1
}

func (repo *Repository) dereference_index_cache_object(object_hash cas.ContentID) {
	references, is_cache := repo.index.cache_objects[object_hash]
	if !is_cache {
		return
	}
	references--
	if references == 0 {
		repo.objects.Delete(object_hash)
		delete(repo.index.cache_objects, object_hash)
		return
	}
	repo.index.cache_objects[object_hash] = references
}

// 	return

func (repo *Repository) Cache(path, origin string) (err error) {
	abs_origin, abs_err := filepath.Abs(origin)
	if abs_err == nil {
		origin = abs_origin
	}

	fi, stat_err := os.Stat(origin)
	if stat_err != nil {
		err = stat_err
		return
	}

	if path == "" && !fi.IsDir() {
		err = ErrCacheEntryCannotBeEmpty
		return
	}

	if fi.IsDir() {
		entries, ls_err := os.ReadDir(origin)
		if ls_err != nil {
			err = ls_err
			return
		}

		for _, entry := range entries {
			if err = repo.Cache(filepath.Join(path, entry.Name()), filepath.Join(origin, entry.Name())); err != nil {
				return
			}
		}

		return
	}

	directory_path := path
	if !strings.HasSuffix(directory_path, "/") {
		directory_path += "/"
	}
	for _, entry := range repo.index.entries {
		if strings.HasPrefix(entry.Path, directory_path) {
			err = fmt.Errorf("faws/repo/cache: path prefix is already used as a directory")
			return
		}
		if entry.Path == path {
			if err = repo.Uncache(path); err != nil {
				return
			}
		}
	}

	var entry cache.IndexEntry
	entry.Path = path
	if fi.Mode()&0111 != 0 {
		// if any executable bit is set, the file is an executable.
		entry.Mode = revision.FileModeExecutable
	}

	var (
		origin_file *os.File
		chunker     multipart.Chunker
	)
	origin_file, err = os.Open(origin)
	if err != nil {
		return
	}

	fmt.Println("scanning", origin)
	chunker, err = multipart.NewChunker(origin_file)
	if err != nil {
		return
	}

	var (
		chunk    []byte
		chunk_id cas.ContentID
		file     []byte
		file_id  cas.ContentID
	)
	var new bool
	for {
		_, chunk, err = chunker.Next()
		if err != nil && errors.Is(err, io.EOF) {
			err = nil
			break
		} else if err != nil {
			return
		}

		new, chunk_id, err = repo.objects.Store(cas.Part, chunk)
		if err != nil {
			return
		}

		// if the object did not exist until now
		// or the object is a cache object,
		// it's definitely not part of the repository
		if new || repo.index_object_is_cache(chunk_id) {
			repo.reference_index_cache_object(chunk_id)
		}

		file = append(file, chunk_id[:]...)
	}

	new, file_id, err = repo.objects.Store(cas.File, file)
	if err != nil {
		return
	}

	if new || repo.index_object_is_cache(file_id) {
		repo.reference_index_cache_object(file_id)
	}

	entry.File = file_id

	err = repo.insert_cache_index_entry(entry)
	return
}

func (repo *Repository) Uncache(path string) (err error) {
	i := repo.find_cache_index_entry(path)
	if i < len(repo.index.entries) && repo.index.entries[i].Path == path {
		entry := &repo.index.entries[i]
		// if the file isn't part of the repository we should delete it
		if repo.index_object_is_cache(entry.File) {
			var (
				file   []byte
				prefix cas.Prefix
			)
			prefix, file, err = repo.objects.Load(entry.File)
			if err != nil {
				return
			}
			if prefix != cas.File {
				err = fmt.Errorf("faws/repo: incorrect object prefix")
				return
			}
			var file_part cas.ContentID
			part := file
			for len(part) > 0 {
				copy(file_part[:], part[:cas.ContentIDSize])
				part = part[cas.ContentIDSize:]

				if repo.index_object_is_cache(file_part) {
					repo.dereference_index_cache_object(file_part)
				}
			}

			if repo.index_object_is_cache(entry.File) {
				repo.dereference_index_cache_object(entry.File)
			}
		}

		repo.index.entries = slices.Delete(repo.index.entries, i, i+1)
		return
	}

	err = fmt.Errorf("faws/repo/cache: path not found")
	return
}

func (repo *Repository) read_index() (err error) {
	var cache_index cache.Index
	if index_data, index_err := os.ReadFile(filepath.Join(repo.directory, "index")); index_err == nil {
		if err = cache.UnmarshalIndex(index_data, &cache_index); err != nil {
			return
		}
	}
	repo.index.cache_objects = make(map[cas.ContentID]uint32, len(cache_index.CacheObjects))
	for _, cache_object := range cache_index.CacheObjects {
		repo.index.cache_objects[cache_object.Hash] = cache_object.References
	}
	repo.index.entries = cache_index.Entries
	return
}

func (repo *Repository) write_index() (err error) {
	var index_data []byte
	index_data, err = cache.MarshalIndex(repo.CacheIndex())
	if err != nil {
		return
	}

	err = os.WriteFile(filepath.Join(repo.directory, "index"), index_data, os.ModePerm)
	return
}
