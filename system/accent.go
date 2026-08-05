package system

import "image/color"

// Accent identifies the OS accent colour, normalized across platforms. The
// zero value, AccentDefault, means "no accent override": the multicolour
// setting on macOS, every platform whose shim has no live accent source,
// and a Source whose Read failed. That choice is what keeps the package's
// error contract honest — the zero Appearance really is "light mode with
// no accent", never a spurious red.
//
// On macOS the raw AppleAccentColor key is an integer the darwin shim maps
// onto this enum (see accentFromIndex): -1 graphite, 0 red, 1 orange,
// 2 yellow, 3 green, 4 blue, 5 purple, 6 pink; an absent key means
// multicolour → AccentDefault.
type Accent int

const (
	// AccentDefault is "no accent override": use the theme's own palette.
	AccentDefault Accent = iota
	AccentRed
	AccentOrange
	AccentYellow
	AccentGreen
	AccentBlue
	AccentPurple
	AccentPink
	AccentGraphite
)

// accentSeeds are the seed colours each accent derives its palette from:
// Apple's published System Colors (HIG "System Colors", macOS light
// appearance, sRGB). Graphite uses systemGray. The light primary base is
// this seed byte-for-byte (ADR-007 pins bases to the seed), so an accented
// button matches the OS accent exactly; the dark base is the seed's dark
// re-tone per tokens.FromSeed.
var accentSeeds = map[Accent]color.NRGBA{
	AccentRed:      {R: 0xFF, G: 0x3B, B: 0x30, A: 0xFF}, // systemRed
	AccentOrange:   {R: 0xFF, G: 0x95, B: 0x00, A: 0xFF}, // systemOrange
	AccentYellow:   {R: 0xFF, G: 0xCC, B: 0x00, A: 0xFF}, // systemYellow
	AccentGreen:    {R: 0x28, G: 0xCD, B: 0x41, A: 0xFF}, // systemGreen
	AccentBlue:     {R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF}, // systemBlue
	AccentPurple:   {R: 0xAF, G: 0x52, B: 0xDE, A: 0xFF}, // systemPurple
	AccentPink:     {R: 0xFF, G: 0x2D, B: 0x55, A: 0xFF}, // systemPink
	AccentGraphite: {R: 0x8E, G: 0x8E, B: 0x93, A: 0xFF}, // systemGray
}

// Seed returns the accent's seed colour. ok is false for AccentDefault and
// any value outside the enum — the "keep the theme's own palette" cases.
func (a Accent) Seed() (seed color.NRGBA, ok bool) {
	seed, ok = accentSeeds[a]
	return seed, ok
}
