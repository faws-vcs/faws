package cas

func (cache *cache) Open(directory string) (err error) {
	cache.directory = directory
	return
}
