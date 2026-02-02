package repo

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/multipart"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/event"
	"github.com/faws-vcs/faws/faws/repo/pathspec"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/faws-vcs/faws/faws/repo/staging"
)

// An IndexEntry associates a path string with an object hash and a filemode
type staging_index_entry struct {
	// The hash of the cached file
	file cas.ContentID
	// The mode of the file
	mode revision.FileMode
}

// a working version of staging.Index
type staging_index struct {
	// maps paths to entries
	entries map[string]staging_index_entry
	// maps lazy signatures to files
	lazy_signatures map[multipart.LazySignature]cas.ContentID
}

// Index returns the index of the staging area
func (repo *Repository) Index() (index *staging.Index) {
	index = new(staging.Index)
	// collect index entries
	index.Entries = make([]staging.IndexEntry, 0, len(repo.index.entries))
	for path, index_entry := range repo.index.entries {
		index.Entries = append(index.Entries, staging.IndexEntry{
			Path: path,
			File: index_entry.file,
			Mode: index_entry.mode,
		})
	}
	// sort entries by path
	slices.SortFunc(index.Entries, func(a, b staging.IndexEntry) int {
		return strings.Compare(a.Path, b.Path)
	})
	// collect lazy signatures
	index.LazySignatures = make([]staging.LazySignature, 0, len(repo.index.lazy_signatures))
	for lazy_signature, file := range repo.index.lazy_signatures {
		index.LazySignatures = append(index.LazySignatures, staging.LazySignature{
			Signature: lazy_signature,
			File:      file,
		})
	}
	// sort by signature
	slices.SortFunc(index.LazySignatures, func(a, b staging.LazySignature) int {
		return bytes.Compare(a.Signature[:], b.Signature[:])
	})
	return
}

// meant to be called recursively:
// with destination as ""
// each tree will be traversed, restoring the index
func (repo *Repository) reset_tree(destination string, tree_hash cas.ContentID) (err error) {
	var tree *revision.Tree
	tree, err = repo.load_tree(tree_hash)
	if err != nil {
		return
	}

	if destination != "" {
		destination += "/"
	}

	for _, tree_entry := range tree.Entries {
		child_destination := destination + tree_entry.Name
		switch tree_entry.Prefix {
		case cas.File:
			repo.index.entries[child_destination] = staging_index_entry{tree_entry.Content, tree_entry.Mode}
		case cas.Tree:
			if err = repo.reset_tree(child_destination, tree_entry.Content); err != nil {
				return
			}
		}
	}

	return
}

// Reset resets the index to a specific commit
func (repo *Repository) Reset(commit_hash cas.ContentID) (err error) {
	if commit_hash == cas.Nil {
		repo.index.entries = nil
	} else {
		var info *revision.CommitInfo
		_, info, err = repo.check_commit(commit_hash)
		if err != nil {
			return
		}
		err = repo.reset_tree("", info.Tree)
	}

	return
}

type staging_options struct {
	set_mode bool
	mode     revision.FileMode
	lazy     bool
}

// A StagingOption can be used to add specific options to a staging operation
type StagingOption func(*staging_options)

// WithFileMode is a [StagingOption] that sets [revision.FileMode] of the cached files to mode
func WithFileMode(mode revision.FileMode) StagingOption {
	return func(c *staging_options) {
		c.set_mode = true
		c.mode = mode
	}
}

// WithLazy is a [CacheOption] that enables lazy-signatures for the cached files
func WithLazy(lazy bool) StagingOption {
	return func(c *staging_options) {
		c.lazy = lazy
	}
}

