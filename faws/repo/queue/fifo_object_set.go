package queue

// type fifo_object_set[T comparable] struct {

// }

// func (fifo_object_set *fifo_object_set[T]) Init() {

// }

// func (fifo_object_set *fifo_object_set[T]) Len() int {
// 	return len(fifo_object_set)
// }

// func (fifo_object_set *fifo_object_set[T]) Push(object T) (added bool) {
// 	_, replaced := fifo_object_set.items.ReplaceOrInsert(object)
// 	added = !replaced
// 	return
// }

// // returns the
// func (fifo_object_set *fifo_object_set[T]) Pop() (object T, removed bool) {
// 	object, removed = fifo_object_set.items.DeleteMin()
// 	return
// }

// func (fifo_object_set *fifo_object_set[T]) Remove(object T) (removed bool) {
// 	_, removed = fifo_object_set.items.Delete(object)
// 	return
// }

// func (fifo_object_set *fifo_object_set[T]) Contains(object T) (contains bool) {
// 	sort.Search(fifo_object_set.Len(), func(i int) bool {
// 		return
// 	})
// 	_, contains = fifo_object_set.items.Get(object)
// 	return
// }

// func (fifo_object_set *fifo_object_set[T]) Clear() (cleared bool) {
// 	cleared = fifo_object_set.items.Len() > 0
// 	fifo_object_set.items.Clear(false)
// 	return
// }
