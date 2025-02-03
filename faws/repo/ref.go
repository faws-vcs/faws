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

func (repo *Repository) ParseRef(ref string) (hash cas.ContentID, err error) {
	if validate.Hex(ref) {
		if len(ref) >= cas.ContentIDSize*2 {
			ref_bytes := []byte(ref)
			hex.Decode(hash[:], ref_bytes[:cas.ContentIDSize*2])
			return
		}
		// ref is a
		hash, err = repo.objects.Deabbreviate(ref)
		return
	}

	// search for tags
	if hash, err = repo.read_tag(ref); err != nil {
		return
	}

	// for i := range repo.commit_manifest.Len() {
	// 	commit_hash, commit_index_err := repo.commit_manifest.Index(i)
	// 	if commit_index_err == nil {
	// 		_, commit_info, commit_check_err := repo.check_commit(commit_hash)
	// 		if commit_check_err == nil {
	// 			if commit_info.Tag == ref {
	// 				// found it
	// 				hash = commit_hash
	// 				return
	// 			}
	// 		}
	// 	}
	// }

	return
}
