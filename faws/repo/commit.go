package repo

import (
	"fmt"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/faws-vcs/faws/faws/repo/revision"
	"github.com/faws-vcs/faws/faws/validate"
)

// reads and verifies a commit in the repository.
func (repo *Repository) check_commit(revision_hash cas.ContentID) (author identity.ID, info *revision.CommitInfo, err error) {
	var (
		prefix        cas.Prefix
		commit_record []byte
	)
	prefix, commit_record, err = repo.objects.Load(revision_hash)
	if err != nil {
		return
	}
	if prefix != cas.Commit {
		err = ErrCommitInvalidPrefix
		return
	}
	var commit revision.Commit
	err = revision.UnmarshalCommit(commit_record, &commit)
	if err != nil {
		return
	}

	// verify signature
	if !identity.Verify(commit.Author, &commit.Signature, commit.Info) {
		err = fmt.Errorf("%w: commit signature from %s is a failure", ErrBadCommit, commit.Author)
		return
	}

	// unmarshal commit info
	info = new(revision.CommitInfo)
	err = revision.UnmarshalCommitInfo(commit.Info, info)
	if err != nil {
		return
	}

	// finally, check the author's identity. (invokes user prompt typically)
	if !repo.trust.Check(commit.Author, &info.AuthorAttributes) {
		err = fmt.Errorf("%w: %s", ErrCommitAuthorNotTrusted, commit.Author)
		return
	}

	// everything is in order!
	author = commit.Author

	return
}

// CommitTree commit a previously written tree
func (repo *Repository) CommitTree(signing *identity.Pair, info *revision.CommitInfo) (commit_hash cas.ContentID, err error) {
	// ensure that commit info is well-formed
	if err = validate.CommitTag(info.Tag); err != nil {
		return
	}

	if info.Parent != cas.Nil {
		_, _, err = repo.check_commit(info.Parent)
		if err != nil {
			err = fmt.Errorf("faws/repo: error verifying parent commit: %w", err)
			return
		}
	}

	// marshal commit info
	new_commit_info_bytes, marshal_commit_info_err := revision.MarshalCommitInfo(info)
	if marshal_commit_info_err != nil {
		err = marshal_commit_info_err
		return
	}

	var new_commit revision.Commit
	new_commit.Author = signing.ID()
	if new_commit.Author == identity.Nobody {
		err = fmt.Errorf("faws/repo: author cannot be nobody")
		return
	}
	new_commit.Info = new_commit_info_bytes

	// sign the commit info
	identity.Sign(signing, new_commit.Info, &new_commit.Signature)

	// marshal the signed commit
	new_commit_bytes, marshal_commit_err := revision.MarshalCommit(&new_commit)
	if marshal_commit_err != nil {
		err = marshal_commit_err
		return
	}

	// store the commit
	_, commit_hash, err = repo.objects.Store(cas.Commit, new_commit_bytes)
	if err != nil {
		return
	}

	// clear cached objects
	repo.index.cache_objects = make(map[cas.ContentID]uint32)

	if err = repo.write_tag(info.Tag, commit_hash); err != nil {
		return
	}

	if err = repo.write_index(); err != nil {
		return
	}

	// ta-da!
	return
}

func (repo *Repository) GetCommit(commit_hash cas.ContentID) (author identity.ID, info *revision.CommitInfo, err error) {
	author, info, err = repo.check_commit(commit_hash)
	return
}
