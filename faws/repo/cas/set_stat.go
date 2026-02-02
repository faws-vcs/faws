package cas

// Stat tests the existence of an object named by the [ContentID], and returns its size if it does exist.
// If it does not exist, err will be [ErrObjectNotFound].
func (set *Set) Stat(id ContentID) (size int64, err error) {
	size, err = set.cache.Stat(id)
	if err == nil {
		return
	}

	size, err = set.pack.Stat(id)
	return
}
