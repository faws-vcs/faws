package revision

import (
	"encoding/binary"
	"fmt"
	"unicode/utf8"

	"github.com/faws-vcs/faws/faws/identity"
	"github.com/faws-vcs/faws/faws/repo/cas"
)

var (
	ErrCommitEmpty         = fmt.Errorf("faws/repo/revision: commit is empty")
	ErrCommitInfoMalformed = fmt.Errorf("faws/repo/revision: commit info is malformed")
)

// CommitInfo contains the information, prior to cryptographic signature.
type CommitInfo struct {
	// The author's details at the time the commit was made.
	// By its inclusion here, it is cryptographically signed by author when a commit is made.
	// Therefore, it can be stored in the keyring if it is more recent than the version already in the keyring.
	AuthorAttributes identity.Attributes
	// ID of the parent commit.
	// The first commit has no parent, so cas.Nil
	Parent cas.ContentID
	// The new filesystem tree
	Tree cas.ContentID
	// Unix seconds for when the filesystem is dated to.
	// (the date of the tree at this repository)
	// Faws does not manage individual timestamps for files.
	TreeDate int64
	// Unix seconds for when the commit was made.
	CommitDate int64
	// Tag string that describes the content/feature.
	// This cannot be the same string as any ancestor/parent commit.
	Tag string
}

// Commit contains the whole commit, in serialized and cryptographically signed form.
type Commit struct {
	// The cryptographic ID of the author
	Author identity.ID
	// ID signature for the commit. This adds entropy, making the hash of this Commit more unique
	Signature identity.Signature
	// Serialized CommitInfo
	Info []byte
}

func MarshalCommitInfo(info *CommitInfo) (data []byte, err error) {
	data = make([]byte, 0, 4+(cas.ContentIDSize*2)+8+8+4+len(info.Tag))

	attributes_data, attributes_err := identity.MarshalAttributes(&info.AuthorAttributes)
	if attributes_err != nil {
		return
	}
	var attributes_size [4]byte
	binary.LittleEndian.PutUint32(attributes_size[:], uint32(len(attributes_data)))
	data = append(data, attributes_size[:]...)
	data = append(data, attributes_data...)

	// write content ids
	data = append(data, info.Parent[:]...)
	data = append(data, info.Tree[:]...)

	// write date of commit tree files
	var tree_date [8]byte
	binary.LittleEndian.PutUint64(tree_date[:], uint64(info.TreeDate))
	data = append(data, tree_date[:]...)

	// write date of commit itself
	var commit_date [8]byte
	binary.LittleEndian.PutUint64(commit_date[:], uint64(info.CommitDate))
	data = append(data, commit_date[:]...)

	// write tag
	var tag_length [4]byte
	binary.LittleEndian.PutUint32(tag_length[:], uint32(len(info.Tag)))
	data = append(data, tag_length[:]...)
	data = append(data, []byte(info.Tag)...)

	return
}

func MarshalCommit(commit *Commit) (data []byte, err error) {
	data = make([]byte, 0, 32+64+len(commit.Info))

	data = append(data, commit.Author[:]...)
	data = append(data, commit.Signature[:]...)

	data = append(data, commit.Info...)

	if len(commit.Info) == 0 {
		err = ErrCommitEmpty
		return
	}

	return
}

func UnmarshalCommitInfo(data []byte, info *CommitInfo) (err error) {
	if len(data) < (cas.ContentIDSize + cas.ContentIDSize + 8 + 8 + 4 + cas.ContentIDSize) {
		err = ErrCommitInfoMalformed
		return
	}
	field := data

	// read attributes
	attributes_size := binary.LittleEndian.Uint32(data[:4])
	field = field[4:]
	attributes_data := field[:attributes_size]
	field = field[attributes_size:]
	if err = identity.UnmarshalAttributes(attributes_data, &info.AuthorAttributes); err != nil {
		return
	}

	// read content ids
	copy(info.Parent[:], field[:cas.ContentIDSize])
	field = field[cas.ContentIDSize:]

	copy(info.Tree[:], field[:cas.ContentIDSize])
	field = field[cas.ContentIDSize:]

	// read date of commit tree files
	info.TreeDate = int64(binary.LittleEndian.Uint64(field[:8]))
	field = field[8:]

	// read date of commit itself
	info.CommitDate = int64(binary.LittleEndian.Uint64(field[:8]))
	field = field[8:]

	tag_size := uint32(binary.LittleEndian.Uint32(field[:4]))
	field = field[4:]
	if len(field) < int(tag_size) {
		err = ErrCommitInfoMalformed
		return
	}
	if !utf8.Valid(field[:tag_size]) {
		err = ErrCommitInfoMalformed
		return
	}
	info.Tag = string(field[:tag_size])

	return
}

func UnmarshalCommit(data []byte, commit *Commit) (err error) {
	if len(data) < 32+64+1 {
		err = ErrCommitEmpty
		return
	}
	field := data

	copy(commit.Author[:], field[:identity.IDSize])
	field = field[identity.IDSize:]

	copy(commit.Signature[:], field[:identity.SignatureSize])
	field = field[identity.SignatureSize:]

	commit.Info = field
	return
}
