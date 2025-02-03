package multipart

import (
	"fmt"

	"github.com/shirou/gopsutil/v4/mem"
)

const default_bs = 16384

// returns a good buffering size for the current machine
func good_buffer_size() (bs int) {
	bs = default_bs
	v, err := mem.VirtualMemory()
	if err != nil {
		return
	}
	bs = min(int(v.Free/32), int(v.Total/64))

	fmt.Println("buffer size", bs)
	return
}
