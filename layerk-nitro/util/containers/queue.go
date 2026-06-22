// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package containers

const (
	minQueueCapacity  = 8
	queueShrinkFactor = 4
)

// Queue is a FIFO queue backed by a ring-buffer slice for O(1) push and pop.
type Queue[T any] struct {
	slice []T
	head  int
	count int
}

func (q *Queue[T]) Push(item T) {
	if q.count == len(q.slice) {
		q.grow()
	}
	tail := (q.head + q.count) % len(q.slice)
	q.slice[tail] = item
	q.count++
}

func (q *Queue[T]) grow() {
	newCap := minQueueCapacity
	if len(q.slice) > 0 {
		newCap = len(q.slice) * 2
	}
	q.resize(newCap)
}

func (q *Queue[T]) shrink() {
	if q.count == 0 {
		q.slice = nil
		q.head = 0
		return
	}
	if len(q.slice) <= minQueueCapacity || q.count > len(q.slice)/queueShrinkFactor {
		return
	}
	newCap := len(q.slice) / 2
	if newCap < minQueueCapacity {
		newCap = minQueueCapacity
	}
	q.resize(newCap)
}

func (q *Queue[T]) resize(newCap int) {
	newSlice := make([]T, newCap)
	if q.count > 0 {
		if q.head+q.count <= len(q.slice) {
			copy(newSlice, q.slice[q.head:q.head+q.count])
		} else {
			first := copy(newSlice, q.slice[q.head:])
			copy(newSlice[first:], q.slice[:q.count-first])
		}
	}
	q.slice = newSlice
	q.head = 0
}

func (q *Queue[T]) Pop() T {
	var empty T
	if q.count == 0 {
		return empty
	}
	item := q.slice[q.head]
	q.slice[q.head] = empty // release reference for GC
	q.head = (q.head + 1) % len(q.slice)
	q.count--
	q.shrink()
	return item
}

func (q *Queue[T]) Peek() (T, bool) {
	var empty T
	if q.count == 0 {
		return empty, false
	}
	return q.slice[q.head], true
}

func (q *Queue[T]) Len() int {
	return q.count
}
