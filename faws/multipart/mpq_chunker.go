package multipart

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"slices"
	"sort"

	"github.com/Gophercraft/mpq/bit"
	"github.com/Gophercraft/mpq/compress"
	"github.com/Gophercraft/mpq/crypto"
	"github.com/Gophercraft/mpq/info"
)

type mpq_section_type uint8

const (
	mpq_section_extra mpq_section_type = iota
	mpq_section_user_data_header
	mpq_section_archive_header
	mpq_section_file_block
	mpq_section_het_table
	mpq_section_bet_table
	mpq_section_hash_table
	mpq_section_block_table
)

var (
	mpq_section_type_names = []string{
		"extra data",
		"userdata header",
		"archive header",
		"file block",
		"HET table",
		"BET table",
		"hash table",
		"block table",
	}

	hash_table_decryption_key  = crypto.HashString("(hash table)", crypto.HashEncryptKey)
	block_table_decryption_key = crypto.HashString("(block table)", crypto.HashEncryptKey)

	attributes_name_hash_1 = crypto.HashString("(attributes)", crypto.HashNameA)
	attributes_name_hash_2 = crypto.HashString("(attributes)", crypto.HashNameB)
)

func (t mpq_section_type) String() string {
	return mpq_section_type_names[t]
}

type mpq_section struct {
	Type   mpq_section_type
	Offset int64
}

// mpq_chunker works by building a map of data sections based on header information
//
// 0-10 header
// 10+  etc...
//
// then reading from
type mpq_chunker struct {
	archive_header   info.Header
	archive_position int64
	// the file
	file_size   int64
	file        io.ReadSeeker
	file_reader io.Reader
	// file_reader  *bufio.Reader
	chunk_reader Chunker
	// sorted by offset
	sections         []mpq_section
	index            int
	hash_table_data  []byte
	block_table_data []byte
	// largest used offset
	max_offset int64
	// largest section size
	max_section int64
	// the index of the (attributes) file in the block table (-1 if not found)
	attributes_block_entry int64
	// the offset of the (attributes) file block (-1 if not found)
	attributes_block_offset int64
}

func (c *mpq_chunker) insert_section(t mpq_section_type, offset int64) (err error) {
	var section mpq_section
	section.Type = t
	section.Offset = offset

	if offset > c.max_offset {
		c.max_offset = offset
		c.sections = append(c.sections, section)
		return
	}

	i := sort.Search(len(c.sections), func(i int) bool {
		return c.sections[i].Offset >= section.Offset
	})

	if i < len(c.sections) && c.sections[i].Offset == section.Offset {
		// if it's of the same  type, no problem
		if c.sections[i].Type == section.Type {
			return
		}

		if t == mpq_section_file_block && offset == 0 {
			panic(offset)
		}

		err = fmt.Errorf("faws/multipart: mpq_chunker.insert_section: duplicate section: new: %s, current %s, offset %d", t, c.sections[i].Type, offset)
		return
	}

	c.sections = slices.Insert(c.sections, i, section)
	return
}

// // in the case that any of the sections are unmarked, returns an error
// func (c *mpq_chunker) check_unmarked_sessions() (err error) {
// 	return
// }

func (c *mpq_chunker) detect_header() (err error) {
	c.archive_position, err = c.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return
	}
	var signature [4]byte
	if _, err = io.ReadFull(c.file, signature[:]); err != nil {
		return
	}

	if signature != info.HeaderDataSignature {
		err = fmt.Errorf("faws/multipart: MPQ file does not begin with header")
		return
	}

	if err = info.ReadHeader(c.file, &c.archive_header); err != nil {
		return
	}

	if err = c.insert_section(mpq_section_archive_header, c.archive_position); err != nil {
		return
	}

	return
}

func (c *mpq_chunker) detect_hash_table() (err error) {
	rel_hash_table_position := info.HashTablePos(&c.archive_header)

	if rel_hash_table_position == 0 {
		// no hash table
		return
	}

	hash_table_position := int64(uint64(c.archive_position) + rel_hash_table_position)

	_, err = c.file.Seek(hash_table_position, io.SeekStart)
	if err != nil {
		return
	}
	hash_table_length := c.archive_header.HashTableSize
	c.hash_table_data = make([]byte, hash_table_length*info.HashTableEntrySize)
	if _, err = io.ReadFull(c.file, c.hash_table_data); err != nil {
		return
	}
	hash_table_reader := bytes.NewReader(c.hash_table_data)

	// start decrypting
	cipher := crypto.NewBlock(hash_table_decryption_key)

	// look for attributes
	var entry_reader bytes.Reader
	var entry_data [info.HashTableEntrySize]byte
	var entry info.HashTableEntry
	for range hash_table_length {
		if _, err = io.ReadFull(hash_table_reader, entry_data[:]); err != nil {
			return
		}
		for n := 0; n < info.HashTableEntrySize; n += 4 {
			cipher.Decrypt(entry_data[n:], entry_data[n:])
		}
		entry_reader.Reset(entry_data[:])
		if err = info.ReadHashTableEntry(&entry_reader, &entry); err != nil {
			return
		}
		if entry.Name1 == attributes_name_hash_1 && entry.Name2 == attributes_name_hash_2 {
			c.attributes_block_entry = int64(entry.BlockIndex)
			break
		}
	}

	err = c.insert_section(mpq_section_hash_table, hash_table_position)
	return
}

