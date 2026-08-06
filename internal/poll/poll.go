// Package poll implements the one shared poll loop behind spectrum's live
// OS-preference streams (spectrum/system, spectrum/a11y).
//
// [Shared] is the FX.5 fix for the cold-stream defect the register recorded:
// a plain Ticker+Map observable starts a fresh ticker per subscription, so a
// stream handed to n consumers polled the OS n times per interval — and on
// macOS every poll is a `defaults` fork+exec. Shared multicasts the loop:
// however many subscribers attach, the read function runs on exactly one
// ticker, and every subscriber observes the same emissions.
package poll

import (
	"time"

	"github.com/reactivego/rx"
)

// slot wraps a polled value so the multicast seed — the state before the
// first read completes — is distinguishable from every real value,
// including T's zero value. Only ok slots escape to subscribers.
type slot[T any] struct {
	ok bool
	v  T
}

// Shared returns a hot, multicast observable that calls read every interval
// on one shared loop and emits the result whenever it changes.
//
//   - One loop, n subscribers: the ticker starts when the first subscriber
//     attaches and every later subscriber shares it — read runs once per
//     interval regardless of subscriber count.
//   - Replay-latest: a subscriber arriving after the first read immediately
//     observes the most recent value (rx Behavior semantics), then tracks
//     changes. A subscriber that lags converges on the latest value rather
//     than draining intermediates.
//   - The loop stops when the subscriber count drops to zero (rx RefCount)
//     and restarts on the next subscription; across such a restart the last
//     value is replayed first, then refreshed by the restarted loop's
//     immediate first read.
//   - Emits only on change: each subscriber sees consecutive duplicates
//     collapsed, so replay plus the next identical read is one emission.
//
// The first read is scheduled immediately (no initial delay). The seed the
// multicast holds before that read completes never escapes: subscribers see
// nothing until the first read lands, exactly as the cold shape behaved.
func Shared[T comparable](read func() T, interval time.Duration) rx.Observable[T] {
	ticks := rx.Map(rx.Ticker(0, interval), func(time.Time) slot[T] {
		return slot[T]{ok: true, v: read()}
	})
	shared := ticks.Behavior(slot[T]{}).RefCount()
	return rx.Map(
		shared.Filter(func(s slot[T]) bool { return s.ok }),
		func(s slot[T]) T { return s.v },
	).DistinctUntilChanged(rx.Equal[T]())
}
