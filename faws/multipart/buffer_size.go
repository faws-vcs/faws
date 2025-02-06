package multipart

const default_bs = 16384

// returns a good buffering size for the current machine
func good_buffer_size() (bs int) {
	bs = default_bs * 1000
	return
}
