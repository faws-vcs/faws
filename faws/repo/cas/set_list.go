package cas

// ListFunc is a callback for enumerating unsorted objects in the repository. If the function returns non-nil, the list will be aborted.
type ListFunc func(packed bool, id ContentID) (err error)

// List will enumerate all objects in the Set using the supplied [ListFunc] callback.
//
// If the function returns non-nil, the list will be aborted.
// You may directly read objects while using the ListFunc, but you may not write or remove objects
func (set *Set) List(fn ListFunc) (err error) {
	if err = set.cache.List(fn); err != nil {
		return
	}

	err = set.pack.List(fn)
	return
}
