package repo

import "github.com/faws-vcs/faws/faws/identity"

// WithTrust is an [Option] to open a [Repository] with a particular trust mechanism
func WithTrust(trust Trust) Option {
	return func(repo *Repository) {
		repo.trust = trust
	}
}

// Trust defines the parameters of a custom trust mechanism, which may vary wildly depending on the user's application
type Trust interface {
	// Checks if ID is trusted. If so, signed_attributes are committed into keyring
	Check(id identity.ID, signed_attributes *identity.Attributes) (trusted bool)
}
