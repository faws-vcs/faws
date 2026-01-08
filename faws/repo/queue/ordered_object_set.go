package queue

import "github.com/google/btree"

type OrderedSetItem[T any] interface {
	Less(T) bool
}

type ordered_object_set[T OrderedSetItem[T]] struct {
	items *btree.BTreeG[T]
}

func (ordered_object_set *ordered_object_set[T]) Init() {
	ordered_object_set.items = btree.NewG(2, T.Less)
}

func (ordered_object_set *ordered_object_set[T]) Len() int {
	return ordered_object_set.items.Len()
}

func (ordered_object_set *ordered_object_set[T]) Push(task_object T) (added bool) {
	_, replaced := ordered_object_set.items.ReplaceOrInsert(task_object)
	added = !replaced
	return
}

// if empty, removed = true
func (ordered_object_set *ordered_object_set[T]) Pop() (task_object T, removed bool) {
	task_object, removed = ordered_object_set.items.DeleteMin()
	return
}

func (ordered_object_set *ordered_object_set[T]) Remove(task_object T) (removed bool) {
	_, removed = ordered_object_set.items.Delete(task_object)
	return
}

func (ordered_object_set *ordered_object_set[T]) Contains(task_object T) (contains bool) {
	_, contains = ordered_object_set.items.Get(task_object)
	return
}

func (ordered_object_set *ordered_object_set[T]) Clear() (cleared bool) {
	cleared = ordered_object_set.items.Len() > 0
	ordered_object_set.items.Clear(false)
	return
}
