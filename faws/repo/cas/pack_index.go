package cas

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/faws-vcs/faws/faws/fs"
)

const (
	pack_index_header_size = PrefixSize + (256 * 8)
	pack_index_entry_size  = 4 + 8 + ContentIDSize
)

type pack_index_header struct {
	Prefix      Prefix
	FanoutTable [256]uint64
}

type pack_index_entry struct {
	// which pack file is the object stored in?
	ArchiveID int
	//
	FileOffset int64
	//
	Name ContentID
}

//

type pack_index struct {
	guard     sync.RWMutex
	header    pack_index_header
	file      *os.File
	file_size int64
}

func (pack_index *pack_index) write_header() (err error) {
	writer := io.NewOffsetWriter(pack_index.file, 0)

	if _, err = writer.Write(pack_index.header.Prefix[:]); err != nil {
		return
	}
	var fanout_table [8 * 256]byte
	for i := range 256 {
		binary.LittleEndian.PutUint64(fanout_table[i*8:], pack_index.header.FanoutTable[i])
	}

	if _, err = writer.Write(fanout_table[:]); err != nil {
		return
	}
	return
}

func (pack_index *pack_index) read_header() (err error) {
	_, err = io.ReadFull(pack_index.file, pack_index.header.Prefix[:])
	if err != nil {
		return
	}

	if pack_index.header.Prefix != index_prefix {
		err = ErrInvalidPackIndexFile
		return
	}

	var fanout_table [8 * 256]byte
	if _, err = io.ReadFull(pack_index.file, fanout_table[:]); err != nil {
		return
	}
	for i := range 256 {
		pack_index.header.FanoutTable[i] = binary.LittleEndian.Uint64(fanout_table[i*8:])
	}
	return
}

func (pack_index *pack_index) num_entries() int64 {
	return (pack_index.file_size - pack_index_header_size) / pack_index_entry_size
}

func (pack_index *pack_index) Open(name string) (err error) {
	pack_index.file, err = os.OpenFile(name, os.O_CREATE|os.O_RDWR, fs.DefaultPublicPerm)
	if err != nil {
		return
	}

	pack_index.file_size, err = pack_index.file.Seek(0, io.SeekEnd)
	if err != nil {
		return
	}
	_, err = pack_index.file.Seek(0, io.SeekStart)
	if err != nil {
		return
	}
	if pack_index.file_size < pack_index_header_size {
		pack_index.header.Prefix = index_prefix
		err = pack_index.write_header()
		if err != nil {
			return
		}
		pack_index.file_size = pack_index_header_size
	} else {
		err = pack_index.read_header()
	}

	return
}

func search(n int64, f func(int64) bool) int64 {
	// Define f(-1) == false and f(n) == true.
	// Invariant: f(i-1) == false, f(j) == true.
	i, j := int64(0), n
	for i < j {
		h := int64(uint64(i+j) >> 1) // avoid overflow when computing h
		// i ≤ h < j
		if !f(h) {
			i = h + 1 // preserves f(i-1) == false
		} else {
			j = h // preserves f(j) == true
		}
	}
	// i == j, f(i-1) == false, and f(j) (= f(i)) == true  =>  answer is i.
	return i
}

func (pack_index *pack_index) write_entry(index int64, entry pack_index_entry) (err error) {
	entry_start := pack_index_header_size + (index * pack_index_entry_size)
	entry_end := entry_start + pack_index_entry_size

	var entry_bytes [pack_index_entry_size]byte
	binary.LittleEndian.PutUint32(entry_bytes[0:4], uint32(entry.ArchiveID))
	binary.LittleEndian.PutUint64(entry_bytes[4:12], uint64(entry.FileOffset))
	copy(entry_bytes[12:], entry.Name[:])

	if _, err = pack_index.file.WriteAt(entry_bytes[:], entry_start); err != nil {
		return
	}

	if entry_end > pack_index.file_size {
		pack_index.file_size = entry_end
	}

	return
}

func (pack_index *pack_index) read_entry(index int64) (entry pack_index_entry, err error) {
	if index >= pack_index.num_entries() {
		err = ErrInvalidPackIndexFile
		return
	}

	// var entry_bytes [pack_index_entry_size]byte
	entry_bytes := make([]byte, pack_index_entry_size)
	if _, err = pack_index.file.ReadAt(entry_bytes, pack_index_header_size+(pack_index_entry_size*index)); err != nil {
		err = fmt.Errorf("cas: in reading pack index entry %d: %w", index, err)
		return
	}

	entry.ArchiveID = int(binary.LittleEndian.Uint32(entry_bytes[:4]))
	entry_bytes = entry_bytes[4:]

	entry.FileOffset = int64(binary.LittleEndian.Uint64(entry_bytes[:8]))
	entry_bytes = entry_bytes[8:]

	copy(entry.Name[:], entry_bytes)

	return
}

func (pack_index *pack_index) swap_entry(i, j int64) (err error) {
	var (
		i_entry, j_entry pack_index_entry
	)
	i_entry, err = pack_index.read_entry(i)
	if err != nil {
		return
	}
	j_entry, err = pack_index.read_entry(j)
	if err != nil {
		return
	}
	err = pack_index.write_entry(i, j_entry)
	if err != nil {
		return
	}
	err = pack_index.write_entry(j, i_entry)
	return
}

