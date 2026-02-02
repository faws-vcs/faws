package about

import (
	"encoding/binary"
	"fmt"
	"os"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/repo/cas"
)

type InspectFileParams struct {
	Name string
}

func InspectFile(params *InspectFileParams) {
	app.Open()
	defer func() {
		app.Close()
	}()

	file, err := os.Open(params.Name)
	if err != nil {
		app.Fatal(err)
	}

	var prefix cas.Prefix
	if _, err = file.ReadAt(prefix[:], 0); err != nil {
		panic(err)
	}
	switch prefix {
	case cas.Prefix{'I', 'N', 'D', 'X'}:
		inspect_pack_index(file)
	default:
		app.Fatal("unknown file type")
	}

	file.Close()
}

func inspect_pack_index(file *os.File) {
	app.Header("pack index file")
	var fanout_table [256 * 8]byte
	if _, err := file.ReadAt(fanout_table[:], cas.PrefixSize); err != nil {
		app.Fatal(err)
	}
	app.Info("")
	app.Info("--")
	app.Info("")
	app.Header("fanout table:")
	for i := range 256 {
		fanout_value := binary.LittleEndian.Uint64(fanout_table[i*8 : (i+1)*8])
		app.Info(fmt.Sprintf("%02x %d", i, fanout_value))
	}
	app.Info("")
	app.Info("--")
	app.Info("")
	app.Header("index entries:")
	index_info, _ := file.Stat()
	size := index_info.Size()
	const header_size int64 = (cas.PrefixSize + (256 * 8))
	const entry_size int64 = 4 + 8 + cas.ContentIDSize
	entry_count := (size - header_size) / entry_size

	for i := int64(0); i < entry_count; i++ {
		var entry_data [entry_size]byte
		if _, err := file.ReadAt(entry_data[:], header_size+(entry_size*i)); err != nil {
			app.Fatal(err)
		}

		var (
			archive_id  uint32
			file_offset int64
			name        cas.ContentID
		)
		archive_id = binary.LittleEndian.Uint32(entry_data[0:4])
		file_offset = int64(binary.LittleEndian.Uint64(entry_data[4:12]))
		copy(name[:], entry_data[12:])

		app.Info("--")
		app.Info("archive id", archive_id)
		app.Info("file offset", file_offset)
		app.Info("name", name)
	}
}
