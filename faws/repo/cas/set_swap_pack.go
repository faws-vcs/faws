package cas

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func copy_file(destination, source string) (err error) {
	var (
		destination_file *os.File
		source_file      *os.File
	)
	destination_file, err = os.Create(destination)
	if err != nil {
		return
	}
	source_file, err = os.Open(source)
	if err != nil {
		return
	}
	_, err = io.Copy(destination_file, source_file)
	if err != nil {
		return
	}
	destination_file.Close()
	source_file.Close()
	return
}

func move_file(destination, source string) (err error) {
	if err = os.Rename(source, destination); err != nil {
		if err = copy_file(destination, source); err != nil {
			return
		}
		err = os.Remove(source)
	}
	return
}

// SwapPack removes the current pack and swaps in the named pack. If keep == true, the source pack is preserved.
func (set *Set) SwapPack(name string, keep bool) (err error) {
	// scan for pack index and archives
	source_directory := filepath.Dir(name)
	var (
		source_directory_entries []os.DirEntry
	)
	source_pack_id := filepath.Base(name)
	source_directory_entries, err = os.ReadDir(source_directory)
	if err != nil {
		return
	}

	// collect the list of pack files we need to copy (or move) into the set
	var source_pack_names []string
	for _, source_directory_entry := range source_directory_entries {
		if !source_directory_entry.IsDir() {
			if source_directory_entry.Name() == source_pack_id {
				source_pack_names = append(source_pack_names, source_pack_id)
			} else {
				sibling_name, archive_id_name, found := strings.Cut(source_directory_entry.Name(), ".")
				if found {
					if sibling_name == source_pack_id {
						_, err = strconv.ParseInt(archive_id_name, 10, 32)
						if err == nil {
							source_pack_names = append(source_pack_names, source_directory_entry.Name())
						}
						err = nil
					}
				}
			}
		}
	}

	if err = set.pack.Close(); err != nil {
		return
	}

	// remove current pack files
	var set_directory_entries []os.DirEntry
	set_directory_entries, err = os.ReadDir(set.directory)
	if err != nil {
		return
	}

	for _, set_directory_entry := range set_directory_entries {
		if !set_directory_entry.IsDir() {
			if set_directory_entry.Name() == "pack" {
				if err = os.Remove(filepath.Join(set.directory, "pack")); err != nil {
					return
				}
			} else {
				sibling_name, archive_id_name, found := strings.Cut(set_directory_entry.Name(), ".")
				if found {
					if sibling_name == "pack" {
						_, err = strconv.ParseInt(archive_id_name, 10, 32)
						if err == nil {
							if err = os.Remove(filepath.Join(set.directory, set_directory_entry.Name())); err != nil {
								return
							}
						}
						err = nil
					}
				}
			}
		}
	}

	// swap in new pack files
	for _, source_pack_name := range source_pack_names {
		var set_pack_name string
		if source_pack_name == source_pack_id {
			set_pack_name = "pack"
		} else {
			_, archive_id_name, found := strings.Cut(source_pack_name, ".")
			if !found {
				panic(source_pack_name)
			}
			set_pack_name = "pack." + archive_id_name
		}
		if keep {
			if err = copy_file(filepath.Join(set.directory, set_pack_name), filepath.Join(source_directory, source_pack_name)); err != nil {
				return
			}
		} else {
			if err = move_file(filepath.Join(set.directory, set_pack_name), filepath.Join(source_directory, source_pack_name)); err != nil {
				return
			}
		}
	}

	err = set.pack.Open(set.directory+"/pack", -1)
	return
}
