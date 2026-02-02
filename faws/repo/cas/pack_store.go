package cas

func (pack *Pack) Store(prefix Prefix, data []byte) (new bool, id ContentID, err error) {
	id = hash_content(prefix, data)

	pack.guard.Lock()
	defer pack.guard.Unlock()
	// check if the object already exists
	var index_entry pack_index_entry
	index_entry, err = pack.index_get(id)
	if err == nil {
		var existing_object_archive *pack_archive
		existing_object_archive, err = pack.get_archive(index_entry.ArchiveID)
		if err == nil {
			_, err = existing_object_archive.StatEntry(index_entry.FileOffset)
			if err == nil {
				return
			}
		}
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
	for archive_id = range pack.archives {
		if pack.archives[archive_id] != nil {
			archive = pack.archives[archive_id]
		}
	}

	if archive != nil {
		if archive.Size()+pack_archive_entry_size(archive_entry) >= pack.max_archive_size {
			// a new archive must be created
			archive = nil
		}
	}

	if archive == nil {
		archive_id = len(pack.archives)
		if err = pack.open_archive(archive_id); err != nil {
			return
		}
		archive = pack.archives[archive_id]
	}

	var offset int64
	offset, err = archive.WriteEntry(archive_entry)
	if err != nil {
		return
	}

	index_entry.Name = id
	index_entry.ArchiveID = archive_id
	index_entry.FileOffset = offset

	err = pack.index_put(index_entry)
	return
}
