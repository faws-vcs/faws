package multipart

import (
	"io"
)

type single_chunk struct {
	rs io.ReadSeeker
}

func (sc single_chunk) Section() (n string) {
	n = "file"
	return
}

func (sc single_chunk) Next() (position int64, data []byte, err error) {
	data, err = io.ReadAll(sc.rs)
	if len(data) == 0 {
		err = io.EOF
	}
	return
}
