package cas

import (
	"os"
)

// Attempt to load the corresponding data segment for a ContentID.
func (set *Set) Load(id ContentID) (prefix Prefix, data []byte, err error) {
	// AABBCCDDEEFF => cas_set/aa/bb/ccddeeff
	var path string
	path, err = set.path(id)
	if err != nil {
		return
	}
	var object []byte
	object, err = os.ReadFile(path)
	if err != nil {
		err = ErrObjectNotFound
		return
	}
	copy(prefix[:], object)
	data = object[4:]

	// ensure consistency
	disk_id := hash_content(prefix, data)
	if id != disk_id {
		err = ErrBadChecksum
		return
	}

	return
}
