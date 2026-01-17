package queue

import (
	"io"
	"sync"
	"sync/atomic"
	"time"
)

// TaskHeap offers a continuous heap of tasks.
type TaskHeap[T comparable] struct {
	ttl         time.Duration
	guard_items sync.RWMutex
	heap_count  atomic.Int64
	// available_items UnorderedSet[T]
	available_items map[T]heap_task
	completed_items UnorderedSet[T]
}

type heap_task struct {
	new       bool
	timestamp time.Time
}

func (task_heap *TaskHeap[T]) Init() {
	task_heap.ttl = 200 * time.Millisecond
	task_heap.available_items = make(map[T]heap_task)
	task_heap.completed_items.Init()
}

func (task_heap *TaskHeap[T]) read_lock() {
	task_heap.guard_items.RLock()
}

func (task_heap *TaskHeap[T]) read_unlock() {
	task_heap.guard_items.RUnlock()
}

func (task_heap *TaskHeap[T]) write_lock() {
	task_heap.guard_items.Lock()
}

func (task_heap *TaskHeap[T]) write_unlock() {
	task_heap.guard_items.Unlock()
}

// Pick: returns a random item from the heap.
//
// There is no guarantee that the same item won't removed twice or at the same time in another goroutine
func (task_heap *TaskHeap[T]) Pick() (item T, err error) {
	for {
		start := time.Now()
		task_heap.read_lock()
		err = io.EOF
		var (
			count int
			task  heap_task
		)
		for item, task = range task_heap.available_items {
			count++
			if task.new {
				err = nil
				break
			}
			if start.Sub(task.timestamp) > task_heap.ttl {
				err = nil
				break
			}
		}
		task_heap.read_unlock()

		if count == 0 {
			break
		}

		if err == nil {
			task_heap.write_lock()
			task_heap.available_items[item] = heap_task{new: false, timestamp: time.Now()}
			task_heap.write_unlock()
			break
		}

		time.Sleep(10 * time.Millisecond)
	}
	return
}

// Complete removes the item from the list of available items,
// and adds it to the list of completed items.
// This ensures that it won't get picked ever again,
// even if the same item gets pushed.
func (task_heap *TaskHeap[T]) Complete(item T) (completed bool) {
	task_heap.write_lock()
	_, completed = task_heap.available_items[item]
	if completed {
		delete(task_heap.available_items, item)
	}
	task_heap.completed_items.Push(item)
	task_heap.write_unlock()
	return
}

func (task_heap *TaskHeap[T]) Push(item T) (pushed bool) {
	task_heap.write_lock()
	if !task_heap.completed_items.Contains(item) {
		_, was_available := task_heap.available_items[item]
		if !was_available {
			task_heap.available_items[item] = heap_task{
				new:       true,
				timestamp: time.Now(),
			}
			pushed = true
		}
	}
	if pushed {
		task_heap.heap_count.Add(1)
	}
	task_heap.write_unlock()
	return
}

// Contains returns true if the item is in the heap, either as an available or completed item
func (task_heap *TaskHeap[T]) Contains(item T) (contains bool) {
	task_heap.read_lock()
	defer task_heap.read_unlock()
	// contains = task_heap.available_items.Contains(item)
	_, contains = task_heap.available_items[item]
	if contains {
		return
	}

	contains = task_heap.completed_items.Contains(item)
	return
}

func (task_heap *TaskHeap[T]) IsCompleted(item T) (completed bool) {
	task_heap.read_lock()
	defer task_heap.read_unlock()
	completed = task_heap.completed_items.Contains(item)
	return
}

func (task_heap *TaskHeap[T]) IsAvailable(item T) (available bool) {
	task_heap.read_lock()
	defer task_heap.read_unlock()
	available = task_heap.completed_items.Contains(item)
	return
}

func (task_heap *TaskHeap[T]) Len() (n int) {
	task_heap.read_lock()
	defer task_heap.read_unlock()
	n = int(task_heap.heap_count.Load())
	return
}

func (task_heap *TaskHeap[T]) AvailableLen() (n int) {
	task_heap.read_lock()
	defer task_heap.read_unlock()
	n = len(task_heap.available_items)
	return
}
