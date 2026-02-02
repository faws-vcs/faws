package cas

import (
	"sort"
)

type pack_writer_index_entry struct {
	archive_id  int
	file_offset int64
}

// PackWriter is a tool for efficiently generating pack databases
type PackWriter struct {
	index map[ContentID]pack_writer_index_entry
	// stored map[ContentID]
	pack Pack
}

func (pack_writer *PackWriter) Open(name string, max_archive_size int64) (err error) {
	pack_writer.index = make(map[ContentID]pack_writer_index_entry)
	err = pack_writer.pack.Open(name, max_archive_size)
	return
}

func (pack_writer *PackWriter) Store(prefix Prefix, data []byte) (new bool, id ContentID, err error) {
	id = hash_content(prefix, data)

	pack_writer.pack.guard.Lock()
	defer pack_writer.pack.guard.Unlock()
	// check if the object already exists

	_, ok := pack_writer.index[id]
	if ok {
		// already written!
		return
	}

	new = true

	var archive_entry pack_archive_entry
	archive_entry.Flag = pack_exists
	if prefix == Part {
		archive_entry.Flag |= pack_part_prefix
	} else if prefix == File {
		archive_entry.Flag |= pack_file_prefix
	} else if prefix == Tree {
		archive_entry.Flag |= pack_tree_prefix
	} else if prefix == Commit {
		archive_entry.Flag |= pack_commit_prefix
	} else {
		archive_entry.Flag |= pack_tbd_prefix
		archive_entry.TBDPrefix = prefix
	}
	if len(data) > 0 {
		archive_entry.Flag |= pack_contains_data
		archive_entry.Content = data
	}

	// append to the last archive
	var (
		archive    *pack_archive
		archive_id int
	)
	for archive_id = range pack_writer.pack.archives {
		if pack_writer.pack.archives[archive_id] != nil {
			archive = pack_writer.pack.archives[archive_id]
		}
	}

	if archive != nil {
		if archive.Size()+pack_archive_entry_size(archive_entry) >= pack_writer.pack.max_archive_size {
			// a new archive must be created
			archive = nil
		}
	}

	if archive == nil {
		archive_id = len(pack_writer.pack.archives)
		if err = pack_writer.pack.open_archive(archive_id); err != nil {
			return
		}
		archive = pack_writer.pack.archives[archive_id]
	}

	var offset int64
	offset, err = archive.WriteEntry(archive_entry)
	if err != nil {
		return
	}

	var index_entry pack_writer_index_entry
	index_entry.archive_id = archive_id
	index_entry.file_offset = offset

	pack_writer.index[id] = index_entry

	return
}

type index_sorter struct {
	index *pack_index
}

func (index_sorter *index_sorter) Len() (n int) {
	n = int(index_sorter.index.num_entries())
	return
}

func (index_sorter *index_sorter) Less(i, j int) (less bool) {
	i_entry, err := index_sorter.index.read_entry(int64(i))
	if err != nil {
		panic(err)
	}
	j_entry, err := index_sorter.index.read_entry(int64(j))
	if err != nil {
		panic(err)
	}
	less = i_entry.Name.Less(j_entry.Name)
	return
}

func (index_sorter *index_sorter) Swap(i, j int) {
	if err := index_sorter.index.swap_entry(int64(i), int64(j)); err != nil {
		panic(err)
	}
}

func (pack_writer *PackWriter) Close() (err error) {
	var i int64
	for id, entry := range pack_writer.index {
		var real_entry pack_index_entry
		real_entry.Name = id
		real_entry.ArchiveID = entry.archive_id
		real_entry.FileOffset = entry.file_offset
		if err = pack_writer.pack.index.write_entry(i, real_entry); err != nil {
			return
		}
		if err = pack_writer.pack.index.add_fanout_value(id[0]); err != nil {
			return
		}
		i++
	}

	var index_sorter index_sorter
	index_sorter.index = &pack_writer.pack.index
	sort.Sort(&index_sorter)

	err = pack_writer.pack.Close()
	return
}
