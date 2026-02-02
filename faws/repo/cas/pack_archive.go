package cas

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/faws-vcs/faws/faws/fs"
)

type pack_archive_entry_flag uint8

const (
	pack_archive_header_size = PrefixSize
)

const (
	// the entry is active. if this flag is not present, this entry was "removed" from the pack
	pack_exists pack_archive_entry_flag = 1 << iota
	// the entry is not prefix-only it contains data.
	pack_contains_data
	// the entry uses the prefix 'PART'
	pack_part_prefix
	// the entry uses the prefix 'FILE'
	pack_file_prefix
	// the entry uses the prefix 'TREE'
	pack_tree_prefix
	// the entry uses the prefix 'EDIT'
	pack_commit_prefix
	// the entry uses some other prefix, which will be specified.
	pack_tbd_prefix
)

type pack_archive_entry struct {
	Flag      pack_archive_entry_flag
	TBDPrefix Prefix
	Content   []byte
}

func pack_archive_entry_size(entry pack_archive_entry) (size int64) {
	size = 1

	if entry.Flag&pack_tbd_prefix != 0 {
		size += PrefixSize
	}

	if entry.Flag&pack_contains_data != 0 {
		var content_len [8]byte
		size += int64(binary.PutUvarint(content_len[:], uint64(len(entry.Content))))
		size += int64(len(entry.Content))
	}

	return
}

type pack_archive_header struct {
	// 'PACK'
	Prefix Prefix
}

// archive represents an actively loaded PACK file. a pack can be made of many different files
type pack_archive struct {
	guard     sync.RWMutex
	header    pack_archive_header
	file      *os.File
	file_size int64
}

func (pack_archive *pack_archive) read_header() (err error) {
	_, err = pack_archive.file.ReadAt(pack_archive.header.Prefix[:], 0)
	if err != nil {
		err = fmt.Errorf("cas: in reading pack archive header: %w", err)
	}
	return
}

func (pack_archive *pack_archive) write_header() (err error) {
	_, err = pack_archive.file.Write(pack_archive.header.Prefix[:])
	return
}

func (pack_archive *pack_archive) Open(name string) (err error) {
	pack_archive.file, err = os.OpenFile(name, os.O_CREATE|os.O_RDWR, fs.DefaultPublicPerm)
	if err != nil {
		return
	}

	pack_archive.file_size, err = pack_archive.file.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}

	if pack_archive.file_size < pack_archive_header_size {
		pack_archive.header.Prefix = archive_prefix
		err = pack_archive.write_header()
		if err != nil {
			return
		}
		pack_archive.file_size = pack_archive_header_size
	} else {
		err = pack_archive.read_header()
		if err != nil {
			return
		}
	}

	return
}

func (pack_archive *pack_archive) ReadEntry(offset int64) (entry pack_archive_entry, err error) {
	pack_archive.guard.RLock()
	defer pack_archive.guard.RUnlock()

	// not a valid offset
	if offset >= pack_archive.file_size {
		err = ErrPackArchiveEntryNotExist
		return
	}

	reader := bufio.NewReader(io.NewSectionReader(pack_archive.file, offset, pack_archive.file_size-offset))

	var flag_byte uint8
	if flag_byte, err = reader.ReadByte(); err != nil {
		err = fmt.Errorf("cas: in reading flag from pack entry @%d: %w", offset, err)
		return
	}

	entry.Flag = pack_archive_entry_flag(flag_byte)
	if entry.Flag&pack_exists == 0 {
		// the entry was deleted
		err = ErrPackArchiveEntryNotExist
		return
	}

	if entry.Flag&pack_tbd_prefix != 0 {
		if _, err = io.ReadFull(reader, entry.TBDPrefix[:]); err != nil {
			err = fmt.Errorf("cas: in reading tbd prefix from pack entry @%d: %w", offset, err)
			return
		}
	}

	// it's technically valid if an object contains no data
	if entry.Flag&pack_contains_data != 0 {
		var content_length uint64
		content_length, err = binary.ReadUvarint(reader)
		if err != nil {
			err = fmt.Errorf("cas: in reading length from pack entry @%d: %w", offset, err)
			return
		}
		if content_length > MaxObjectSize {
			err = ErrPackArchiveBadEntry
			return
		}
		entry.Content = make([]byte, content_length)
		_, err = io.ReadFull(reader, entry.Content)
		if err != nil {
			err = fmt.Errorf("cas: in reading content from pack entry @%d: %w", offset, err)
			return
		}
	}

	return
}

// appends an object to the pack archive
func (pack_archive *pack_archive) WriteEntry(entry pack_archive_entry) (offset int64, err error) {
	pack_archive.guard.Lock()
	defer pack_archive.guard.Unlock()

	offset = pack_archive.file_size

	// the archive's file offset is always at the end
	writer := io.NewOffsetWriter(pack_archive.file, pack_archive.file_size)

	// each entry is at least a flag, if nothing else.
	_, err = writer.Write([]byte{byte(entry.Flag)})
	if err != nil {
		return
	}
	pack_archive.file_size++

	// the entry is using a special prefix
	if entry.Flag&pack_tbd_prefix != 0 {
		if _, err = writer.Write(entry.TBDPrefix[:]); err != nil {
			return
		}
		pack_archive.file_size += PrefixSize
	}

	// the entry contains data: we must encode its length and data
	if entry.Flag&pack_contains_data != 0 {
		var content_length [10]byte
		content_length_width := binary.PutUvarint(content_length[:], uint64(len(entry.Content)))
		_, err = writer.Write(content_length[:content_length_width])
		if err != nil {
			return
		}
		pack_archive.file_size += int64(content_length_width)
		_, err = writer.Write(entry.Content)
		if err != nil {
			return
		}
		pack_archive.file_size += int64(len(entry.Content))
	}

	return
}

func (pack_archive *pack_archive) StatEntry(offset int64) (size int64, err error) {
	// we don't want to create a sparse file by seeking out too far
	if offset >= pack_archive.file_size {
		err = ErrPackArchiveEntryNotExist
		return
	}
	reader := bufio.NewReader(io.NewSectionReader(pack_archive.file, offset, pack_archive.file_size-offset))

	var flag_byte uint8
	flag_byte, err = reader.ReadByte()
	if err != nil {
		err = fmt.Errorf("cas: in stat entry flag @%d: %w", offset, err)
		return
	}

	flag := pack_archive_entry_flag(flag_byte)
	if flag&pack_exists == 0 {
		// the entry was deleted
		err = fmt.Errorf("%w: %d %d", ErrPackArchiveEntryNotExist, offset, flag)
		return
	}

	if flag&pack_tbd_prefix != 0 {
		var tbd_prefix Prefix
		if _, err = io.ReadFull(reader, tbd_prefix[:]); err != nil {
			err = fmt.Errorf("cas: in stat entry tbd prefix @%d: %w", offset, err)
			return
		}
	}

	if flag&pack_contains_data != 0 {
		// it's okay if an object is just a prefix and nothing else.
		var content_length uint64
		content_length, err = binary.ReadUvarint(reader)
		if err != nil {
			err = fmt.Errorf("cas: in stat entry data length @%d: %w", offset, err)
			return
		}
		if content_length > MaxObjectSize {
			err = ErrPackArchiveBadEntry
			return
		}
		size = int64(content_length)
	}

	return

}

func (pack_archive *pack_archive) Size() (n int64) {
	n = pack_archive.file_size
	return
}

func (pack_archive *pack_archive) Close() (err error) {
	pack_archive.file_size = 0
	pack_archive.header = pack_archive_header{}
	err = pack_archive.file.Close()
	return
}
