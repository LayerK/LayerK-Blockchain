package timeboost

import (
	"context"
	"time"
)

type roundTicker struct {
	c               chan time.Time
	roundTimingInfo RoundTimingInfo
}

func newRoundTicker(roundTimingInfo RoundTimingInfo) *roundTicker {
	return &roundTicker{
		c:               make(chan time.Time, 1),
		roundTimingInfo: roundTimingInfo,
	}
}

func (t *roundTicker) tickAtAuctionClose(ctx context.Context) {
	t.start(ctx, t.roundTimingInfo.AuctionClosing)
}

func (t *roundTicker) tickAtReserveSubmissionDeadline(ctx context.Context) {
	t.start(ctx, t.roundTimingInfo.AuctionClosing+t.roundTimingInfo.ReserveSubmission)
}

func (t *roundTicker) start(ctx context.Context, timeBeforeRoundStart time.Duration) {
	for {
		nextTick := t.roundTimingInfo.TimeTilNextRound() - timeBeforeRoundStart
		if nextTick < 0 {
			nextTick += t.roundTimingInfo.Round
		}

		select {
		case <-time.After(nextTick):
			// The channel is buffered (cap 1) so a slow consumer must not wedge the
			// ticker goroutine past ctx cancellation.
			select {
			case t.c <- time.Now():
			case <-ctx.Done():
				close(t.c)
				return
			}
		case <-ctx.Done():
			close(t.c)
			return
		}
	}
}
