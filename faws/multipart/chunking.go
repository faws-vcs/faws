package multipart

import (
	"fmt"
	"io"

	mpqinfo "github.com/Gophercraft/mpq/info"
	"github.com/davecgh/go-spew/spew"
)

type Chunker interface {
	// If applicable, returns a non-unique name for the next chunk
	Section() string
	Next() (position int64, chunk []byte, err error)
}

func NewChunker(file io.ReadSeeker) (chunker Chunker, err error) {
	// detect magic
	var magic [4]byte

	if _, err = io.ReadFull(file, magic[:]); err != nil {
		err = nil
		file.Seek(0, io.SeekStart)
		chunker = new_generic_chunker(file)
		return
	}
	file.Seek(0, io.SeekStart)

	switch magic {
	case mpqinfo.HeaderDataSignature:
		chunker, err = new_mpq_chunker(file)
		if err == nil {
			return
		}
		if err != nil {
			panic(err)
		}
		fmt.Println("mpq failed to start: ", err)
	default:
	}

	fmt.Println("no smart chunker for this signature", spew.Sdump(magic))
	chunker = new_generic_chunker(file)
	return
}
