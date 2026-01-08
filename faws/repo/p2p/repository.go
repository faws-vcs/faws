package p2p

import (
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/google/uuid"
)

// Interface for loading/saving repository objects
type Repository interface {
	// get the UUID of the repository
	UUID() uuid.UUID
	// Load an object from the repository
	LoadObject(id cas.ContentID) (prefix cas.Prefix, object []byte, err error)
	// Store an object in the repository
	StoreObject(prefix cas.Prefix, data []byte) (new bool, id cas.ContentID, err error)
	// Read basic info about object
	StatObject(id cas.ContentID) (size int64, err error)
	// Write tag
	WriteTag(name string, commit_hash cas.ContentID) (err error)
	// Read tag
	ReadTag(name string) (commit_hash cas.ContentID, err error)
}
