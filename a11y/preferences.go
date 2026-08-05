package a11y

import (
	"time"

	"github.com/reactivego/rx"
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

// FromSource returns an Observable that polls src every interval,
// emitting A11yPrefs only when the value changes.
// The first emission happens immediately (no initial delay).
func FromSource(src Source, interval time.Duration) rx.Observable[A11yPrefs] {
	return rx.Map(rx.Ticker(0, interval), func(_ time.Time) A11yPrefs {
		prefs, _ := src.Read()
		return prefs
	}).DistinctUntilChanged(rx.Equal[A11yPrefs]())
}

// Live returns an Observable backed by the OS accessibility APIs,
// polling every interval and emitting whenever a preference changes.
//
// Recommended interval: ≥1s. OS accessibility properties are cached by the
// platform and typically won't reflect a toggle for several hundred ms.
func Live(interval time.Duration) rx.Observable[A11yPrefs] {
	return FromSource(defaultSource(), interval)
}
