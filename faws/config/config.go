package config

import (
	"fmt"
	"os"

	"github.com/faws-vcs/faws/faws/fs"
	"github.com/faws-vcs/faws/faws/identity"
)

type Configuration struct {
	directory string
	ring      identity.Ring
}

func (config *Configuration) Open(directory string) (err error) {
	// Ensure directory exists
	if _, stat_err := os.Stat(directory); stat_err != nil {
		err = os.MkdirAll(directory, fs.DefaultPerm)
		if err != nil {
			return
		}
	}

	config.directory = directory

	// load ring or create empty ring
	ring_name := config.RingPath()
	if _, stat_err := os.Stat(ring_name); stat_err == nil {
		if err = identity.ReadRing(ring_name, &config.ring); err != nil {
			return
		}
	} else {
		fmt.Println("not loading", ring_name)
	}

	return
}

func (config *Configuration) Close() (err error) {
	if err = identity.WriteRing(config.RingPath(), &config.ring); err != nil {
		return
	}

	return
}
