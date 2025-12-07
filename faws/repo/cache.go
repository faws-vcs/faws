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

	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/multipart"
	"github.com/faws-vcs/faws/faws/repo/cache"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/revision"
)

type cache_index struct {
	cache_objects   map[cas.ContentID]uint32
	entries         []cache.IndexEntry
	lazy_signatures []cache.LazySignature
}

// CacheIndex returns the current index
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
	cache_index.LazySignatures = repo.index.lazy_signatures
	return
}

// CachedFiles returns the repository's list of index entries
func (repo *Repository) CachedFiles() []cache.IndexEntry {
	return repo.index.entries
}

// UncacheAll removes all files from the index and uncaches them
func (repo *Repository) UncacheAll() (err error) {
	if len(repo.index.cache_objects) == 0 {
		repo.index.entries = nil
		return
	}

	entries := repo.index.entries
	paths := make([]string, len(entries))
	for index, entry := range entries {
		paths[index] = entry.Path
	}

	for _, path := range paths {
		if err = repo.Uncache(path); err != nil {
			return
		}
	}
	return
}

// ResetCache fully resets the index and removes cached index files
func (repo *Repository) ResetCache() (err error) {
	// keep lazy signatures that point to persistent objects
	var new_lazy_signatures []cache.LazySignature
	for _, lazy_signature := range repo.index.lazy_signatures {
		if !repo.index_object_is_cache(lazy_signature.File) {
			new_lazy_signatures = append(new_lazy_signatures, lazy_signature)
		}
	}

	// remove non-persistent objects
	err = repo.UncacheAll()
	if err != nil {
		return
	}

	// keep lazy signatures that point to persistent objects
	repo.index.lazy_signatures = new_lazy_signatures
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
		repo.remove_lazy_file(object_hash)
		repo.objects.Remove(object_hash)
		delete(repo.index.cache_objects, object_hash)
		return
	}
	repo.index.cache_objects[object_hash] = references
}

func (repo *Repository) find_lazy_file(signature multipart.LazySignature) (file_hash cas.ContentID, err error) {
	i := sort.Search(len(repo.index.lazy_signatures), func(i int) bool {
		return !repo.index.lazy_signatures[i].Signature.Less(signature)
	})
	if i < len(repo.index.lazy_signatures) && repo.index.lazy_signatures[i].Signature == signature {
		file_hash = repo.index.lazy_signatures[i].File
		return
	}
	err = ErrCacheEntryNotFound
	return
}

func (repo *Repository) insert_lazy_file(signature multipart.LazySignature, file_hash cas.ContentID) {
	i := sort.Search(len(repo.index.lazy_signatures), func(i int) bool {
		return !repo.index.lazy_signatures[i].Signature.Less(signature)
	})
	if i < len(repo.index.lazy_signatures) && repo.index.lazy_signatures[i].Signature == signature {
		repo.index.lazy_signatures[i].File = file_hash
	} else {
		repo.index.lazy_signatures = slices.Insert(repo.index.lazy_signatures, i, cache.LazySignature{
			Signature: signature,
			File:      file_hash,
		})
	}
}

func (repo *Repository) remove_lazy_file(file_hash cas.ContentID) (err error) {
	for i, signature := range repo.index.lazy_signatures {
		if signature.File == file_hash {
			repo.index.lazy_signatures = slices.Delete(repo.index.lazy_signatures, i, i+1)
			break
		}
	}
	err = ErrCacheEntryNotFound
	return
}

type cache_options struct {
	set_mode bool
	mode     revision.FileMode
	lazy     bool
}

// A CacheOption can be used to add specific options to a cache operation
type CacheOption func(*cache_options)

// WithFileMode is a [CacheOption] that sets [revision.FileMode] of the cached files to mode
func WithFileMode(mode revision.FileMode) CacheOption {
	return func(c *cache_options) {
		c.set_mode = true
		c.mode = mode
	}
}

// WithLazy is a [CacheOption] that enables lazy-signatures for the cached files
func WithLazy(lazy bool) CacheOption {
	return func(c *cache_options) {
		c.lazy = lazy
	}
}

