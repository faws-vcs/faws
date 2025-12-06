package multipart

import (
	"bytes"
	"encoding/hex"
	"io"

	mpqinfo "github.com/Gophercraft/mpq/info"
	"github.com/faws-vcs/faws/faws/app"
)

type Chunker interface {
	// If applicable, returns a non-unique name for the next chunk
	Section() string
	Next() (position int64, chunk []byte, err error)
}

const LazySignatureSize = 32

type LazySignature [LazySignatureSize]byte

func (ls LazySignature) String() (s string) {
	s = hex.EncodeToString(ls[:])
	return
}

func (ls LazySignature) Less(than LazySignature) bool {
	return bytes.Compare(ls[:], than[:]) == -1
}

// A LazyChunker provides a LazySignature, a very quickly obtained hash of the file's headers.
//
// This allows you to quickly check the index to see if you've already chunked a large file.
//
// There's no point to doing this if a lazy signature does not,
// in some way, describe the entire file by reading a very small subset of it.
// Remember, it's a LAZY signature. It's here because reading the entire file
// is far too cumbersome and expensive.
type LazyChunker interface {
	Chunker
	// Returns a (hopefully unique) signature of the archive file, simply
	// by looking at its file size, header and content tables
	// this can be very dangerous if you aren't careful
	LazySignature() (signature LazySignature, err error)
}

func NewChunker(file io.ReadSeeker) (chunker Chunker, err error) {
	var size int64
	size, err = file.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return
	}
	if size < min_chunk_size {
		chunker = single_chunk{file}
		return
	}

	// detect magic
	var magic [4]byte
	if _, err = io.ReadFull(file, magic[:]); err != nil {
		return
	}
	if _, err = file.Seek(0, io.SeekStart); err != nil {
		return
	}

	switch magic {
	case mpqinfo.HeaderDataSignature:
		chunker, err = new_mpq_chunker(file, size)
		if err == nil {
			return
		}
		if err != nil {
			panic(err)
		}
		app.Info("mpq failed to start: ", err)
	default:
	}

	chunker = new_generic_chunker(file)
	return
}
