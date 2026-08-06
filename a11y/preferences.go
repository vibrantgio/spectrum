package a11y

import (
	"time"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/internal/poll"
)

// A11yPrefs carries the current accessibility display preferences reported by
// the operating system. All fields are comparable types — DistinctUntilChanged
// relies on struct equality; do not add slice or map fields.
type A11yPrefs struct {
	ReduceMotion     bool // OS "Reduce Motion" preference
	HighContrast     bool // OS "Increase Contrast" / "High Contrast" preference
	IncreaseTextSize bool // OS "Larger Text" preference (platform support varies)
}

// Source reads the current OS accessibility preferences.
// Implement this interface to provide a custom or test-double backend.
type Source interface {
	Read() (A11yPrefs, error)
}

// FromSource returns a shared Observable that polls src every interval,
// emitting A11yPrefs only when the value changes. The first read is
// scheduled immediately (no initial delay).
//
// The returned observable is multicast (FX.5): all subscribers to this one
// value share a single poll loop, a subscriber arriving after the first
// read immediately observes the latest A11yPrefs before tracking changes,
// and the loop stops when the last subscriber unsubscribes (restarting on
// the next subscription). Each FromSource call builds its own loop —
// sharing is per returned value, not per Source.
func FromSource(src Source, interval time.Duration) rx.Observable[A11yPrefs] {
	return poll.Shared(func() A11yPrefs {
		prefs, _ := src.Read()
		return prefs
	}, interval)
}

// Live returns an Observable backed by the OS accessibility APIs,
// polling every interval and emitting whenever a preference changes.
// Like [FromSource] it is shared: n subscribers to one Live value cost
// one poll loop, not n.
//
// Recommended interval: ≥1s. OS accessibility properties are cached by the
// platform and typically won't reflect a toggle for several hundred ms.
func Live(interval time.Duration) rx.Observable[A11yPrefs] {
	return FromSource(defaultSource(), interval)
}
