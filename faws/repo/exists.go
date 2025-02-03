package repo

import (
	"os"
	"path/filepath"
)

// Exists returns whether a repository exists at directory
func Exists(directory string) (exists bool) {
	stat, stat_err := os.Stat(directory)
	if stat_err != nil {
		return
	}
	if !stat.IsDir() {
		return
	}

	// check config file
	var config Config
	stat_err = ReadConfig(filepath.Join(directory, "config"), &config)
	if stat_err != nil {
		return
	}

	exists = true
	return
}
