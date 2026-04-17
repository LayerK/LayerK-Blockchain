// Copyright 2021-2022, Offchain Labs, Inc.
// For license information, see https://github.com/OffchainLabs/nitro/blob/master/LICENSE.md

package stopwaiter

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/log"

	"github.com/offchainlabs/nitro/util/testhelpers"
)

const testStopDelayWarningTimeout = 350 * time.Millisecond

type TestStruct struct{}

func TestStopWaiterStopAndWaitTimeoutShouldWarn(t *testing.T) {
	logHandler := testhelpers.InitTestLog(t, log.LvlTrace)
	sw := StopWaiter{}
	testCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sw.Start(context.Background(), &TestStruct{})
	sw.LaunchThread(func(ctx context.Context) {
		<-testCtx.Done()
	})
	go func() {
		err := sw.stopAndWaitImpl(testStopDelayWarningTimeout)
		testhelpers.RequireImpl(t, err)
	}()
	time.Sleep(testStopDelayWarningTimeout + 100*time.Millisecond)
	if !logHandler.WasLogged("taking too long to stop") {
		testhelpers.FailImpl(t, "Failed to log about waiting long on StopAndWait")
	}
}

func TestStopWaiterStopAndWaitTimeoutShouldNotWarn(t *testing.T) {
	logHandler := testhelpers.InitTestLog(t, log.LvlTrace)
	sw := StopWaiter{}
	sw.Start(context.Background(), &TestStruct{})
	sw.LaunchThread(func(ctx context.Context) {
		<-ctx.Done()
	})
	sw.StopAndWait()
	if logHandler.WasLogged("taking too long to stop") {
		testhelpers.FailImpl(t, "Incorrectly logged about waiting long on StopAndWait")
	}
}

func TestStopWaiterStopAndWaitBeforeStart(t *testing.T) {
	sw := StopWaiter{}
	sw.StopAndWait()
}

func TestStopWaiterStopAndWaitAfterStop(t *testing.T) {
	sw := StopWaiter{}
	sw.Start(context.Background(), &TestStruct{})
	ctx := sw.GetContext()
	sw.StopOnly()
	<-ctx.Done()
	sw.StopAndWait()
}

func TestStopWaiterStopAndWaitMultipleTimes(t *testing.T) {
	sw := StopWaiter{}
	sw.StopAndWait()
	sw.StopAndWait()
	sw.StopAndWait()
	sw.Start(context.Background(), &TestStruct{})
	sw.StopAndWait()
	sw.StopAndWait()
	sw.StopAndWait()
}

func TestStopWaiterStopOnlyThenStopAndWait(t *testing.T) {
	t.Parallel()
	sw := StopWaiter{}
	sw.Start(context.Background(), &TestStruct{})
	var threadStopping atomic.Bool
	sw.LaunchThread(func(context.Context) {
		time.Sleep(time.Second)
		threadStopping.Store(true)
	})
	sw.StopOnly()
	sw.StopAndWait()
	if !threadStopping.Load() {
		t.Error("StopAndWait returned before background thread stopped")
	}
}

func TestCallWhenTriggeredWithStopsWhenChannelCloses(t *testing.T) {
	t.Parallel()

	sw := StopWaiter{}
	sw.Start(context.Background(), &TestStruct{})

	triggerChan := make(chan int, 1)
	var calls atomic.Int64
	testhelpers.RequireImpl(t, CallWhenTriggeredWith(&sw.StopWaiterSafe, func(_ context.Context, val int) {
		calls.Add(int64(val))
	}, triggerChan))

	triggerChan <- 2
	close(triggerChan)
	sw.StopAndWait()

	if calls.Load() != 2 {
		t.Fatalf("unexpected values processed after channel close: %d", calls.Load())
	}
}

func TestChanRateLimiterClosesWhenInputChannelCloses(t *testing.T) {
	t.Parallel()

	sw := StopWaiter{}
	sw.Start(context.Background(), &TestStruct{})

	input := make(chan int, 1)
	output, err := ChanRateLimiter(&sw.StopWaiterSafe, input, func() time.Duration { return 0 })
	testhelpers.RequireImpl(t, err)

	input <- 7
	if got, ok := <-output; !ok || got != 7 {
		t.Fatalf("unexpected first rate-limited value: got=%d ok=%v", got, ok)
	}

	close(input)
	if _, ok := <-output; ok {
		t.Fatal("expected output channel to close after input channel closed")
	}

	sw.StopAndWait()
}
