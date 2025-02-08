package repo

import "fmt"

var (
	ErrRevisionNotExist = fmt.Errorf("faws/repo: revision does not exist")

	ErrLocked                  = fmt.Errorf("faws/repo: repository locked")
	ErrMalformedHead           = fmt.Errorf("faws/repo: malformed head")
	ErrInitializeCannotExist   = fmt.Errorf("faws/repo: cannot initialize non-empty repository")
	ErrBadCommit               = fmt.Errorf("faws/repo: bad commit")
	ErrBadObject               = fmt.Errorf("faws/repo: bad object")
	ErrBadRef                  = fmt.Errorf("faws/repo: bad ref")
	ErrCommitInvalidPrefix     = fmt.Errorf("faws/repo: the commit object does not have the appropriate prefix")
	ErrCommitAuthorNotTrusted  = fmt.Errorf("faws/repo: commit author isn't trusted")
	ErrBadFilename             = fmt.Errorf("faws/repo: filename isn't usable by repository hierarchy")
	ErrTreeFileNotFound        = fmt.Errorf("faws/repo: the file could not be found in tree")
	ErrTreeInvalidPrefix       = fmt.Errorf("faws/repo: the tree object does not have the appropriate prefix")
	ErrCommitDuplicateTag      = fmt.Errorf("faws/repo: commit tag already exists in history")
	ErrRepoNotExist            = fmt.Errorf("faws/repo: you are not currently in a Faws repository")
	ErrCacheEntryCannotBeEmpty = fmt.Errorf("faws/repo: a file can't be added to the cache without a name")
	ErrCacheEntryNotFound      = fmt.Errorf("faws/repo: a cache entry by that name was not found")
)