func (c *mpq_chunker) detect_block_table() (err error) {
	if info.BlockTablePos(&c.archive_header) == 0 {
		// no block table
		return
	}

	var hi_block_table []uint16
	if c.archive_header.HiBlockTablePos64 != 0 {
		hi_block_table_data := make([]byte, c.archive_header.BlockTableSize*2)
		hi_block_table = make([]uint16, c.archive_header.BlockTableSize)
		if _, err = c.file.Seek(c.archive_position+int64(c.archive_header.HiBlockTablePos64), io.SeekStart); err != nil {
			return
		}
		if _, err = io.ReadFull(c.file, hi_block_table_data[:]); err != nil {
			return
		}
		for i := range hi_block_table {
			hi_block_table[i] = binary.LittleEndian.Uint16(hi_block_table_data[i*2 : (i+1)*2])
		}
	}

	block_table_position := int64(uint64(c.archive_position) + info.BlockTablePos(&c.archive_header))
	// we want to read the whole block table and decrypt it
	c.block_table_data = make([]byte, c.archive_header.BlockTableSize*info.BlockTableEntrySize)

	if _, err = c.file.Seek(block_table_position, io.SeekStart); err != nil {
		return
	}

	if _, err = io.ReadFull(c.file, c.block_table_data); err != nil {
		return
	}

	if err = c.insert_section(mpq_section_block_table, block_table_position); err != nil {
		return
	}

	// crypto.Decrypt(block_table_decryption_key, c.block_table_data)
	block_table_reader := bytes.NewReader(c.block_table_data)

	cipher := crypto.NewBlock(block_table_decryption_key)

	var entry_reader bytes.Reader
	var entry_data [info.BlockTableEntrySize]byte
	var entry info.BlockTableEntry

	for i := int64(0); i < int64(c.archive_header.BlockTableSize); i++ {
		if _, err = io.ReadFull(block_table_reader, entry_data[:]); err != nil {
			return
		}
		for n := 0; n < info.BlockTableEntrySize; n += 4 {
			cipher.Decrypt(entry_data[n:], entry_data[n:])
		}
		entry_reader.Reset(entry_data[:])
		if err = info.ReadBlockTableEntry(&entry_reader, &entry); err != nil {
			return
		}

		block_position := uint64(entry.Position)
		if len(hi_block_table) != 0 {
			block_position |= (uint64(hi_block_table[i]) << 32)
		}

		if block_position == 0 {
			continue
		}

		real_block_position := c.archive_position + int64(block_position)

		if i == c.attributes_block_entry {
			// if we can detect the attributes, this greatly improves our chances of
			// providing a unique lazy signature
			c.attributes_block_offset = real_block_position
		}

		if err = c.insert_section(mpq_section_file_block, real_block_position); err != nil {
			return
		}
	}

	return
}

func (c *mpq_chunker) detect_het_table() (err error) {
	if c.archive_header.HetTablePos64 == 0 {
		// HET table does not exist
		return
	}

	het_table_position := c.archive_position + int64(c.archive_header.HetTablePos64)
	err = c.insert_section(mpq_section_het_table, het_table_position)
	return
}

