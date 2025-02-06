package app

import (
	"os"
	"path/filepath"

	"github.com/faws-vcs/faws/faws/config"
)

var Configuration config.Configuration

func Open() {
	directory := os.Getenv("FAWS_CONFIG")

	if directory == "" {
		user_config_directory, err := os.UserConfigDir()
		if err != nil {
			Fatal(err)
		}
		directory = filepath.Join(user_config_directory, "faws")
	}

	if err := Configuration.Open(directory); err != nil {
		Fatal(err)
	}
}

func Close() {
	if err := Configuration.Close(); err != nil {
		Fatal(err)
	}
}
