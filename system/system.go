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
// Platform support is uneven, and the matrix below is the contract; where
// a cell says "no", the shim reports the zero value for that dimension and
// an application that looks like it is ignoring the setting is not
// misconfigured — the source has nothing to read.
//
//	platform  dark mode                    accent colour
//	macOS     yes — AppleInterfaceStyle    yes — AppleAccentColor index,
//	          via `defaults read -g`       normalized to [Accent] (throttled)
//	Windows   no (always light)            yes — HKCU\Software\Microsoft\
//	                                       Windows\DWM AccentColor, an
//	                                       arbitrary colour → AccentSeed
//	Linux     no (always light)            GNOME 47+: the named accent via
//	                                       `gsettings`, mapped to libadwaita's
//	                                       published colour → AccentSeed
//	                                       KDE Plasma: kdeglobals [General]
//	                                       AccentColor r,g,b → AccentSeed
//	                                       other desktops, older GNOME, or a
//	                                       KDE scheme with no explicit accent:
//	                                       none — the default seed's palette
//	other     no (always light)            no
//
// Dark-mode sources for Windows and Linux are a later milestone. The two
// accent shapes are deliberate: macOS's accent is one of eight named
// choices, carried as the [Accent] enum; Windows and Linux accents are
// arbitrary colours, carried raw in Appearance.AccentSeed. Both feed the
// same tokens.FromSeed derivation.
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
// indistinguishable from light mode with no accent. The accent is not just
// carried: with no palette option, LiveTheme follows it — each [Accent]
// maps to Apple's published seed colour and the emitted palette is
// tokens.FromSeed of that seed, derived once per accent value and cached.
// An explicit [WithSeed] or [WithPalette] beats the OS accent: the app
// chose its brand, so the accent is ignored entirely.
package system