// ingest a file or directory into the staging area
func (repo *Repository) stage_file(o *staging_options, destination, source string) (err error) {
	source_info, stat_err := os.Stat(source)
	if stat_err != nil {
		err = stat_err
		return
	}

	// with destination as "", a directory is mapped directly into the staging area
	// you can't have a file with no name
	if destination == "" && !source_info.IsDir() {
		err = ErrIndexNameCannotBeEmpty
		return
	}

	if source_info.IsDir() {
		source_directory_entries, ls_err := os.ReadDir(source)
		if ls_err != nil {
			err = ls_err
			return
		}

		for _, directory_entry := range source_directory_entries {
			// combine parent and child
			// but avoid leading slash if the destination is "" i.e. (map the source's children to the root of the index)
			// if destination != '',       child destination = 'child' not '/child'
			// if destination == 'parent', child_destination = 'parent/child'
			child_destination := destination
			if child_destination != "" {
				child_destination += "/"
			}
			child_destination += directory_entry.Name()
			if err = repo.stage_file(o, child_destination, filepath.Join(source, directory_entry.Name())); err != nil {
				return
			}
		}

		return
	}

	// check for conflicts with existing files
	if err = repo.check_index_destination_for_dir_conflict(destination); err != nil {
		return
	}

	var entry staging_index_entry

	if o.set_mode {
		entry.mode = o.mode
	} else {
		if source_info.Mode()&0111 != 0 {
			// if any executable bit is set, the file is an executable.
			entry.mode = revision.FileModeExecutable
		}
	}

	var (
		source_file *os.File
		chunker     multipart.Chunker
	)
	source_file, err = os.Open(source)
	if err != nil {
		return
	}

	defer source_file.Close()

	var notify_params event.NotifyParams
	notify_params.Name1 = destination
	notify_params.Name2 = source
	notify_params.Count = source_info.Size()
	repo.notify(event.NotifyCacheFile, &notify_params)

	chunker, err = multipart.NewChunker(source_file)
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
	// this trick allows us to wholly ingest previously-scanned MPQ files
	// in a fraction of the time
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

			lazy_file_hash, ok := repo.index.lazy_signatures[lazy_signature]
			if ok {
				// notify UI that we're lazy
				var notify_params event.NotifyParams
				notify_params.Name1 = destination
				notify_params.Object1 = lazy_file_hash
				repo.notify(event.NotifyCacheUsedLazySignature, &notify_params)

				// we
				var lazy_file []byte
				_, lazy_file, err = repo.objects.Load(lazy_file_hash)
				if err != nil {
					return
				}
				var lazy_file_part_hash cas.ContentID
				for len(lazy_file) > 0 {
					copy(lazy_file_part_hash[:], lazy_file[:cas.ContentIDSize])
					lazy_file = lazy_file[cas.ContentIDSize:]
				}

				entry.file = lazy_file_hash
				repo.index.entries[destination] = entry

				chunking_file_stage.Success = true
				repo.notify(event.NotifyCompleteStage, &chunking_file_stage)

				return
			}
		}
	}

	// Remove existing entries with this path (deferred in case of detection of lazy signature)
	delete(repo.index.entries, destination)

	var (
		chunk    []byte
		chunk_id cas.ContentID
		file     []byte
		file_id  cas.ContentID
	)
	for {
		_, chunk, err = chunker.Next()
		if err != nil && errors.Is(err, io.EOF) {
			err = nil
			break
		} else if err != nil {
			err = fmt.Errorf("faws/repo: cache_file: error reading chunker: %w", err)
			return
		}

		_, chunk_id, err = repo.objects.Store(cas.Part, chunk)
		if err != nil {
			return
		}

		var notify_cache_file_part event.NotifyParams
		notify_cache_file_part.Count = int64(len(chunk))
		repo.notify(event.NotifyCacheFilePart, &notify_cache_file_part)

		file = append(file, chunk_id[:]...)
	}

	_, file_id, err = repo.objects.Store(cas.File, file)
	if err != nil {
		return
	}

	entry.file = file_id
	repo.index.entries[destination] = entry

	chunking_file_stage.Success = err == nil
	repo.notify(event.NotifyCompleteStage, &chunking_file_stage)

	if err != nil {
		return
	}

	if got_lazy_signature {
		repo.index.lazy_signatures[lazy_signature] = file_id
	}

	return
}

// Add imports a file into the repository, and then adds it to the index (or staging area)
func (repo *Repository) Add(destination, source string, options ...StagingOption) (err error) {
	var o staging_options
	for _, option := range options {
		option(&o)
	}

	// convert source to absolute path
	abs_source, abs_err := filepath.Abs(source)
	if abs_err == nil {
		source = abs_source
	}
	// convert back to forward slash
	source = strings.ReplaceAll(source, "\\", "/")

	var notify_params event.NotifyParams
	notify_params.Stage = event.StageCacheFiles
	repo.notify(event.NotifyBeginStage, &notify_params)

	err = repo.stage_file(&o, destination, source)

	notify_params.Success = err == nil
	repo.notify(event.NotifyCompleteStage, &notify_params)

	return
}

// Remove removes all files in the index that are matched by a pathspec pattern
func (repo *Repository) Remove(pattern string) (err error) {
	if pattern == "" {
		err = ErrNoPathspec
		return
	}
	var pathspec_ *pathspec.Pathspec
	pathspec_, err = pathspec.Compile(pattern)
	if err != nil {
		return
	}
	n := 0

	index_names := slices.Collect(maps.Keys(repo.index.entries))
	sort.Strings(index_names)

	for _, name := range index_names {
		if pathspec_.MatchString(name) {
			n++
			delete(repo.index.entries, name)

			var deleted_entry event.NotifyParams
			deleted_entry.Name1 = name
			repo.notify(event.NotifyIndexRemoveFile, &deleted_entry)
		}
	}
	if n == 0 {
		err = fmt.Errorf("%w: %s", ErrNoPathspecMatch, pattern)
	}
	return
}

// // RemoveFile removes a staged file by its exact path
// func (repo *Repository) RemoveFile(path string) (err error) {
// 	if _, ok := repo.index.entries[path]; ok {
// 		delete(repo.index.entries, path)
// 		return
// 	}
// 	err = fmt.Errorf("faws/repo: path not found: '%s'", path)
// 	return
// }

