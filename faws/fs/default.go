package fs

import "os"

const (
	DefaultPublicDirPerm  os.FileMode = 0775
	DefaultPrivateDirPerm os.FileMode = 0700
	DefaultPublicPerm     os.FileMode = 0664
	DefaultPrivatePerm    os.FileMode = 0600
)