import (
	"image/color"
	"sync"
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

	// Accent is the OS accent colour, normalized to this package's
	// [Accent] enum — the shape for platforms whose accent is one of a
	// small named set. On macOS the darwin shim maps the raw
	// AppleAccentColor key (-1 graphite, 0..6 red through pink, absent =
	// multicolour) onto it; platforms without an enum-shaped accent report
	// the zero value. The zero value, AccentDefault, means "no accent
	// override", so the zero Appearance keeps the theme's own palette.
	Accent Accent

	// AccentSeed is the OS accent as a raw colour, for platforms whose
	// accent is an arbitrary colour rather than a named choice: the
	// Windows shim decodes the DWM AccentColor registry value into it, and
	// the Linux shim the GNOME named accent or the KDE kdeglobals RGB.
	// It is meaningful only when AccentSeedSet is true; when set it takes
	// precedence over Accent in palette resolution (an explicit WithSeed
	// or WithPalette still beats both).
	AccentSeed color.NRGBA

	// AccentSeedSet reports whether AccentSeed carries a value. A separate
	// flag rather than a sentinel colour keeps every colour — including
	// black — representable, and keeps Appearance comparable for
	// rx.DistinctUntilChanged.
	AccentSeedSet bool
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
// default — no options — is tokens.DefaultLight/DefaultDark, except that
// with no option the stream also follows the OS accent: a non-default
// [Accent] swaps in tokens.FromSeed of that accent's seed colour. Giving
// any palette option pins the pair — the app chose its brand, so the OS
// accent is ignored. Options choose which light/dark pair is emitted;
// they never affect when emissions happen, so OS dark-mode tracking keeps
// working with a branded palette.
type Option func(*palette)

// palette is the light/dark pair an Appearance flips between. When pinned
// is false (no palette option given) an OS accent — a raw AccentSeed or a
// non-default Accent — overrides the pair with the seed's derived pair;
// bySeed caches those derivations so tokens.FromSeed runs once per
// distinct seed colour, not once per emission.
type palette struct {
	light, dark tokens.ColorTokens
	pinned      bool // an explicit option chose the pair; ignore the OS accent

	mu     sync.Mutex
	bySeed map[color.NRGBA]colorPair
}

type colorPair struct {
	light, dark tokens.ColorTokens
}

// newPalette applies opts over the default pair. When several palette
// options are given, the last one wins.
func newPalette(opts []Option) *palette {
	p := &palette{light: tokens.DefaultLight, dark: tokens.DefaultDark}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// WithSeed derives the light/dark pair from one brand colour via
// tokens.FromSeed (derived once, up front — not per emission). The light
// primary is the seed byte-for-byte; everything else is generated per
// ADR-007. The pair is pinned: a stream given WithSeed ignores the OS
// accent colour.
func WithSeed(seed color.NRGBA) Option {
	return func(p *palette) {
		p.light, p.dark = tokens.FromSeed(seed)
		p.pinned = true
	}
}

// WithPalette supplies both modes explicitly, for callers that need full
// control beyond what a seed derives. The appearance stream still decides
// which of the two is live. The pair is pinned: a stream given
// WithPalette ignores the OS accent colour.
func WithPalette(light, dark tokens.ColorTokens) Option {
	return func(p *palette) {
		p.light, p.dark = light, dark
		p.pinned = true
	}
}

// LiveTheme bridges system-appearance changes to a theme.Theme stream.
// Each emission is a fresh theme.Theme whose Color field matches the OS
// dark-mode setting; the remaining token categories use their package
// defaults.
//
// Which light/dark pair flips is decided by precedence: an explicit
// [WithSeed] or [WithPalette] wins outright — the app chose its brand, and
// the OS accent is ignored. With no palette option the stream follows the
// OS accent live: a raw Appearance.AccentSeed (Windows, Linux) or a
// non-default [Accent] (macOS) emits tokens.FromSeed of that seed colour
// (the light primary is the seed byte-for-byte per ADR-007), the raw seed
// beating the enum if a source ever sets both. No accent at all —
// AccentDefault with no AccentSeed: multicolour on macOS, an unsupported
// desktop, or a failed read — emits tokens.DefaultLight/DefaultDark. An
// accent change re-emits the theme with the new pair; each pair is derived
// once per seed colour and cached.
func LiveTheme(interval time.Duration, opts ...Option) rx.Observable[theme.Theme] {
	return rx.Map(Live(interval), newPalette(opts).theme)
}

// FromSourceTheme is the test-friendly variant of LiveTheme: it lets a
// caller plug in a fake Source while exercising the same Appearance →
// theme.Theme bridge, including any palette options.
func FromSourceTheme(src Source, interval time.Duration, opts ...Option) rx.Observable[theme.Theme] {
	return rx.Map(FromSource(src, interval), newPalette(opts).theme)
}

func (p *palette) theme(a Appearance) theme.Theme {
	light, dark := p.pair(a)
	colors := light
	if a.Dark {
		colors = dark
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

// pair resolves the light/dark pair for an appearance, applying the
// precedence rule: a pinned palette (explicit WithSeed/WithPalette) always
// wins; then a raw AccentSeed (Windows registry colour, GNOME/KDE colour)
// yields its derived pair; then a non-default accent enum yields its
// seed's derived pair; the rest — AccentDefault, any unknown enum value,
// no raw seed — falls back to the palette's own pair. Derived pairs are
// cached per seed colour — tokens.FromSeed runs on first sight of a seed,
// not on every emission.
func (p *palette) pair(a Appearance) (light, dark tokens.ColorTokens) {
	if p.pinned {
		return p.light, p.dark
	}
	if a.AccentSeedSet {
		return p.seedPair(a.AccentSeed)
	}
	seed, ok := a.Accent.Seed()
	if !ok {
		return p.light, p.dark
	}
	return p.seedPair(seed)
}

// seedPair returns the memoized tokens.FromSeed derivation for one seed
// colour. The mutex covers concurrent subscriptions to one observable,
// which share this palette.
func (p *palette) seedPair(seed color.NRGBA) (light, dark tokens.ColorTokens) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if c, ok := p.bySeed[seed]; ok {
		return c.light, c.dark
	}
	l, d := tokens.FromSeed(seed)
	if p.bySeed == nil {
		p.bySeed = make(map[color.NRGBA]colorPair)
	}
	p.bySeed[seed] = colorPair{light: l, dark: d}
	return l, d
}
