package config

import (
	"path/filepath"

	"github.com/faws-vcs/faws/faws/identity"
)

func (config *Configuration) RingPath() string {
	return filepath.Join(config.directory, "identity_ring")
}

func (config *Configuration) Ring() *identity.Ring {
	return &config.ring
}
