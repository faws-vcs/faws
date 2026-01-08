package repo

import (
	"os"
	"path/filepath"

	"github.com/faws-vcs/faws/faws/repo/config"
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
	var repo_config config.Config
	stat_err = config.ReadConfig(filepath.Join(directory, "config"), &repo_config)
	if stat_err != nil {
		return
	}

	exists = true
	return
}