func (c *mpq_chunker) detect_bet_table() (err error) {
	if c.archive_header.BetTablePos64 == 0 {
		// BET table does not exist
		return
	}

	bet_table_position := c.archive_position + int64(c.archive_header.BetTablePos64)
	err = c.insert_section(mpq_section_bet_table, bet_table_position)
	if err != nil {
		return
	}

	// read entire BET table at once
	if _, err = c.file.Seek(bet_table_position, io.SeekStart); err != nil {
		return
	}
	bet_table_data := make([]byte, c.archive_header.BetTableSize64)
	if _, err = io.ReadFull(c.file, bet_table_data); err != nil {
		return
	}
	// get MD5 checksum of encrypted + compressed BET table
	bet_table_data_md5 := md5.Sum(bet_table_data[:])
	if bet_table_data_md5 != c.archive_header.MD5_BetTable {
		err = fmt.Errorf("faws/multipart: invalid BET table MD5 checksum")
		return
	}
	// decrypt BET table
	crypto.Decrypt(block_table_decryption_key, bet_table_data[info.ExtTableHeaderSize:])
	var ext_table_header info.ExtTableHeader
	if err = info.ReadExtTableHeader(bytes.NewReader(bet_table_data[:info.ExtTableHeaderSize]), &ext_table_header); err != nil {
		return
	}
	if ext_table_header.Version != 1 {
		err = fmt.Errorf("faws/multipart: BET table extended header: invalid version")
		return
	}
	if ext_table_header.Signature != info.BetTableSignature {
		err = fmt.Errorf("faws/multipart: malformed BET table position")
		return
	}

	var bet_table []byte

	if ext_table_header.Size+info.ExtTableHeaderSize == uint32(len(bet_table_data)) {
		// no decompression required
		bet_table = bet_table_data[info.ExtTableHeaderSize:]
	} else {
		// decompression required
		bet_table, err = compress.Decompress2(bet_table_data[info.ExtTableHeaderSize:])
		if err != nil {
			return
		}
	}

	var bet_table_header info.BetTableHeader

	bet_table_reader := bytes.NewReader(bet_table)

	if err = info.ReadBetTableHeader(bet_table_reader, &bet_table_header); err != nil {
		return
	}

	// skip file flags
	bet_table_reader.Seek(int64(bet_table_header.FlagCount*4), io.SeekCurrent)

	// read BET table entries
	table_entries_size := uint64(bet_table_header.TableEntrySize) * uint64(bet_table_header.EntryCount)
	table_entries := make([]byte, (table_entries_size+7)/8)
	if _, err = io.ReadFull(bet_table_reader, table_entries); err != nil {
		return
	}
	var bet_table_entries bit.Set
	bet_table_entries.Init(table_entries_size, table_entries)

	// skip BET name hash2s
	name_hashes_bits_size := uint64(bet_table_header.BitTotal_NameHash2) * uint64(bet_table_header.EntryCount)
	name_hashes_size := int64((name_hashes_bits_size + 7) / 8)
	bet_table_reader.Seek(name_hashes_size, io.SeekCurrent)

	if bet_table_reader.Len() != 0 {
		err = fmt.Errorf("faws/multipart: error reading BET table: %d bytes unread", bet_table_reader.Len())
		return
	}

	var block_position uint64

	for index := range bet_table_header.EntryCount {
		entry_size := uint64(bet_table_header.TableEntrySize)

		entry_position := entry_size * uint64(index)

		// Read the file position
		block_position, err = bet_table_entries.Uint(entry_position+uint64(bet_table_header.BitIndex_FilePos), uint8(bet_table_header.BitCount_FilePos))
		if err != nil {
			return
		}

		if block_position == 0 {
			continue
		}

		if err = c.insert_section(mpq_section_file_block, c.archive_position+int64(block_position)); err != nil {
			return
		}
	}

	return
}

// scans the file, attempting to identify sections that are largely the same as in previous MPQ versions
// detecting these appropriately leads to efficient data deduplication between MPQs.
func (c *mpq_chunker) detect_sections() (err error) {
	if _, err = c.file.Seek(0, io.SeekStart); err != nil {
		return
	}

	if err = c.detect_header(); err != nil {
		err = fmt.Errorf("error detecting header: %w", err)
		return
	}

	if err = c.detect_hash_table(); err != nil {
		err = fmt.Errorf("error detecting hash table: %w", err)
		return
	}

	if err = c.detect_block_table(); err != nil {
		err = fmt.Errorf("error detecting block table: %w", err)
		return
	}

	if err = c.detect_het_table(); err != nil {
		err = fmt.Errorf("error detecting HET table: %w", err)
		return
	}

	if err = c.detect_bet_table(); err != nil {
		err = fmt.Errorf("error detecting BET table: %w", err)
		return
	}

	for i := 0; i < len(c.sections); i++ {
		var (
			start int64
			end   int64
		)
		start, end, err = c.slice_section(i)
		if err != nil {
			panic(err)
		}
		length := end - start
		if length > c.max_section {
			c.max_section = length
		}
	}

	return
}

func (c *mpq_chunker) Section() string {
	if c.index >= len(c.sections) {
		return ""
	}

	section := &c.sections[c.index]

	return mpq_section_type_names[section.Type]
}

