package repo

import (
	"bytes"
	"io"

	"github.com/faws-vcs/faws/faws/repo/cas"
)

type file_reader struct {
	repo        *Repository
	contents    []byte
	part        bool
	part_reader bytes.Reader
}

func (r *file_reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}
	rb := b
	var rn int

	for n < len(b) {
		if !r.part {
			if len(r.contents) == 0 {
				err = io.EOF
				return
			}
			var part_hash cas.ContentID
			copy(part_hash[:], r.contents[:cas.ContentIDSize])
			r.contents = r.contents[cas.ContentIDSize:]
			var (
				p cas.Prefix
				d []byte
			)
			p, d, err = r.repo.objects.Load(part_hash)
			if err != nil {
				return
			}
			if p != cas.Part {
				err = ErrBadObject
				return
			}
			r.part_reader.Reset(d)
			r.part = true
			continue
		}

		rn, err = r.part_reader.Read(rb[:])
		if err != nil {
			r.part = false
			continue
		}

		n += rn
		rb = rb[rn:]
	}

	return
}

func (r *file_reader) Close() (err error) {
	r.contents = nil
	r.part_reader.Reset(nil)
	return
}

func (repo *Repository) OpenFile(file_hash cas.ContentID) (file io.ReadCloser, err error) {
	reader := new(file_reader)
	reader.repo = repo
	var prefix cas.Prefix
	prefix, reader.contents, err = repo.objects.Load(file_hash)
	if err != nil {
		return
	}
	if prefix != cas.File {
		err = ErrBadObject
		return
	}

	file = reader
	return
}
