// Package system publishes the operating system's appearance — dark mode
// and the accent colour — as a reactive stream, and bridges it to the theme
// the components above read. A per-OS shim reads the live state behind a
// [Source]; [FromSource] turns a Source plus a poll interval into an
// rx.Observable that emits only when the value changes; [Live] wires the
// current platform's shim, and [LiveTheme] maps that stream to
// [theme.Theme] values whose Color matches the OS setting.
//
// Reach for it as the theme argument of a window: LiveTheme(time.Second) is
// what every workbench application hands to spectrum/window, and from there
// an appearance change reaches every component with no application code.
// Pass your own Source to [FromSource] or [FromSourceTheme] to stub the OS
// out in a test. The package never imports Gio — it speaks to the OS
// directly, so it is usable with or without a window.
//
// Only macOS has a real source. The Linux and Windows shims compile, link
// and return the zero [Appearance] forever, so on those platforms Live
// emits light mode once and never again: an application that looks like it
// is ignoring the system setting is not misconfigured, it is running on a
// stub. Live implementations are a later milestone.
//
// The stream is cold, and that costs more than it looks. Every subscription
// starts its own ticker and polls the Source independently, so one
// LiveTheme observable shared by n consumers polls n times per interval,
// not once — and on macOS each poll forks and execs `defaults`. Multicast
// it (rx Publish plus AutoConnect) if more than one consumer needs the
// theme, and keep the interval at the intended one second; the OS caches
// these values and will not report a change much sooner.
//
// Errors are invisible by design: a failing Read is folded into the zero
// Appearance rather than an error emission, so a broken source is
// indistinguishable from light mode with no accent. AccentIndex is
// likewise carried but never used — LiveTheme maps only Dark onto a
// palette, and mapping the accent to colours is a later milestone.
package system

import (
	"image/color"
	"time"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/theme"
	"github.com/vibrantgio/spectrum/tokens"
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

// Option customizes the palette pair a theme stream flips between. The
// default — no options — is tokens.DefaultLight/DefaultDark, exactly the
// pre-option behaviour. Options choose which light/dark pair is emitted;
// they never affect when emissions happen, so OS dark-mode tracking keeps
// working with a branded palette.
type Option func(*palette)

// palette is the light/dark pair an Appearance flips between.
type palette struct {
	light, dark tokens.ColorTokens
}

// newPalette applies opts over the default pair. When several palette
// options are given, the last one wins.
func newPalette(opts []Option) palette {
	p := palette{light: tokens.DefaultLight, dark: tokens.DefaultDark}
	for _, opt := range opts {
		opt(&p)
	}
	return p
}

// WithSeed derives the light/dark pair from one brand colour via
// tokens.FromSeed (derived once, up front — not per emission). The light
// primary is the seed byte-for-byte; everything else is generated per
// ADR-007. An OS accent colour is just such a seed, which is how
// accent-driven palettes will plug in.
func WithSeed(seed color.NRGBA) Option {
	return func(p *palette) {
		p.light, p.dark = tokens.FromSeed(seed)
	}
}

// WithPalette supplies both modes explicitly, for callers that need full
// control beyond what a seed derives. The appearance stream still decides
// which of the two is live.
func WithPalette(light, dark tokens.ColorTokens) Option {
	return func(p *palette) {
		p.light, p.dark = light, dark
	}
}

// LiveTheme bridges system-appearance changes to a theme.Theme stream.
// Each emission is a fresh theme.Theme whose Color field matches the OS
// dark-mode setting — by default tokens.DefaultLight or tokens.DefaultDark,
// or the injected pair when [WithSeed] or [WithPalette] is given; the
// remaining token categories use their package defaults.
//
// Accent-driven palette swaps are intentionally out of scope here — the
// AccentIndex is observed and propagated, but mapping it to a WithSeed
// palette belongs to a later spectrum milestone.
func LiveTheme(interval time.Duration, opts ...Option) rx.Observable[theme.Theme] {
	return rx.Map(Live(interval), newPalette(opts).theme)
}

// FromSourceTheme is the test-friendly variant of LiveTheme: it lets a
// caller plug in a fake Source while exercising the same Appearance →
// theme.Theme bridge, including any palette options.
func FromSourceTheme(src Source, interval time.Duration, opts ...Option) rx.Observable[theme.Theme] {
	return rx.Map(FromSource(src, interval), newPalette(opts).theme)
}

func (p palette) theme(a Appearance) theme.Theme {
	colors := p.light
	if a.Dark {
		colors = p.dark
	}
	return theme.Theme{
		Color:      rx.Of(colors),
		Type:       rx.Of(tokens.DefaultTypeScale),
		Typography: rx.Of(tokens.DefaultTypography),
		Motion:     rx.Of(tokens.Motion),
		Spacing:    rx.Of(tokens.Spacing),
		Radius:     rx.Of(tokens.Radius),
		Elevation:  rx.Of(tokens.Elevation),
	}
}
