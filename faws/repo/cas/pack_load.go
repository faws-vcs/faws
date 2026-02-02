package cas

func (pack *Pack) Load(name ContentID) (prefix Prefix, content []byte, err error) {
	pack.guard.RLock()
	// lookup the hash from the index. if there's no hit, the object does not exist.
	var (
		index_entry   pack_index_entry
		archive_entry pack_archive_entry
	)
	index_entry, err = pack.index_get(name)
	if err != nil {
		pack.guard.RUnlock()
		return
	}

	archive_id := int(index_entry.ArchiveID)
	var archive *pack_archive
	archive, err = pack.get_archive(archive_id)
	if err != nil {
		pack.guard.RUnlock()
		return
	}

	archive_entry, err = archive.ReadEntry(index_entry.FileOffset)
	if err != nil {
		pack.guard.RUnlock()
		return
	}
	pack.guard.RUnlock()

	if archive_entry.Flag&pack_tbd_prefix != 0 {
		prefix = archive_entry.TBDPrefix
	} else if archive_entry.Flag&pack_part_prefix != 0 {
		prefix = Part
	} else if archive_entry.Flag&pack_file_prefix != 0 {
		prefix = File
	} else if archive_entry.Flag&pack_tree_prefix != 0 {
		prefix = Tree
	} else if archive_entry.Flag&pack_commit_prefix != 0 {
		prefix = Commit
	} else {
		err = ErrObjectCorrupted
		return
	}

	content = archive_entry.Content
	// check hash
	if hash_content(prefix, content) != name {
		err = ErrObjectCorrupted
		return
	}

	return
}