func (pack_index *pack_index) add_fanout_value(bucket byte) (err error) {
	for n := int(bucket); n < 256; n++ {
		pack_index.header.FanoutTable[n]++
	}
	err = pack_index.write_header()
	return
}

func (pack_index *pack_index) remove_fanout_value(bucket byte) (err error) {
	for n := int(bucket); n < 256; n++ {
		if pack_index.header.FanoutTable[n] == 0 {
			err = ErrInvalidPackIndexFile
			return
		}
		pack_index.header.FanoutTable[n]--
	}
	err = pack_index.write_header()
	return
}

func (pack_index *pack_index) Put(entry pack_index_entry) (err error) {
	index := search(pack_index.num_entries(), func(i int64) bool {
		entry_i, err := pack_index.read_entry(i)
		if err != nil {
			panic(err)
		}
		return !entry_i.Name.Less(entry.Name)
	})
	if index >= pack_index.num_entries() {
		// since this is a new entry, adjust the fanout table
		if err = pack_index.add_fanout_value(entry.Name[0]); err != nil {
			return
		}
		// insert at the tail-end
		err = pack_index.write_entry(index, entry)
		return
	}
	// see if there's a current entry that we can overwrite
	var current_entry pack_index_entry
	current_entry, err = pack_index.read_entry(index)
	if err != nil {
		return
	}
	if current_entry.Name == entry.Name {
		// we're replacing this entry (this should probably not ever happen)
		err = pack_index.write_entry(index, entry)
		return
	}

	tail := pack_index.num_entries()

	// write the entry at the tail end
	err = pack_index.write_entry(tail, entry)
	if err != nil {
		return
	}
	// but, swap all preceding elements so that it ends up at the index with all higher elements shifted up
	for i := tail; i > index; i-- {
		err = pack_index.swap_entry(i, i-1)
		if err != nil {
			return
		}
	}
	// since this is a new entry, adjust the fanout table
	if err = pack_index.add_fanout_value(entry.Name[0]); err != nil {
		return
	}

	// TODO: remove this (defensive programming)
	chk_index, chk_err := pack_index.read_entry(index)
	if chk_err != nil {
		panic(err)
	}
	if chk_index.Name != entry.Name {
		panic(chk_index.Name)
	}
	return
}

func (pack_index *pack_index) Delete(name ContentID) (err error) {
	pack_index.guard.Lock()
	defer pack_index.guard.Unlock()
	num_fanout_entries := pack_index.header.FanoutTable[int(name[0])]
	if num_fanout_entries == 0 {
		err = object_error{ErrObjectNotFound, name}
		return
	}

	num_entries := pack_index.num_entries()

	// we need the actual index, so let's search through the entire file here
	index := search(num_entries, func(i int64) bool {
		entry, err := pack_index.read_entry(i)
		if err != nil {
			panic(err)
		}
		return !entry.Name.Less(name)
	})

	if index >= num_entries {
		err = object_error{ErrObjectNotFound, name}
		return
	}

	var index_entry pack_index_entry
	index_entry, err = pack_index.read_entry(index)
	if err != nil {
		return
	}

	if index_entry.Name != name {
		err = object_error{ErrObjectNotFound, name}
		return
	}

	// shuffle entries so that the deleted entry is swapped to the end of the file
	for i := index + 1; i < num_entries; i++ {
		if err = pack_index.swap_entry(i-1, i); err != nil {
			return
		}
	}

	// then, truncate, discarding the entry.
	err = pack_index.file.Truncate(pack_index_header_size + ((num_entries - 1) * pack_index_entry_size))
	return
}

func (pack_index *pack_index) Get(name ContentID) (entry pack_index_entry, err error) {
	pack_index.guard.RLock()
	defer pack_index.guard.RUnlock()
	num_fanout_entries := pack_index.header.FanoutTable[int(name[0])]
	var previous_fanout_entry_count uint64
	if name[0] != 0x00 {
		previous_fanout_entry_count = pack_index.header.FanoutTable[name[0]-1]
	}
	if num_fanout_entries == previous_fanout_entry_count {
		err = object_error{ErrObjectNotFound, name}
		return
	}
	// narrow down binary search using fanout table
	upper := int64(pack_index.header.FanoutTable[name[0]])
	var lower int64
	if name[0] != 0x00 {
		lower = int64(pack_index.header.FanoutTable[name[0]-1])
	}
	bucket_entries := upper - lower
	// perform the binary search on a bucket-subset of the index
	index := search(bucket_entries, func(i int64) bool {
		entry, err := pack_index.read_entry(lower + i)
		if err != nil {
			panic(err)
		}
		return !entry.Name.Less(name)
	})
	if index >= bucket_entries {
		err = object_error{ErrObjectNotFound, name}
		return
	}
	entry, err = pack_index.read_entry(lower + index)
	if err != nil {
		return
	}
	if entry.Name != name {
		err = object_error{ErrObjectNotFound, name}
		return
	}
	return
}

