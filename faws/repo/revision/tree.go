package revision

import (
	"encoding/binary"
	"fmt"

	"github.com/faws-vcs/faws/faws/repo/cas"
)

const (
	MaxFileName    = 256
	MaxTreeEntries = 1e7
	MaxFileHashes  = 1e9
)

var (
	ErrFileNameTooLong            = fmt.Errorf("faws/repo/revision: filename is too long")
	ErrTreeContainsTooManyEntries = fmt.Errorf("faws/repo/revision: tree contains too many entries")
	ErrFileContainsTooManyHashes  = fmt.Errorf("faws/repo/revision: file entry contains too many content ID hashes")
)

type TreeEntry struct {
	// TREE or FILE?
	Prefix cas.Prefix
	// entry name must be < 256 bytes
	Name string
	// File mode
	Mode FileMode
	// Retrieve the file's data from CAS
	//  if FileModeDirectory,
	//    this should be one ID pointing to the tree object
	Content cas.ContentID
}

// A Tree is a root directory or a subdirectory
// in other words, it is one level of a filesystem hierarchy.
type Tree struct {
	// Sorted by name
	Entries []TreeEntry
}

// MarshalTree serializes the Tree into a binary representation
func MarshalTree(tree *Tree) (data []byte, err error) {
	// preallocate minimum tree capacity
	data = make([]byte, 0, 4+4+(len(tree.Entries)*((cas.ContentIDSize*2)+1+4)))

	// store size of tree
	var tree_size [4]byte
	binary.LittleEndian.PutUint32(tree_size[:], uint32(len(tree.Entries)))
	data = append(data, tree_size[:]...)

	// store tree entries
	for _, file := range tree.Entries {
		// store prefix
		data = append(data, file.Prefix[:]...)

		// store name
		var name_size [4]byte
		binary.LittleEndian.PutUint32(name_size[:], uint32(len(file.Name)))
		data = append(data, name_size[:]...)
		data = append(data, []byte(file.Name)...)

		// store file mode
		data = append(data, uint8(file.Mode))

		// store content hash
		data = append(data, file.Content[:]...)
	}

	return
}

func UnmarshalTree(data []byte, tree *Tree) (err error) {
	field := data

	num_entries := int(binary.LittleEndian.Uint32(field[:4]))
	field = field[4:]

	if num_entries > len(data) || num_entries > MaxTreeEntries {
		return ErrTreeContainsTooManyEntries
	}

	tree.Entries = make([]TreeEntry, num_entries)
	for i := range tree.Entries {
		entry := &tree.Entries[i]

		copy(entry.Prefix[:], field[:4])
		field = field[4:]

		name_size := binary.LittleEndian.Uint32(field[:4])
		field = field[4:]

		if name_size > MaxFileName {
			err = fmt.Errorf("%w: %d bytes", ErrFileNameTooLong, name_size)
			return
		}

		entry.Name = string(field[:name_size])
		field = field[name_size:]

		entry.Mode = FileMode(field[0])
		field = field[1:]

		copy(entry.Content[:], field[:cas.ContentIDSize])
		field = field[cas.ContentIDSize:]
	}

	return
}
