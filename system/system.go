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
	// [Accent] enum. On macOS the darwin shim maps the raw
	// AppleAccentColor key (-1 graphite, 0..6 red through pink, absent =
	// multicolour) onto it; platforms without an accent source report the
	// zero value. The zero value, AccentDefault, means "no accent
	// override", so the zero Appearance keeps the theme's own palette.
	Accent Accent
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
// is false (no palette option given) a non-default Appearance.Accent
// overrides the pair with the accent seed's derived pair; byAccent caches
// those derivations so tokens.FromSeed runs once per distinct accent
// value, not once per emission.
type palette struct {
	light, dark tokens.ColorTokens
	pinned      bool // an explicit option chose the pair; ignore the OS accent

	mu       sync.Mutex
	byAccent map[Accent]colorPair
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
// OS accent live: a non-default [Accent] emits tokens.FromSeed of the
// accent's seed colour (Apple's published system colour — the light
// primary is that seed byte-for-byte per ADR-007), and AccentDefault —
// multicolour, unset, or a platform without an accent — emits
// tokens.DefaultLight/DefaultDark. An accent change re-emits the theme
// with the new pair; each pair is derived once per accent value and cached.
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
	light, dark := p.pair(a.Accent)
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

// pair resolves the light/dark pair for an accent, applying the precedence
// rule: a pinned palette (explicit WithSeed/WithPalette) always wins; then
// a non-default accent yields its seed's derived pair; AccentDefault (and
// any unknown value) falls back to the palette's own pair. Derived pairs
// are cached per accent value — tokens.FromSeed runs on first sight of an
// accent, not on every emission. The mutex covers concurrent subscriptions
// to one observable, which share this palette.
func (p *palette) pair(a Accent) (light, dark tokens.ColorTokens) {
	if p.pinned {
		return p.light, p.dark
	}
	seed, ok := a.Seed()
	if !ok {
		return p.light, p.dark
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if c, ok := p.byAccent[a]; ok {
		return c.light, c.dark
	}
	l, d := tokens.FromSeed(seed)
	if p.byAccent == nil {
		p.byAccent = make(map[Accent]colorPair)
	}
	p.byAccent[a] = colorPair{light: l, dark: d}
	return l, d
}
