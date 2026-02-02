package cas

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func (pack *Pack) Open(name string, max_archive_size int64) (err error) {
	pack.max_archive_size = max_archive_size
	if pack.max_archive_size < 0 {
		pack.max_archive_size = math.MaxInt64
	}
	pack.name = filepath.Base(name)
	pack.parent_directory = filepath.Dir(name)
	// open index
	if err = pack.index.Open(name); err != nil {
		return
	}

	// look through directory siblings for archives
	var pack_siblings []os.DirEntry
	pack_siblings, err = os.ReadDir(pack.parent_directory)
	if err != nil {
		return
	}
	var found_archives []int
	max_archive_id := -1
	for _, pack_sibling := range pack_siblings {
		sibling_name, archive_id_name, found := strings.Cut(pack_sibling.Name(), ".")
		if found {
			if sibling_name == pack.name {
				var archive_id64 int64
				archive_id64, err = strconv.ParseInt(archive_id_name, 10, 32)
				if err == nil {
					archive_id := int(archive_id64)
					if archive_id > max_archive_id {
						max_archive_id = archive_id
					}
					found_archives = append(found_archives, int(archive_id))
				}
				err = nil
			}
		}
	}

	// create a lookup table so that archives can be accessed by their ID
	// this is so that, for instance, if one archive is missing or is corrupted
	// other archives can be accessed
	pack.archives = make([]*pack_archive, max_archive_id+1)

	// open existing archives
	for _, archive_id := range found_archives {
		if err = pack.open_archive(archive_id); err != nil {
			return
		}
	}

	return
}
