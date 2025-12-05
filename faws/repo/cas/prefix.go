package cas

import "encoding/hex"

// A Prefix helps clarify the purpose of each object.
// The presence of prefixes allows for many tricks to be performed
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

// String returns an ordinary name for each Prefix type
// 1. File = "file"
// 2. Part = "part"
// 3. Tree = "tree"
// 3. Commit = "commit"
func (p Prefix) String() string {
	switch p {
	case File:
		return "file"
	case Part:
		return "part"
	case Tree:
		return "tree"
	case Commit:
		return "commit"
	}
	return "bad prefix(" + hex.EncodeToString(p[:]) + ")"
}
