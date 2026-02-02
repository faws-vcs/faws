package cas

import "os"

func (cache *cache) Load(id ContentID) (prefix Prefix, data []byte, err error) {
	// AABBCCDDEEFF => cas_set/aa/bb/ccddeeff
	var object []byte
	object, err = os.ReadFile(cache.path(id))
	if err != nil {
		err = object_error{ErrObjectNotFound, id}
		return
	}
	if len(object) < 4 {
		err = object_error{ErrObjectCorrupted, id}
		return
	}
	copy(prefix[:], object)
	data = object[4:]

	// ensure consistency
	disk_id := hash_content(prefix, data)
	if id != disk_id {
		err = object_error{ErrObjectCorrupted, id}
		return
	}

	return
}
