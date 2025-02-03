package config

import (
	"path/filepath"

	"github.com/faws-vcs/faws/faws/identity"
)

func (config *Configuration) ring_name() string {
	return filepath.Join(config.directory, "identity_ring")
}

func (config *Configuration) SaveRing() (err error) {
	ring_name := config.ring_name()

	if err = identity.SaveRing(ring_name, &config.ring); err != nil {
		return
	}

	return
}

func (config *Configuration) Ring() *identity.Ring {
	return &config.ring
}
