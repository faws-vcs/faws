package revision

import "strconv"

type FileMode uint8

const (
	// if true, the file has executable permissions
	FileModeExecutable FileMode = 1 << iota
)

func (m FileMode) String() string {
	s := "-"
	if m&FileModeExecutable != 0 {
		s = "x"
	}
	return s
}

func ParseFileMode(s string) (m FileMode, err error) {
	if s == "-" {
		m = 0
		return
	} else if s == "x" {
		m = FileModeExecutable
		return
	}

	var u uint64
	u, err = strconv.ParseUint(s, 10, 8)
	if err != nil {
		return
	}
	m = FileMode(u)
	return
}
