package config

import (
	"os"

	"github.com/faws-vcs/faws/faws/identity"
)

type Configuration struct {
	directory string
	ring      identity.Ring
}

func (config *Configuration) Open(directory string) (err error) {
	// Ensure directory exists
	if _, stat_err := os.Stat(directory); stat_err != nil {
		err = os.MkdirAll(directory, os.ModePerm)
		if err != nil {
			return
		}
	}

	config.directory = directory

	// load ring or create empty ring
	ring_name := config.ring_name()
	if _, stat_err := os.Stat(ring_name); stat_err == nil {
		if err = identity.LoadRing(ring_name, &config.ring); err != nil {
			return
		}
	} else {
		if err = identity.SaveRing(ring_name, &config.ring); err != nil {
			return
		}
	}

	return
}
