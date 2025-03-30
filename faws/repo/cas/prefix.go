package cas

import "encoding/hex"

// A Prefix helps clarify the purpose of each object
// for instance, you can discover the st
type Prefix [4]byte

var (
	// The entry contains a list of file parts
	File = Prefix{'F', 'I', 'L', 'E'}
	// The entry contains a part of a file
	Part = Prefix{'P', 'A', 'R', 'T'}
	// The entry is one directory in a tree-hierarchy of directories
	Tree = Prefix{'T', 'R', 'E', 'E'}
	// The entry is a commit object
	Commit = Prefix{'E', 'D', 'I', 'T'}
)

func (p Prefix) String() string {
	switch p {
	case File:
		return "file"
	case Part:
		return "filepart"
	case Tree:
		return "tree"
	case Commit:
		return "commit"
	}
	return "bad prefix(" + hex.EncodeToString(p[:]) + ")"
}
