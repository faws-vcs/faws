package repo

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/faws-vcs/faws/faws/app/about"
	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/repo/remote"
)

// Initialize a repository at the directory.
// if reinitialize == true, you are allowed to refresh an existing repository with updated basics.
// if origin_url != "", you start to pull repository information from the remote repository at origin_url
func Initialize(directory string, origin_url string, reinitialize, force bool) (err error) {
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
	config.AppID = "faws"
	config.AppVersion = about.GetVersionString()
	config.RepositoryFormat = Format

	config_name := filepath.Join(directory, "config")
	if _, err = os.Stat(config_name); err == nil {
		config = Config{}
		if err = ReadConfig(config_name, &config); err != nil {
			return
		}
	}

	// Detect the format of the repository by reading the remote config
	// The rest of the config is not important to us (for now)
	if origin_url != "" {
		remote_fs, err = remote.Open(origin_url)
		if err != nil {
			return
		}
		var config_file io.ReadCloser
		config_file, err = remote_fs.Pull("config")
		if err != nil {
			err = fmt.Errorf("faws/repo: remote repository does not exist")
			return
		}
		var remote_config Config
		d := json.NewDecoder(config_file)
		err = d.Decode(&remote_config)
		if err != nil {
			err = fmt.Errorf("faws/repo: cannot decode remote repository's config: %w", err)
			return
		}
		config.RepositoryFormat = remote_config.RepositoryFormat
		// cleaned URL (for instance, if origin_url was a filepath, now it has file: URI)
		config.Origin = remote_fs.URL()
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
