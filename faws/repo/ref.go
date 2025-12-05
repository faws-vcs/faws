package repo

import (
	"encoding/hex"
	"fmt"

	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/validate"
)

var (
	ErrRefNotFound = fmt.Errorf("faws/repo: ref not found")
)

// ParseRef returns a hash from a string, which may be either an [abbreviated] hexadecimal object hash, or a commit tag
func (repo *Repository) ParseRef(ref string) (hash cas.ContentID, err error) {
	ref_is_valid_hex := validate.Hex(ref)

	// abbreviated hashes
	if ref_is_valid_hex && len(ref) == cas.ContentIDSize*2 {
		goto parse_hex
	}

	// search for tags
	if hash, err = repo.read_tag(ref); err == nil {
		// tag is valid!
		return
	}

	if !ref_is_valid_hex {
		err = ErrBadRef
		return
	}

parse_hex:
	if len(ref) >= cas.ContentIDSize*2 {
		ref_bytes := []byte(ref)
		hex.Decode(hash[:], ref_bytes[:cas.ContentIDSize*2])
		return
	}
	// ref is an abbreviation
	hash, err = repo.objects.Deabbreviate(ref)
	return
}
