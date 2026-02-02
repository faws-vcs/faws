package queue

import (
	"io"
	"sync"
	"sync/atomic"
)

// TaskQueue: a linear unique-object task queue.
//
// Designed for the purpose of downloading a Faws merkle tree across many worker routines

// The queue stores unique items, that can only be processed exactly once.
// For instance, pushing "hello world" several times
// will only result in "hello world" being popped once.
// you also need to run Complete() on popped objects,
// this will signal to all p
type TaskQueue[T OrderedSetItem[T]] struct {
	// a pull task starts with 1 count, representing the root object (which is typically a commit)
	// once all child objects (in the form of pull tasks) are added, the task counter will subtract 1 and the task will be removed
	task_counter atomic.Int64
	push_cond    sync.Cond
	guard_tasks  sync.Mutex
	queue_count  int
	// contains tasks that are being processed or were already processed
	popped_tasks OrderedSet[T]
	// contains tasks that are available for workers to process
	available_tasks OrderedSet[T]
}

// Init: required to use the TaskQueue
func (task_queue *TaskQueue[T]) Init() {
	task_queue.push_cond.L = new(sync.Mutex)
	task_queue.popped_tasks.Init()
	task_queue.available_tasks.Init()
}

func (task_queue *TaskQueue[T]) Destroy() {
	task_queue.push_cond.L = nil
	task_queue.popped_tasks.Destroy()
	task_queue.available_tasks.Destroy()
}

// Len return the the amount of all tasks pushed and popped from the queue
func (task_queue *TaskQueue[T]) Len() (n int) {
	task_queue.guard_tasks.Lock()
	n = task_queue.queue_count
	task_queue.guard_tasks.Unlock()
	return
}

// Push if this object wasn't already pushed
func (task_queue *TaskQueue[T]) Push(object T) {
	task_queue.guard_tasks.Lock()
	if !task_queue.popped_tasks.Contains(object) && task_queue.available_tasks.Push(object) {
		task_queue.queue_count++
		task_queue.task_counter.Add(1)
		task_queue.push_cond.Signal()
	}
	task_queue.guard_tasks.Unlock()
}

// complete a task
func (task_queue *TaskQueue[T]) Complete(object T) {
	task_queue.guard_tasks.Lock()
	new_counter := task_queue.task_counter.Add(-1)
	if new_counter == 0 {
		// tells all workers to stop
		task_queue.push_cond.Broadcast()
	}
	task_queue.guard_tasks.Unlock()
}

func (task_queue *TaskQueue[T]) Contains(object T) (contains bool) {
	task_queue.guard_tasks.Lock()
	defer task_queue.guard_tasks.Unlock()

	contains = task_queue.popped_tasks.Contains(object)
	if contains {
		return
	}

	contains = task_queue.available_tasks.Contains(object)
	return
}

func (task_queue *TaskQueue[T]) Stop() {
	task_queue.guard_tasks.Lock()
	task_queue.task_counter.Store(0)
	task_queue.available_tasks.Clear()
	task_queue.popped_tasks.Clear()
	task_queue.push_cond.Broadcast()
	task_queue.guard_tasks.Unlock()
}

// remove and return a task from the set of available tasks
// if the available task set is empty, this will block if future tasks are expected
// once empty and no future tasks are expected, will return with io.EOF
func (task_queue *TaskQueue[T]) Pop() (object T, err error) {
	for {
		task_queue.guard_tasks.Lock()

		if task_queue.available_tasks.Len() == 0 {
			task_queue.guard_tasks.Unlock()
			if task_queue.task_counter.Load() == 0 {
				err = io.EOF
				return
			}

			// wait for more objects to be pushed to the queue
			// TODO: ensure that broadcast is sent when the task counter reaches zero
			task_queue.push_cond.L.Lock()
			task_queue.push_cond.Wait()
			task_queue.push_cond.L.Unlock()

			continue
		}

		var exists bool
		object, exists = task_queue.available_tasks.Pop()
		if !exists {
			panic("cannot remove task from set, though it is non-empty")
		}

		if !task_queue.popped_tasks.Push(object) {
			panic("task was already popped")
		}

		task_queue.guard_tasks.Unlock()

		return
	}
}