// Chmod changes the filemode of a cached file within the index
func (repo *Repository) Chmod(pattern string, mode revision.FileMode) (err error) {
	// entry, ok := repo.index.entries[path]
	// if !ok {
	// 	err = ErrIndexEntryNotFound
	// 	return
	// }
	// entry.mode = mode
	// repo.index.entries[path] = entry

	if pattern == "" {
		err = ErrNoPathspec
		return
	}
	var pathspec_ *pathspec.Pathspec
	pathspec_, err = pathspec.Compile(pattern)
	if err != nil {
		return
	}
	n := 0
	for path, entry := range repo.index.entries {
		if pathspec_.MatchString(path) {
			n++
			entry.mode = mode
			repo.index.entries[path] = entry
		}
	}
	if n == 0 {
		err = fmt.Errorf("%w: %s", ErrNoPathspecMatch, pattern)
	}
	return
}

func (repo *Repository) read_index() (err error) {
	var staging_index staging.Index
	if index_data, index_err := os.ReadFile(filepath.Join(repo.directory, "index")); index_err == nil {
		if err = staging.UnmarshalIndex(index_data, &staging_index); err != nil {
			return
		}
	}
	repo.index.entries = make(map[string]staging_index_entry, len(staging_index.Entries))
	for _, entry := range staging_index.Entries {
		repo.index.entries[entry.Path] = staging_index_entry{file: entry.File, mode: entry.Mode}
	}

	repo.index.lazy_signatures = make(map[multipart.LazySignature]cas.ContentID, len(staging_index.Entries))
	for _, lazy_signature := range staging_index.LazySignatures {
		repo.index.lazy_signatures[lazy_signature.Signature] = lazy_signature.File
	}
	return
}

func (repo *Repository) write_index() (err error) {
	var index_data []byte
	index_data, err = staging.MarshalIndex(repo.Index())
	if err != nil {
		return
	}

	err = os.WriteFile(filepath.Join(repo.directory, "index"), index_data, fs.DefaultPrivatePerm)
	return
}

// check if the file destination is already occupied by a directory name
// e.g.
// if a file named "things/thing"
// exists in the index
// you would not be able to add a file named "things"
func (repo *Repository) check_index_destination_for_dir_conflict(destination string) (err error) {
	conflict_prefix := destination
	if !strings.HasSuffix(conflict_prefix, "/") {
		conflict_prefix += "/"
	}
	current_entries := slices.Collect(maps.Keys(repo.index.entries))
	sort.Strings(current_entries)
	for _, current_entry := range current_entries {
		if strings.HasPrefix(current_entry, conflict_prefix) {
			err = ErrIndexPathConflict
			return
		}
	}
	return
}

func (repo *Repository) add_object(options *staging_options, destination string, source cas.ContentID) (err error) {
	// remove trailing slash from destination
	destination = strings.TrimSuffix(destination, "/")

	var (
		source_prefix  cas.Prefix
		source_content []byte
	)
	source_prefix, source_content, err = repo.objects.Load(source)
	if err != nil {
		return
	}

	switch source_prefix {
	case cas.File:
		if err = repo.check_index_destination_for_dir_conflict(destination); err != nil {
			return
		}
		// very good. since an index is just map of paths to
		var mode revision.FileMode
		if options.set_mode {
			mode = options.mode
		}
		repo.index.entries[destination] = staging_index_entry{source, mode}
	case cas.Tree:
		// not so easy. we load the tree, and recursively add all of its children to the index.
		var (
			tree revision.Tree
		)
		err = revision.UnmarshalTree(source_content, &tree)
		if err != nil {
			return
		}
		for _, entry := range tree.Entries {
			child_destination := destination
			if child_destination != "" {
				child_destination += "/"
			}
			child_destination += entry.Name
			if err = repo.add_object(options, child_destination, source); err != nil {
				return
			}
		}
	case cas.Commit:
		// kinda weird that you would use a commit here. However, we can just use the tree.
		var (
			commit_info revision.CommitInfo
			commit      revision.Commit
		)
		err = revision.UnmarshalCommit(source_content, &commit)
		if err != nil {
			return
		}
		if err = revision.UnmarshalCommitInfo(commit.Info, &commit_info); err != nil {
			return
		}
		err = repo.add_object(options, destination, commit_info.Tree)
		return
	default:
		err = fmt.Errorf("%w: %s", ErrIndexBadObjectPrefix, source_prefix)
		return
	}

	return
}

// Add adds an existing object
func (repo *Repository) AddObject(destination string, source cas.ContentID, options ...StagingOption) (err error) {
	var options_ staging_options
	for _, option := range options {
		option(&options_)
	}
	err = repo.add_object(&options_, destination, source)
	return
}
