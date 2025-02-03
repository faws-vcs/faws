package multipart

import (
	"fmt"
	"io"

	"github.com/restic/chunker"
)

const (
	min_chunk_size = 0x1000
	max_chunk_size = 0x1000000

	chunker_polynomial chunker.Pol = 0x3DA3358B4DC173
)

type generic_chunker struct {
	buffer   []byte
	impl     *chunker.Chunker
	index    uint64
	position uint64
}

func new_generic_chunker(file io.Reader) (c *generic_chunker) {
	c = new(generic_chunker)
	c.buffer = make([]byte, max_chunk_size*2)
	c.impl = chunker.New(file, chunker_polynomial, chunker.WithBoundaries(min_chunk_size, max_chunk_size))
	// c.impl = chunker.New(file, chunker_polynomial)
	return
}

func (c *generic_chunker) Section() string {
	return fmt.Sprintf("index(%d)", c.index)
}

func (c *generic_chunker) Next() (position int64, chunk []byte, err error) {
	var data chunker.Chunk
	data, err = c.impl.Next(c.buffer)
	if err != nil {
		return
	}
	c.index++
	c.position = uint64(data.Start)
	position = int64(c.position)
	chunk = data.Data
	return
}