// create search parameters for the abbreviation
// if abbreviation == "a"
// then
// min_value == "a000000000000000000000000000000000000000"
// max_value == "afffffffffffffffffffffffffffffffffffffff"
// therefore, any deabbrevation where the actual hash >= min_value && <= max_value is valid
func compute_abbreviation_range(text []byte) (min, max ContentID, err error) {
	var min_text, max_text [ContentIDSize * 2]byte
	var i int
	for ; i < len(text); i++ {
		min_text[i] = text[i]
		max_text[i] = text[i]
	}
	for ; i < ContentIDSize*2; i++ {
		min_text[i] = '0'
		max_text[i] = 'F'
	}
	_, err = hex.Decode(min[:], min_text[:])
	if err != nil {
		return
	}
	_, err = hex.Decode(max[:], max_text[:])
	if err != nil {
		return
	}

	return
}

func (pack_index *pack_index) Deabbreviate(abbreviation string) (name ContentID, err error) {
	if len(abbreviation) < 1 {
		err = ErrAbbreviationTooShort
		return
	}

	var min_value, max_value ContentID
	min_value, max_value, err = compute_abbreviation_range([]byte(abbreviation))
	if err != nil {
		return
	}
	// we'll need to decide where to start searching the index
	var bucket_first, bucket_last byte

	//
	if len(abbreviation) < 2 {
		// this won't work unless the pack is very small
		// only 1 hex digit is available. this means our range must include 16 buckets
		var bucket_high uint64
		bucket_high, err = strconv.ParseUint(abbreviation, 16, 4)
		if err != nil {
			// we already determined that the abbreviation was hexadecimal, so this should never occur
			panic(err)
		}
		// for example if the abbreviation was just 'a'
		// bucket_first == a0
		// bucket_last  == af
		bucket_first = byte(bucket_high) << 4
		bucket_last = bucket_first | 0xF
	} else {
		var bucket uint64
		bucket, err = strconv.ParseUint(abbreviation[:2], 16, 8)
		if err != nil {
			// we already determined that the abbreviation was hexadecimal, so this should never occur
			panic(err)
		}
		// the bucket is already included
		bucket_first = byte(bucket)
		bucket_last = byte(bucket)
	}

	pack_index.guard.RLock()
	defer pack_index.guard.RUnlock()

	// using our buckets, let's decide a range of index entries to search.
	var lower, upper int64
	if bucket_first != 0x00 {
		lower = int64(pack_index.header.FanoutTable[bucket_first-1])
	}
	upper = int64(pack_index.header.FanoutTable[bucket_last])
	limited_entries := upper - lower
	if limited_entries == 0 {
		// there are no entries in the range.
		err = ErrObjectNotFound
		return
	}
	total_entries := pack_index.num_entries()

	// now, we narrow down the search by looking for where the minimum and maximum value would be.
	lower_narrow := lower + search(limited_entries, func(i int64) bool {
		entry, err := pack_index.read_entry(lower + i)
		if err != nil {
			panic(err)
		}
		return !entry.Name.Less(min_value)
	})

	if lower_narrow >= upper {
		err = ErrObjectNotFound
		return
	}

	narrow_entries := upper - lower_narrow

	upper_narrow := lower_narrow + search(narrow_entries, func(i int64) bool {
		entry, err := pack_index.read_entry(lower_narrow + i)
		if err != nil {
			panic(err)
		}
		return !entry.Name.Less(max_value)
	})
	if upper_narrow >= total_entries {
		upper_narrow = total_entries - 1
	}

	hit := false

	// now bruteforce the remaining tranche
	for i := lower_narrow; i <= upper_narrow; i++ {
		var index_entry pack_index_entry
		index_entry, err = pack_index.read_entry(i)
		if err != nil {
			return
		}

		min_cmp := bytes.Compare(index_entry.Name[:], min_value[:])
		max_cmp := bytes.Compare(index_entry.Name[:], max_value[:])

		if min_cmp >= 0 && max_cmp <= 0 {
			if hit {
				// we already encountered a satisfactory deabbreviation, so the abbreviation is ambiguous
				err = ErrAbbreviationAmbiguous
				return
			}
			hit = true
			name = index_entry.Name
		}
	}

	if !hit {
		err = ErrObjectNotFound
		return
	}

	return
}

func (pack_index *pack_index) List(fn ListFunc) (err error) {
	pack_index.guard.RLock()
	defer pack_index.guard.RUnlock()
	var entry pack_index_entry
	for i := int64(0); i < pack_index.num_entries(); i++ {
		entry, err = pack_index.read_entry(i)
		if err != nil {
			return
		}
		if err = fn(true, entry.Name); err != nil {
			return
		}
	}
	return
}

// returns the size of the index file, in bytes
func (pack_index *pack_index) Size() (n int64) {
	pack_index.guard.RLock()
	n = pack_index.file_size
	pack_index.guard.RUnlock()
	return
}

func (pack_index *pack_index) Close() (err error) {
	pack_index.file_size = 0
	pack_index.header = pack_index_header{}
	err = pack_index.file.Close()
	return
}
