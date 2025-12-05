package config

import (
	"path/filepath"

	"github.com/faws-vcs/faws/faws/identity"
)

// RingPath returns the path to the user's [identity.Ring] file
func (config *Configuration) RingPath() string {
	return filepath.Join(config.directory, "identity_ring")
}

// Ring returns the user's identity ring
func (config *Configuration) Ring() *identity.Ring {
	return &config.ring
}
