package cas

import (
	"errors"
	"strings"

	"github.com/faws-vcs/faws/faws/validate"
)

// Deabbreviate expands a hash abbreviation string. It will attempt to disambiguate even the shortest possible hexadecimal string.
// If the abbreviation is not hexadecimal, err will be [ErrAbbreviationNotHex]
// If multiple candidates exist for an abbreviation, err will be [ErrAbbreviationAmbiguous]
// If the abbreviation does not fit with any object in the Set, err will be [ErrObjectNotFound]

func (set *Set) Deabbreviate(abbreviation string) (content_id ContentID, err error) {
	abbreviation = strings.ToLower(abbreviation)
	if len(abbreviation) == 0 {
		err = ErrAbbreviationTooShort
		return
	}
	if !validate.Hex(abbreviation) {
		err = ErrAbbreviationNotHex
		return
	}

	// the result depends on both the cache and the pack
	cache_content_id, cache_err := set.cache.Deabbreviate(abbreviation)
	if cache_err != nil {
		// if the cache finds the abbreviation is ambiguous, then there is no point to searching the pack
		if errors.Is(cache_err, ErrAbbreviationAmbiguous) {
			err = cache_err
			return
		}
	}

	pack_content_id, pack_err := set.pack.Deabbreviate(abbreviation)
	if pack_err != nil {
		if errors.Is(pack_err, ErrAbbreviationAmbiguous) {
			err = pack_err
			return
		}
	}
	if pack_err == nil && cache_err == nil {
		if cache_content_id != pack_content_id {
			// if the abbreviation leads to two different content IDs in the cache and the pack, then it is ambiguous
			err = ErrAbbreviationAmbiguous
		} else {
			// they debbreviated to the same ID. while this shouldn't ordinarily happen, it's technically valid.
			content_id = cache_content_id
		}
		return
	}

	if cache_err == nil && pack_err != nil {
		// the cache has the only valid id
		content_id, err = cache_content_id, nil
		return
	}

	if pack_err == nil && cache_err != nil {
		// the pack has the only valid id
		content_id, err = pack_content_id, nil
		return
	}

	// neither are valid
	err = ErrObjectNotFound

	return
}
