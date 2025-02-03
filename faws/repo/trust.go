package repo

import "github.com/faws-vcs/faws/faws/identity"

func WithTrust(trust Trust) Option {
	return func(repo *Repository) {
		repo.trust = trust
	}
}

type Trust interface {
	// Checks if ID is trusted. If so, signed_attributes are committed into keyring
	Check(id identity.ID, signed_attributes *identity.Attributes) (trusted bool)
}