func (repo *Repository) cache_file(o *cache_options, path, origin string) (err error) {
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
			if err = repo.cache_file(o, filepath.Join(path, entry.Name()), filepath.Join(origin, entry.Name())); err != nil {
				return
			}
		}

		return
	}

	// check for conflicts with existing files
	directory_path := path
	if !strings.HasSuffix(directory_path, "/") {
		directory_path += "/"
	}
	entries := repo.index.entries
	for _, entry := range entries {
		if strings.HasPrefix(entry.Path, directory_path) {
			err = fmt.Errorf("faws/repo: path used for new file conflicts with existing use as a directory")
			return
		}
	}

	var entry cache.IndexEntry
	entry.Path = path

	if o.set_mode {
		entry.Mode = o.mode
	} else {
		if fi.Mode()&0111 != 0 {
			// if any executable bit is set, the file is an executable.
			entry.Mode = revision.FileModeExecutable
		}
	}

	var (
		origin_file      *os.File
		origin_file_info os.FileInfo
		chunker          multipart.Chunker
	)
	origin_file, err = os.Open(origin)
	if err != nil {
		return
	}
	origin_file_info, err = origin_file.Stat()
	if err != nil {
		return
	}

	defer origin_file.Close()

	var notify_params event.NotifyParams
	notify_params.Name1 = path
	notify_params.Name2 = origin
	notify_params.Count = origin_file_info.Size()
	repo.notify(event.NotifyCacheFile, &notify_params)

	chunker, err = multipart.NewChunker(origin_file)
	if err != nil {
		return
	}

	var chunking_file_stage event.NotifyParams
	chunking_file_stage.Stage = event.StageCacheFile
	chunking_file_stage.Child = true
	repo.notify(event.NotifyBeginStage, &chunking_file_stage)

	// hold on to this
	var (
		lazy_signature     multipart.LazySignature
		got_lazy_signature bool
	)

	// sometimes you can get away with being a little bit lazy
	if o.lazy {
		lazy_chunker, can_be_lazy := chunker.(multipart.LazyChunker)
		if can_be_lazy {
			var (
				lazy_file_hash cas.ContentID
			)
			lazy_signature, err = lazy_chunker.LazySignature()
			if err != nil {
				return
			}
			got_lazy_signature = true

			lazy_file_hash, err = repo.find_lazy_file(lazy_signature)
			if err == nil {
				var notify_params event.NotifyParams
				notify_params.Name1 = entry.Path
				notify_params.Object1 = lazy_file_hash
				repo.notify(event.NotifyCacheUsedLazySignature, &notify_params)
				// found lazy file id!
				// increase all the references to its contents
				if repo.index_object_is_cache(lazy_file_hash) {
					repo.reference_index_cache_object(lazy_file_hash)
				}
				var lazy_file []byte
				_, lazy_file, err = repo.objects.Load(lazy_file_hash)
				if err != nil {
					return
				}
				var lazy_file_part_hash cas.ContentID
				for len(lazy_file) > 0 {
					copy(lazy_file_part_hash[:], lazy_file[:cas.ContentIDSize])
					if repo.index_object_is_cache(lazy_file_part_hash) {
						repo.reference_index_cache_object(lazy_file_part_hash)
					}
					lazy_file = lazy_file[cas.ContentIDSize:]
				}

				entry.File = lazy_file_hash
				err = repo.insert_cache_index_entry(entry)

				chunking_file_stage.Success = err == nil
				repo.notify(event.NotifyCompleteStage, &chunking_file_stage)

				return
			}
		}
	}

	// Remove existing entries with this path (deferred in case of detection of lazy signature)
	for _, entry := range repo.index.entries {
		if entry.Path == path {
			if err = repo.Uncache(path); err != nil {
				return
			}
		}
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
			err = fmt.Errorf("faws/repo: cache_file: error reading chunker: %w", err)
			return
		}

		new, chunk_id, err = repo.objects.Store(cas.Part, chunk)
		if err != nil {
			return
		}

		var notify_cache_file_part event.NotifyParams
		notify_cache_file_part.Count = int64(len(chunk))
		repo.notify(event.NotifyCacheFilePart, &notify_cache_file_part)

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

	chunking_file_stage.Success = err == nil
	repo.notify(event.NotifyCompleteStage, &chunking_file_stage)

	if err != nil {
		return
	}

	if got_lazy_signature {
		repo.insert_lazy_file(lazy_signature, file_id)
	}

	return
}

// Cache caches a file and adds it to the index
func (repo *Repository) Cache(path, origin string, options ...CacheOption) (err error) {
	var o cache_options
	for _, option := range options {
		option(&o)
	}

	var notify_params event.NotifyParams
	notify_params.Stage = event.StageCacheFiles
	repo.notify(event.NotifyBeginStage, &notify_params)

	err = repo.cache_file(&o, path, origin)

	notify_params.Success = err == nil
	repo.notify(event.NotifyCompleteStage, &notify_params)

	return
}

// Cache uncaches a file and removes it from the index
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
				err = fmt.Errorf("%w: %s", err, entry.Path)
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

	err = fmt.Errorf("faws/repo/cache: path not found: '%s'", path)
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
	repo.index.lazy_signatures = cache_index.LazySignatures
	return
}

func (repo *Repository) write_index() (err error) {
	var index_data []byte
	index_data, err = cache.MarshalIndex(repo.CacheIndex())
	if err != nil {
		return
	}

	err = os.WriteFile(filepath.Join(repo.directory, "index"), index_data, fs.DefaultPrivatePerm)
	return
}

// CacheSetFileMode changes the filemode of a cached file within the index
func (repo *Repository) CacheSetFileMode(path string, mode revision.FileMode) (err error) {
	i := repo.find_cache_index_entry(path)
	if i < len(repo.index.entries) && repo.index.entries[i].Path == path {
		repo.index.entries[i].Mode = mode
		return
	}

	err = ErrCacheEntryNotFound
	return
}
