package containers

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestPromise(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tempPromise := NewPromise[int](nil)

	tempPromise.Produce(1)
	res, err := tempPromise.Await(ctx)
	if res != 1 || err != nil {
		t.Fatal("unexpected Promise.Await")
	}
	res, err = tempPromise.Current()
	if res != 1 || err != nil {
		t.Fatal("unexpected Promise.Current when ready")
	}

	var cancelCalled atomic.Int64
	cancelFunc := func() { cancelCalled.Add(1) }

	tempPromise = NewPromise[int](cancelFunc)
	res, err = tempPromise.Current()
	if res != 0 || !errors.Is(err, ErrNotReady) {
		t.Fatal("unexpected Promise.Current when not ready")
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		res, err = tempPromise.Await(ctx)
		wg.Done()
	}()
	tempPromise.Produce(2)
	wg.Wait()
	if res != 2 || err != nil {
		t.Fatal("unexpected Promise.Await in parallel")
	}
	res, err = tempPromise.Current()
	if res != 2 || err != nil {
		t.Fatal("unexpected Promise.Current 2nd time")
	}

	tempPromise = NewPromise[int](cancelFunc)

	errErrorProduced := errors.New("err produced")
	wg.Add(1)
	go func() {
		res, err = tempPromise.Await(ctx)
		wg.Done()
	}()
	tempPromise.ProduceError(errErrorProduced)
	wg.Wait()
	if res != 0 || !errors.Is(err, errErrorProduced) {
		t.Fatal("unexpected Promise.Await after setError")
	}
	res, err = tempPromise.Current()
	if res != 0 || !errors.Is(err, errErrorProduced) {
		t.Fatal("unexpected Promise.Current 2nd time")
	}

	if cancelCalled.Load() != 0 {
		t.Fatal("cancel called by await/current when it shouldn't be")
	}

	tempPromise.Cancel()
	if cancelCalled.Load() != 0 {
		t.Fatal("cancel called after error produced")
	}

	tempPromise = NewPromise[int](cancelFunc)
	shortCtx, shortCancel := context.WithTimeout(ctx, time.Millisecond*100)
	defer shortCancel()
	res, err = tempPromise.Await(shortCtx)
	if res != 0 || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("unexpected Promise.Await with timeout")
	}
	if cancelCalled.Load() != 1 {
		t.Fatal("cancel not called by await on timeout")
	}
	tempPromise.Cancel()
	if cancelCalled.Load() != 2 {
		t.Fatal("cancel not called by promise.Cancel")
	}
}
