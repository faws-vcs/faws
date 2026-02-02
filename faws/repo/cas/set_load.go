package cas

import (
	"errors"
)

// Load attempts to load the corresponding content for a ContentID.
func (set *Set) Load(id ContentID) (prefix Prefix, data []byte, err error) {
	// attempt to load from the cache
	prefix, data, err = set.cache.Load(id)
	if err == nil {
		return
	} else if !errors.Is(err, ErrObjectNotFound) {
		return
	}

	// now try to load from the pack
	prefix, data, err = set.pack.Load(id)
	return
}
