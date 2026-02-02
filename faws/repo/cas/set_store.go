package cas

// Store attempts to store data, returning the associated ContentID.
func (set *Set) Store(prefix Prefix, data []byte) (new bool, id ContentID, err error) {
	new, id, err = set.cache.Store(prefix, data)
	return
}
