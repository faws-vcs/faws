package repo

import (
	"os"
	"path/filepath"

	"github.com/faws-vcs/faws/faws/app/about"
	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/repo/config"
	"github.com/faws-vcs/faws/faws/repo/p2p/tracker"
	"github.com/faws-vcs/faws/faws/repo/remote"
	"github.com/google/uuid"
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

	// create config
	var config_ config.Config
	config_.AppID = "faws"
	config_.AppVersion = about.GetVersionString()
	config_.RepositoryFormat = config.Format

	config_name := filepath.Join(directory, "config")
	if _, err = os.Stat(config_name); err == nil {
		config_ = config.Config{}
		if err = config.ReadConfig(config_name, &config_); err != nil {
			return
		}
	}

	// Detect the format of the repository by reading the remote config
	// The rest of the config is not important to us (for now)
	if origin_url == "" {
		if config_.UUID == uuid.Nil {
			config_.UUID = uuid.New()
		}
	} else {
		// special handling for topic URI
		if tracker.IsTopicURI(origin_url) {
			var topic tracker.Topic
			err = tracker.ParseTopicURI(origin_url, &topic)
			if err != nil {
				return
			}
			config_.Origin = origin_url
			// UUID is embedded in the topic URI
			config_.UUID = topic.Repository
		} else {
			// treat as nominal origin
			var origin remote.Origin
			origin, err = remote.Open(origin_url)
			if err != nil {
				return
			}

			// read UUID from remote
			config_.UUID, err = origin.UUID()
			if err != nil {
				return
			}
			// cleaned URL (for instance, if origin_url was a filepath, now it has file: URI)
			config_.Origin = origin.URI()
		}
	}

	err = config.WriteConfig(config_name, &config_)
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
