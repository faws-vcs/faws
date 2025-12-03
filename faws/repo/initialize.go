package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/repo/remote"
)

// Initialize a repository at the directory.
// if reinitialize == true, you are allowed to refresh an existing repository with updated basics.
// if remote_url != "", you create a reposito
func Initialize(directory string, remote_url string, reinitialize, force bool) (err error) {
	if Exists(directory) && !reinitialize {
		err = ErrInitializeCannotExist
		return
	}

	// create directory if it doesn't exist
	if _, not_found := os.Stat(directory); not_found != nil {
		err = os.Mkdir(directory, fs.DefaultPublicDirPerm)
		if err != nil {
			return
		}
	} else {
		var dir_entries []os.DirEntry
		dir_entries, err = os.ReadDir(directory)
		if err != nil {
			return
		}
		if len(dir_entries) > 0 {
			if !reinitialize && !force {
				err = ErrRepoCannotInitializeNonEmptyDirectory
				return
			}
		}
	}

	var remote_fs remote.Fs
	// create config
	var config Config
	config.Version = 1

	config_name := filepath.Join(directory, "config")
	if _, err = os.Stat(config_name); err == nil {
		config = Config{}
		if err = ReadConfig(config_name, &config); err != nil {
			return
		}
	}

	// Retrieve the remote's config
	if remote_url != "" {
		remote_fs, err = remote.Open(remote_url)
		if err != nil {
			return
		}
		var config_file io.ReadCloser
		config_file, err = remote_fs.Pull("config")
		if err != nil {
			err = fmt.Errorf("faws/repo: remote repository does not exist")
			return
		}
		config = Config{}
		d := json.NewDecoder(config_file)
		err = d.Decode(&config)
		if err != nil {
			return
		}

		config.Remote = remote_fs.URL()
	}

	err = WriteConfig(config_name, &config)
	if err != nil {
		return
	}

	// create tags
	tags_directory := filepath.Join(directory, "tags")
	if _, not_found := os.Stat(tags_directory); not_found != nil {
		err = os.Mkdir(tags_directory, fs.DefaultPublicDirPerm)
		if err != nil {
			return
		}
	}

	return
}
