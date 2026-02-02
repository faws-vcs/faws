package cas

func (pack *Pack) Close() (err error) {
	if err = pack.index.Close(); err != nil {
		return
	}
	for _, archive := range pack.archives {
		if archive != nil {
			if err = archive.Close(); err != nil {
				return
			}
		}
	}

	pack.archives = nil

	return
}
