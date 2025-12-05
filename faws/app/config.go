package app

import (
	"os"
	"path/filepath"

	"github.com/faws-vcs/console"
	"github.com/faws-vcs/faws/faws/config"
)

// The user's local Faws configuration (not accessible until [Open] is called)
var Configuration config.Configuration

// Opens the console, and loads the user's local Faws configuration
func Open() {
	console.Open()

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

// Close terminates the program and saves any changes made to the configuration
func Close() {
	if err := Configuration.Close(); err != nil {
		Fatal(err)
	}

	console.Close()
}
