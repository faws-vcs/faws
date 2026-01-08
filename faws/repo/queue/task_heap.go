package queue

import (
	"io"
	"sync"
)

// TaskHeap offers a continuous heap of tasks.
type TaskHeap[T comparable] struct {
	guard_items     sync.Mutex
	available_items map[T]struct{}
	completed_items map[T]struct{}
}

func (task_heap *TaskHeap[T]) Init() {
	task_heap.available_items = make(map[T]struct{})
	task_heap.completed_items = make(map[T]struct{})
}

// Pick: returns a random item from the heap.
//
// There is no guarantee that the same item won't removed twice or at the same time in another goroutine
func (task_heap *TaskHeap[T]) Pick() (item T, err error) {
	task_heap.guard_items.Lock()
	err = io.EOF
	for key := range task_heap.available_items {
		err = nil
		item = key
		break
	}
	task_heap.guard_items.Unlock()
	return
}

// Complete removes the item from the list of available items,
// and adds it to the list of completed items.
// This ensures that it won't get picked ever again,
// even if the same item gets pushed.
func (task_heap *TaskHeap[T]) Complete(item T) (completed bool) {
	task_heap.guard_items.Lock()
	_, completed = task_heap.available_items[item]
	delete(task_heap.available_items, item)
	task_heap.completed_items[item] = struct{}{}
	task_heap.guard_items.Unlock()
	return
}

func (task_heap *TaskHeap[T]) Push(item T) (pushed bool) {
	task_heap.guard_items.Lock()
	if _, is_completed := task_heap.completed_items[item]; !is_completed {
		_, pushed = task_heap.available_items[item]
		task_heap.available_items[item] = struct{}{}
	}
	task_heap.guard_items.Unlock()
	return
}

// Contains returns true if the item is in the heap, either as an available or completed item
func (task_heap *TaskHeap[T]) Contains(item T) (contains bool) {
	task_heap.guard_items.Lock()
	defer task_heap.guard_items.Unlock()
	_, contains = task_heap.available_items[item]
	if contains {
		return
	}

	_, contains = task_heap.completed_items[item]
	return
}

func (task_heap *TaskHeap[T]) IsCompleted(item T) (completed bool) {
	task_heap.guard_items.Lock()
	_, completed = task_heap.completed_items[item]
	task_heap.guard_items.Unlock()
	return
}

func (task_heap *TaskHeap[T]) IsAvailable(item T) (available bool) {
	task_heap.guard_items.Lock()
	_, available = task_heap.available_items[item]
	task_heap.guard_items.Unlock()
	return
}

func (task_heap *TaskHeap[T]) Len() (n int) {
	task_heap.guard_items.Lock()
	n = len(task_heap.available_items) + len(task_heap.completed_items)
	task_heap.guard_items.Unlock()
	return
}
