package cas

// Set is the set of all objects held by the repository and index
type Set struct {
	// the location of the cas.Set. this never contains a trailing slash
	directory string
	cache     cache
	pack      Pack
}
