package cas

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