func (c *mpq_chunker) slice_section_bytes(index int) (data []byte, err error) {
	var (
		i, j int64
	)
	i, j, err = c.slice_section(index)
	if err != nil {
		err = fmt.Errorf("faws/multipart: mpq_chunker.slice_section_bytes(%d): %w", index, err)
		return
	}
	data = make([]byte, j-i)
	if _, err = c.file.Seek(i, io.SeekStart); err != nil {
		err = fmt.Errorf("faws/multipart: mpq_chunker.slice_section_bytes(%d): Seek(): %w", index, err)
		return
	}
	_, err = io.ReadFull(c.file, data)
	if err != nil {
		err = fmt.Errorf("faws/multipart: mpq_chunker.slice_section_bytes(%d): io.ReadFull: %w", index, err)
	}
	return
}

func (c *mpq_chunker) slice_section(index int) (i, j int64, err error) {
	if index >= len(c.sections) {
		err = fmt.Errorf("faws/multipart: mpq_chunker.slice_section_bytes(%d): out of bounds", index)
		return
	}

	section := &c.sections[index]
	i = section.Offset

	if index+1 == len(c.sections) {
		j = c.file_size
	} else {
		j = c.sections[index+1].Offset
	}

	return
}

func (c *mpq_chunker) Next() (start int64, data []byte, err error) {
	if c.index >= len(c.sections) {
		err = io.EOF
		return
	}

	if c.chunk_reader != nil {
		section := &c.sections[c.index]
		start, data, err = c.chunk_reader.Next()
		start += section.Offset
		if err != nil && errors.Is(err, io.EOF) {
			c.chunk_reader = nil
			c.index++
			return c.Next()
		}
		return
	}

	var end int64
	start, end, err = c.slice_section(c.index)
	if err != nil {
		return
	}

	length := end - start

	// CDC-chunk the large file
	if length >= min_chunk_size {
		section := &c.sections[c.index]
		c.chunk_reader = new_generic_chunker(io.LimitReader(c.file_reader, length))
		start, data, err = c.chunk_reader.Next()
		if err != nil && errors.Is(err, io.EOF) {
			c.chunk_reader = nil
			c.index++
			return c.Next()
		}
		start += section.Offset
		return
	}

	if end > c.file_size {
		err = fmt.Errorf("read past file end")
		return
	}

	data = make([]byte, length)

	_, err = io.ReadFull(c.file_reader, data[:])
	c.index++
	return
}

func (c *mpq_chunker) LazySignature() (s LazySignature, err error) {
	var offset int64
	offset, err = c.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return
	}

	h := sha256.New()
	var size_bytes [8]byte
	binary.LittleEndian.PutUint64(size_bytes[:], uint64(c.file_size))
	// encode the size of the file
	h.Write(size_bytes[:])

	for i := 0; i < len(c.sections); i++ {
		section := &c.sections[i]
		switch section.Type {
		case mpq_section_archive_header:
			var header_bytes []byte
			header_bytes, err = c.slice_section_bytes(i)
			if err != nil {
				return
			}
			h.Write(header_bytes)
		case mpq_section_hash_table:
			h.Write(c.hash_table_data[:])
		case mpq_section_block_table:
			h.Write(c.block_table_data[:])
		case mpq_section_het_table:
			var het_table_bytes []byte
			het_table_bytes, err = c.slice_section_bytes(i)
			if err != nil {
				return
			}
			h.Write(het_table_bytes[:])
		case mpq_section_bet_table:
			var bet_table_bytes []byte
			bet_table_bytes, err = c.slice_section_bytes(i)
			if err != nil {
				return
			}
			h.Write(bet_table_bytes[:])
		case mpq_section_file_block:
			if section.Offset == c.attributes_block_offset {
				var attributes_file_block_bytes []byte
				attributes_file_block_bytes, err = c.slice_section_bytes(i)
				if err != nil {
					return
				}
				h.Write(attributes_file_block_bytes[:])
			}
		}
	}

	copy(s[:], h.Sum(nil))
	_, err = c.file.Seek(offset, io.SeekStart)
	return
}

func new_mpq_chunker(file io.ReadSeeker, size int64) (c *mpq_chunker, err error) {
	c = new(mpq_chunker)

	c.attributes_block_entry = -1
	c.attributes_block_offset = -1

	c.file = file
	c.file_size = size

	if err = c.detect_sections(); err != nil {
		err = fmt.Errorf("error detection sections: %w", err)
		return
	}

	if _, err = c.file.Seek(0, io.SeekStart); err != nil {
		return
	}

	c.file_reader = bufio.NewReaderSize(c.file, min(int(c.max_section)*2, 256000000))

	return
}
