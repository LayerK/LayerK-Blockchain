// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package containers

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
	newCap := 8
	if len(q.slice) > 0 {
		newCap = len(q.slice) * 2
	}
	newSlice := make([]T, newCap)
	for i := 0; i < q.count; i++ {
		newSlice[i] = q.slice[(q.head+i)%len(q.slice)]
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
	return item
}

func (q *Queue[T]) Len() int {
	return q.count
}
