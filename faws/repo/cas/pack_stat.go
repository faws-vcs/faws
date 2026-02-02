package cas

func (pack *Pack) Stat(name ContentID) (size int64, err error) {
	pack.guard.RLock()
	defer pack.guard.RUnlock()

	var index_entry pack_index_entry
	index_entry, err = pack.index_get(name)
	if err != nil {
		return
	}

	var archive *pack_archive
	archive, err = pack.get_archive(int(index_entry.ArchiveID))
	if err != nil {
		return
	}

	size, err = archive.StatEntry(index_entry.FileOffset)
	return
}
