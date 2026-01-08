package remote

import (
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/google/uuid"
)

// Origin: the source from which we can obtain a repository
//
// possibilities:
//   - A local filesystem directory
//   - A remote HTTP autoindex filesystem such as Apache or Nginx. Autoindex is required to obtain a list of tags.
type Origin interface {
	// - file:///
	// - http://, https://
	URI() (uri string)

	UUID() (id uuid.UUID, err error)

	// Read the list of tags from the remote repository
	Tags() (tags []string, err error)

	// Read a tag
	ReadTag(name string) (commit_hash cas.ContentID, err error)

	// Get an object from the remote
	GetObject(object_hash cas.ContentID) (prefix cas.Prefix, data []byte, err error)

	// attempt to deabbreviate
	Deabbreviate(ref string) (object_hash cas.ContentID, err error)
}
