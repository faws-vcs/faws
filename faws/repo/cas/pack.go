package cas

import (
	"fmt"
	"sync"
)

var (
	archive_prefix = Prefix{'P', 'A', 'C', 'K'}
	index_prefix   = Prefix{'I', 'N', 'D', 'X'}
)

type Pack struct {
	guard sync.RWMutex
	// how
	max_archive_size int64
	// the directory that holds the pack files
	// for instance: objects/packed
	// a
	parent_directory string
	name             string
	// .
	index    pack_index
	archives []*pack_archive
}

func (pack *Pack) index_get(name ContentID) (entry pack_index_entry, err error) {
	entry, err = pack.index.Get(name)
	return
}

func (pack *Pack) get_archive(archive_id int) (archive *pack_archive, err error) {
	if archive_id >= len(pack.archives) {
		err = fmt.Errorf("%w: %s.%06d", ErrPackMissingArchive, pack.name, archive_id)
		return
	}

	archive = pack.archives[archive_id]
	if archive == nil {
		err = fmt.Errorf("%w: %s.%06d", ErrPackMissingArchive, pack.name, archive_id)
		return
	}

	return
}

func (pack *Pack) open_archive(archive_id int) (err error) {
	archive := new(pack_archive)
	if err = archive.Open(fmt.Sprintf("%s/%s.%06d", pack.parent_directory, pack.name, archive_id)); err != nil {
		return
	}
	if archive_id >= len(pack.archives) {
		missing_slots := make([]*pack_archive, (archive_id+1)-len(pack.archives))
		pack.archives = append(pack.archives, missing_slots...)
	}
	pack.archives[archive_id] = archive
	return
}

func (pack *Pack) index_put(entry pack_index_entry) (err error) {
	// write index entry
	err = pack.index.Put(entry)
	return
}

//
//

// type pack_index struct {
// // 	file *os.File
// // }

// type pack_file struct {
// 	object_file *os.File
// 	index_file  *os.File
// }

// type pack_object_entry struct {
// 	Flag    pack_object_entry_flag
// 	Content []byte
// }

// func (pack *pack) index_count() int {

// }

// func (pack *pack) index_object(index int) (entry pack_index_entry, err error) {

// }

// func (pack *pack) search_index(object_hash cas.ContentID) (index int) {
// 	index = sort.Search(pack.index_count(), func(i int) bool {

// 	})
// 	sort.Search(l)
// }

// func (pack *pack) Store(prefix cas.Prefix, data []byte) (new bool, id ContentID, err error) {

// }
