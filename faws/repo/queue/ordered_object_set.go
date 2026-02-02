package queue

import (
	"github.com/google/btree"
)

type OrderedSetItem[T any] interface {
	Less(T) bool
}

type OrderedSet[T OrderedSetItem[T]] struct {
	items *btree.BTreeG[T]
}

func (ordered_set *OrderedSet[T]) Init() {
	ordered_set.items = btree.NewG(2, T.Less)
}

func (ordered_set *OrderedSet[T]) Len() int {
	return ordered_set.items.Len()
}

func (ordered_set *OrderedSet[T]) Push(task_object T) (added bool) {
	_, replaced := ordered_set.items.ReplaceOrInsert(task_object)
	added = !replaced
	return
}

// if empty, removed = true
func (ordered_set *OrderedSet[T]) Pop() (task_object T, removed bool) {
	task_object, removed = ordered_set.items.DeleteMin()
	return
}

func (ordered_set *OrderedSet[T]) Remove(task_object T) (removed bool) {
	_, removed = ordered_set.items.Delete(task_object)
	return
}

func (ordered_set *OrderedSet[T]) Contains(task_object T) (contains bool) {
	_, contains = ordered_set.items.Get(task_object)
	return
}

func (ordered_set *OrderedSet[T]) Clear() (cleared bool) {
	cleared = ordered_set.items.Len() > 0
	ordered_set.items.Clear(false)
	return
}

func (ordered_set *OrderedSet[T]) Destroy() (destroyed bool) {
	if ordered_set.items != nil {
		return
	}

	ordered_set.items.Clear(false)
	ordered_set.items = nil
	destroyed = true
	return
}
