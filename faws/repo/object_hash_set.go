package repo

import (
	"github.com/faws-vcs/faws/faws/repo/cas"
	"github.com/google/btree"
)

type object_hash_set struct {
	items *btree.BTreeG[cas.ContentID]
}

func (set *object_hash_set) Init() {
	set.items = btree.NewG(2, cas.ContentID.Less)
}

func (set *object_hash_set) Len() int {
	return set.items.Len()
}

func (set *object_hash_set) Push(object_hash cas.ContentID) (added bool) {
	_, replaced := set.items.ReplaceOrInsert(object_hash)
	added = !replaced
	return
}

// if empty, removed = true
func (set *object_hash_set) Pop() (object_hash cas.ContentID, removed bool) {
	object_hash, removed = set.items.DeleteMin()
	return
}

func (set *object_hash_set) Remove(object_hash cas.ContentID) (removed bool) {
	_, removed = set.items.Delete(object_hash)
	return
}

func (set *object_hash_set) Contains(object_hash cas.ContentID) (contains bool) {
	_, contains = set.items.Get(object_hash)
	return
}

func (set *object_hash_set) Clear() (cleared bool) {
	cleared = set.items.Len() > 0
	set.items.Clear(false)
	return
}
