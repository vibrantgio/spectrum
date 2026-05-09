// Package system bridges OS-level appearance signals (dark mode, accent
// colour) into the reactive theme runtime. Per-OS shims read the live state
// behind a [Source] interface; [FromSource] turns a Source plus a poll
// interval into an rx.Observable that emits only on change. [Live] wires
// the OS-default source for the current platform; [LiveTheme] then
// converts the appearance stream into [theme.Theme] emissions whose Color
// observable matches the OS dark-mode setting.
//
// The package never imports Gio. It speaks to the OS directly so it can
// be reused from any spectrum consumer, with or without a window.
package system

import (
	"time"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/prism/theme"
	"github.com/vibrantgio/prism/tokens"
)

// Appearance is the OS-level appearance state we observe. All fields are
// comparable so the value can be used with rx.DistinctUntilChanged.
type Appearance struct {
	// Dark is true iff the OS reports a dark interface style.
	Dark bool

	// AccentIndex carries a platform-defined accent identifier. On macOS
	// it is the AppleAccentColor key (-1..7, where -1 means "multicolor"
	// and 4 is the default Blue); on platforms that expose no equivalent
	// it is zero. The mapping from index to a tokens.* colour belongs to
	// later milestones.
	AccentIndex int
}

// Source reads the current OS appearance state.
// Implement this interface to provide a custom or test-double backend.
type Source interface {
	Read() (Appearance, error)
}

// FromSource returns an Observable that polls src every interval, emitting
// Appearance only when the value changes. The first emission happens
// immediately (no initial delay).
//
// Read errors are folded into the zero-value Appearance — the stream is
// never an error stream. This keeps the contract simple for consumers
// that only care about the last good value, and matches a11y.FromSource.
func FromSource(src Source, interval time.Duration) rx.Observable[Appearance] {
	return rx.Map(rx.Ticker(0, interval), func(_ time.Time) Appearance {
		a, _ := src.Read()
		return a
	}).DistinctUntilChanged(rx.Equal[Appearance]())
}

// Live returns an Observable backed by the current OS's appearance APIs,
// polling every interval and emitting whenever a value changes.
//
// Recommended interval: 100–250 ms. The G2.2 acceptance budget allows up
// to one second between an external `defaults write` and the corresponding
// emission, but most desktop UIs prefer to feel snappier than that.
func Live(interval time.Duration) rx.Observable[Appearance] {
	return FromSource(defaultSource(), interval)
}

// LiveTheme bridges system-appearance changes to a theme.Theme stream.
// Each emission is a fresh theme.Theme whose Color field matches the OS
// dark-mode setting (tokens.DefaultLight or tokens.DefaultDark); the
// remaining token categories use their package defaults.
//
// Accent-driven palette swaps are intentionally out of scope here — the
// AccentIndex is observed and propagated, but mapping it to concrete
// ColorTokens belongs to a later spectrum milestone.
func LiveTheme(interval time.Duration) rx.Observable[theme.Theme] {
	return rx.Map(Live(interval), themeFromAppearance)
}

// FromSourceTheme is the test-friendly variant of LiveTheme: it lets a
// caller plug in a fake Source while exercising the same Appearance →
// theme.Theme bridge.
func FromSourceTheme(src Source, interval time.Duration) rx.Observable[theme.Theme] {
	return rx.Map(FromSource(src, interval), themeFromAppearance)
}

func themeFromAppearance(a Appearance) theme.Theme {
	colors := tokens.DefaultLight
	if a.Dark {
		colors = tokens.DefaultDark
	}
	return theme.Theme{
		Color:     rx.Of(colors),
		Type:      rx.Of(tokens.DefaultTypeScale),
		Motion:    rx.Of(tokens.Motion),
		Spacing:   rx.Of(tokens.Spacing),
		Radius:    rx.Of(tokens.Radius),
		Elevation: rx.Of(tokens.Elevation),
	}
}
