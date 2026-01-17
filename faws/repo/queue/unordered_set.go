package queue

import (
	"io"
)

type UnorderedSet[T comparable] struct {
	items map[T]struct{}
}

func (unordered_set *UnorderedSet[T]) Init() {
	unordered_set.items = make(map[T]struct{})
}

func (unordered_set *UnorderedSet[T]) Contains(item T) (contains bool) {
	_, contains = unordered_set.items[item]
	return
}

func (unordered_set *UnorderedSet[T]) Push(item T) (pushed bool) {
	if !unordered_set.Contains(item) {
		unordered_set.items[item] = struct{}{}
		pushed = true
	}
	return
}

func (unordered_set *UnorderedSet[T]) Remove(item T) (removed bool) {
	removed = unordered_set.Contains(item)
	delete(unordered_set.items, item)
	return
}

func (unordered_set *UnorderedSet[T]) Get() (item T, err error) {
	err = io.EOF
	for key := range unordered_set.items {
		err = nil
		item = key
		break
	}
	return
}

func (unordered_set *UnorderedSet[T]) Len() (n int) {
	n = len(unordered_set.items)
	return
}
