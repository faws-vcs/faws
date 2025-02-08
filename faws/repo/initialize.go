package repo

import (
	"os"
	"path/filepath"

	"github.com/faws-vcs/faws/faws/fs"
)

// Initialize an empty repository at the directory.
// if reinitialize == true, you are allowed to refresh an existing repository with updated basics.
func Initialize(directory string, reinitialize bool) (err error) {
	if Exists(directory) && !reinitialize {
		err = ErrInitializeCannotExist
		return
	}

	// create directory if it doesn't exist
	if _, not_found := os.Stat(directory); not_found != nil {
		err = os.Mkdir(directory, fs.DefaultPerm)
		if err != nil {
			return
		}
	}

	// create config
	config_file := filepath.Join(directory, "config")
	if _, not_found := os.Stat(config_file); not_found != nil {
		err = WriteConfig(config_file, &Config{
			Version: 1,
		})
		if err != nil {
			return
		}
	}

	// create tags
	tags_directory := filepath.Join(directory, "tags")
	if _, not_found := os.Stat(tags_directory); not_found != nil {
		err = os.Mkdir(tags_directory, fs.DefaultPerm)
		if err != nil {
			return
		}
	}

	return
}
