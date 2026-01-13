package queue

import (
	"sync"
)

// TaskHeap offers a continuous heap of tasks.
type TaskHeap[T comparable] struct {
	guard_items     sync.Mutex
	heap_count      int64
	available_items UnorderedSet[T]
	completed_items UnorderedSet[T]
}

func (task_heap *TaskHeap[T]) Init() {
	task_heap.available_items.Init()
	task_heap.completed_items.Init()
}

// Pick: returns a random item from the heap.
//
// There is no guarantee that the same item won't removed twice or at the same time in another goroutine
func (task_heap *TaskHeap[T]) Pick() (item T, err error) {
	task_heap.guard_items.Lock()
	item, err = task_heap.available_items.Get()
	task_heap.guard_items.Unlock()
	return
}

// Complete removes the item from the list of available items,
// and adds it to the list of completed items.
// This ensures that it won't get picked ever again,
// even if the same item gets pushed.
func (task_heap *TaskHeap[T]) Complete(item T) (completed bool) {
	task_heap.guard_items.Lock()
	completed = task_heap.available_items.Remove(item)
	task_heap.completed_items.Push(item)
	task_heap.guard_items.Unlock()
	return
}

func (task_heap *TaskHeap[T]) Push(item T) (pushed bool) {
	task_heap.guard_items.Lock()
	if !task_heap.completed_items.Contains(item) {
		pushed = task_heap.available_items.Push(item)
	}
	if pushed {
		task_heap.heap_count++
	}
	task_heap.guard_items.Unlock()
	return
}

// Contains returns true if the item is in the heap, either as an available or completed item
func (task_heap *TaskHeap[T]) Contains(item T) (contains bool) {
	task_heap.guard_items.Lock()
	defer task_heap.guard_items.Unlock()
	contains = task_heap.available_items.Contains(item)
	if contains {
		return
	}

	contains = task_heap.completed_items.Contains(item)
	return
}

func (task_heap *TaskHeap[T]) IsCompleted(item T) (completed bool) {
	task_heap.guard_items.Lock()
	completed = task_heap.completed_items.Contains(item)
	task_heap.guard_items.Unlock()
	return
}

func (task_heap *TaskHeap[T]) IsAvailable(item T) (available bool) {
	task_heap.guard_items.Lock()
	available = task_heap.completed_items.Contains(item)
	task_heap.guard_items.Unlock()
	return
}

func (task_heap *TaskHeap[T]) Len() (n int) {
	task_heap.guard_items.Lock()
	n = int(task_heap.heap_count)
	task_heap.guard_items.Unlock()
	return
}
