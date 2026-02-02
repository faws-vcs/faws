package cas

// Remove removes a *cached* object by its [ContentID]
func (set *Set) Remove(id ContentID) (err error) {
	err = set.cache.Remove(id)
	return

}
