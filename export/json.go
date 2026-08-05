package export

import (
	"encoding/json"
	"math"

	"github.com/vibrantgio/spectrum/color"
	"github.com/vibrantgio/spectrum/tokens"
)

// Parameters is theme.json's shape: the generative parameters that
// reproduce the theme. tokens.FromSeed(Seed) regenerates every ramp and pin
// — the round-trip test asserts it — so the file alone rebuilds the
// palette; the pins, scales, fonts and radius are recorded alongside so a
// reader (or a prototype) need not run the generator to know them.
//
// Density and the motion set join in E5.1; they do not exist yet.
type Parameters struct {
	// Seed is the brand seed as lowercase #rrggbb; Hue and Sat are its
	// OKLCh hue (degrees, 2 decimals) and chroma (4 decimals), recorded for
	// the reader — regeneration starts from the hex.
	Seed string  `json:"seed"`
	Hue  float64 `json:"hue"`
	Sat  float64 `json:"sat"`

	// Pins are the pinned role bases per mode.
	Pins ModePins `json:"pins"`

	// Fonts names the heading and body faces. Both are Roboto until a
	// heading face exists.
	Fonts Fonts `json:"fonts"`

	// Radius is the base radius in dp — tokens.RadiusScale.Base, the sheet's
	// --radius-base.
	Radius float64 `json:"radius"`

	// Scale is the shared CIELAB L* lightness scale per mode, steps
	// 100–900, measured back from the emitted neutral ramps. It documents
	// the generator's fixed scale (ADR-007); it is not itself an input —
	// FromSeed carries it.
	Scale ModeScale `json:"scale"`
}

// ModePins carries the pinned bases for both modes.
type ModePins struct {
	Light Pins `json:"light"`
	Dark  Pins `json:"dark"`
}

// Pins records one mode's pinned bases as lowercase #rrggbb hexes. Accent
// is the primary pin — the sheet's --color-accent.
type Pins struct {
	Bg        string `json:"bg"`
	Text      string `json:"text"`
	Accent    string `json:"accent"`
	Secondary string `json:"secondary"`
	Tertiary  string `json:"tertiary"`
	Error     string `json:"error"`
}

// Fonts names the typefaces.
type Fonts struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}

// ModeScale carries the L* scale for both modes.
type ModeScale struct {
	Light [9]int `json:"light"`
	Dark  [9]int `json:"dark"`
}

// pinsOf reads one scheme's pinned bases.
func pinsOf(t tokens.ColorTokens) Pins {
	return Pins{
		Bg:        hexRGB(t.Background),
		Text:      hexRGB(t.Text),
		Accent:    hexRGB(t.Primary),
		Secondary: hexRGB(t.Secondary),
		Tertiary:  hexRGB(t.Tertiary),
		Error:     hexRGB(t.Error),
	}
}

// measuredScale reads the CIELAB L* of each ramp step back from the ramp
// itself, rounded to the nearest integer — the documented scale values are
// integral, and 8-bit quantisation keeps the measurement well within
// rounding distance.
func measuredScale(r tokens.Ramp) [9]int {
	var s [9]int
	for i, c := range r {
		L, _, _ := color.LabFromNRGBA(c)
		s[i] = int(math.Round(L))
	}
	return s
}

// parameters assembles the Parameters for a snapshot.
func parameters(s Snapshot) Parameters {
	_, chroma, hue := color.OKLChFromNRGBA(s.Seed)
	return Parameters{
		Seed:   hexRGB(s.Seed),
		Hue:    math.Round(hue*100) / 100,
		Sat:    math.Round(chroma*10000) / 10000,
		Pins:   ModePins{Light: pinsOf(s.Light), Dark: pinsOf(s.Dark)},
		Fonts:  Fonts{Heading: s.Typography.HeadlineLarge.Typeface, Body: s.Typography.BodyLarge.Typeface},
		Radius: float64(s.Radius.Base),
		Scale: ModeScale{
			Light: measuredScale(s.Light.Ramps.Neutral),
			Dark:  measuredScale(s.Dark.Ramps.Neutral),
		},
	}
}

// themeJSON renders theme.json, indented and newline-terminated.
func themeJSON(s Snapshot) ([]byte, error) {
	js, err := json.MarshalIndent(parameters(s), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(js, '\n'), nil
}
