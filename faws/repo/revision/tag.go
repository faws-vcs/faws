package revision

import "github.com/faws-vcs/faws/faws/repo/cas"

// Tags are used to identify commits with memorable, short identifiers
type Tag struct {
	// Short string that passes validate.Tag
	// It is a ref, because you can use it in place of a hash
	Name string
	// Content ID pointing to commit
	CommitHash cas.ContentID
}
