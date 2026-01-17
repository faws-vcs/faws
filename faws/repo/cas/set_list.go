package cas

import (
	"encoding/hex"
	"os"
	"path/filepath"

	"github.com/faws-vcs/console"
)

// ListFunc is a callback for enumerating objects in the repository
type ListFunc func(id ContentID) (err error)

// List will enumerate all objects in the Set using the supplied [ListFunc] callback
func (set *Set) List(fn ListFunc) (err error) {
	var id ContentID
	var buckets []os.DirEntry
	var bucket2s []os.DirEntry
	var objects []os.DirEntry
	var name_err error
	buckets, err = os.ReadDir(set.directory)
	if err != nil {
		console.Println(set.directory)
		return
	}

	for _, bucket := range buckets {
		if bucket.IsDir() && len(bucket.Name()) == 2 {
			_, name_err = hex.Decode(id[0:1], []byte(bucket.Name()))
			if name_err == nil {
				bucket1path := filepath.Join(set.directory, bucket.Name())
				bucket2s, err = os.ReadDir(bucket1path)
				if err != nil {
					return
				}

				for _, bucket2 := range bucket2s {
					if bucket2.IsDir() && len(bucket2.Name()) == 2 {
						_, name_err = hex.Decode(id[1:2], []byte(bucket2.Name()))
						if name_err == nil {
							bucket2path := filepath.Join(bucket1path, bucket2.Name())
							objects, err = os.ReadDir(bucket2path)
							if err != nil {
								return
							}

							for _, object := range objects {
								if !object.IsDir() && len(object.Name()) == ((ContentIDSize-2)*2) {
									_, name_err = hex.Decode(id[2:], []byte(object.Name()))
									if name_err == nil {
										if err = fn(id); err != nil {
											return
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return
}
