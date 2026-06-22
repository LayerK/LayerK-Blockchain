// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package containers

import (
	"fmt"
	"testing"

	"github.com/offchainlabs/nitro/util/testhelpers"
)

func TestQueue(t *testing.T) {
	q := Queue[int]{}
	if _, ok := q.Peek(); ok {
		testhelpers.FailImpl(t, "empty queue reported a peek value")
	}

	for i := 0; i < 16; i++ {
		q.Push(i)
	}
	if q.Len() != 16 {
		testhelpers.FailImpl(t, fmt.Sprintf("Unexpected queue length: want %d, got %d", 16, q.Len()))
	}
	if got, ok := q.Peek(); !ok || got != 0 {
		testhelpers.FailImpl(t, fmt.Sprintf("Unexpected peek: want %d, got %d, ok=%t", 0, got, ok))
	}

	for i := 0; i < 12; i++ {
		requirePop(t, &q, i)
	}
	for i := 16; i < 40; i++ {
		q.Push(i)
	}
	for i := 12; i < 40; i++ {
		requirePop(t, &q, i)
	}

	if q.Len() != 0 {
		testhelpers.FailImpl(t, fmt.Sprintf("Unexpected queue length after drain: want %d, got %d", 0, q.Len()))
	}
	if len(q.slice) != 0 {
		testhelpers.FailImpl(t, fmt.Sprintf("Non-empty queue buffer after drain: len=%d", len(q.slice)))
	}
	if got := q.Pop(); got != 0 {
		testhelpers.FailImpl(t, fmt.Sprintf("Unexpected element popped: want %d, got %d", 0, got))
	}
}

func TestQueueShrinksRetainingWrappedOrder(t *testing.T) {
	q := Queue[int]{}
	initNumElements := 10000
	for i := 0; i < initNumElements; i++ {
		q.Push(i)
	}

	bigCap := len(q.slice)
	if bigCap < initNumElements {
		testhelpers.FailImpl(t, fmt.Sprintf("Unexpected capacity %d<%d", bigCap, initNumElements))
	}

	for i := 0; i < 9000; i++ {
		requirePop(t, &q, i)
	}
	if len(q.slice) >= bigCap {
		testhelpers.FailImpl(t, fmt.Sprintf("Queue did not shrink: before=%d, after=%d", bigCap, len(q.slice)))
	}

	for i := initNumElements; i < initNumElements+100; i++ {
		q.Push(i)
	}
	for i := 9000; i < initNumElements+100; i++ {
		requirePop(t, &q, i)
	}
	if len(q.slice) != 0 {
		testhelpers.FailImpl(t, fmt.Sprintf("Non-empty queue buffer after drain: len=%d", len(q.slice)))
	}
}

func requirePop(t *testing.T, q *Queue[int], want int) {
	t.Helper()
	got := q.Pop()
	if got != want {
		testhelpers.FailImpl(t, fmt.Sprintf("Unexpected element popped: want %d, got %d", want, got))
	}
}
