package identities

import (
	"errors"

	"github.com/faws-vcs/faws/faws/app"
	"github.com/faws-vcs/faws/faws/identity"
)

// RingTrust implements [github.com/faws-vcs/faws/faws/repo.Trust]
//
// It's a permissive trust mechanism by default, as it will always accept untrusted identities. It only violates trust when an external ID purports to have the same nametag as one already in the user's ring
type RingTrust struct {
	ring *identity.Ring
}

// NewRingTrust creates a new RingTrust using the user's ring, trusting on first use when there is no nametag conflict
func NewRingTrust(ring *identity.Ring) (trust *RingTrust) {
	trust = new(RingTrust)
	trust.ring = ring
	return
}

func (trust *RingTrust) Check(id identity.ID, signed_attributes *identity.Attributes) (trusted bool) {
	var previously_signed_attributes identity.Attributes
	var trust_err error
	trust_err = trust.ring.GetTrustedAttributes(id, &previously_signed_attributes)
	trusted = trust_err == nil
	if trust_err != nil && errors.Is(trust_err, identity.ErrRingKeyNotFound) {
		// lookup nametag.
		id_for_nametag, err := trust.ring.Lookup(signed_attributes.Nametag)
		if err == nil {
			app.Warning("Mismatch in ID for nametag", "'"+signed_attributes.Nametag+"'")
			app.Warning("The ID in your ring is", id_for_nametag)
			app.Warning("The ID claimed by this commit is", id_for_nametag)
			app.Warning("This can indicate a supply-chain attack. rejecting this identity")
			trusted = false
			return
		} else if errors.Is(err, identity.ErrRingKeyNotFound) {
			app.Warning("user", signed_attributes.Nametag, "("+id.String()+")", "is automatically imported into your identity ring")
			// automatically import the key on first use.
			// this is not great but better than disabling identity verification completely
			trusted = true
		} else if errors.Is(err, identity.ErrRingNoNametag) {
			trusted = true
		} else {
			// something weird is going on
			panic(err)
		}
	}

	if trusted {
		if trust_err = trust.ring.TrustAttributes(id, signed_attributes); trust_err != nil {
			if !errors.Is(trust_err, identity.ErrRingCannotTrustSecretKey) {
				app.Warning(trust_err)
			}
		}
	}

	return
}
