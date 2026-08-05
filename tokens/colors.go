package tokens

import (
	"fmt"
	"image/color"
)

// ColorScale holds the eleven Tailwind shade stops for one hue family (50–950).
type ColorScale struct {
	C50, C100, C200, C300, C400, C500, C600, C700, C800, C900, C950 color.NRGBA
}

// Optional named palettes, taken verbatim from the Tailwind CSS v3 default
// config. They are exactly that — palettes an application may reach for by
// name — and no part of the semantic layer resolves from them: since D2.2
// every role ramp, pin and semantic colour derives from a seed (see
// FromSeed). Per ADR-002 the Tailwind values may survive only in this
// arrangement, never behind a role name.
var (
	Slate = ColorScale{
		C50:  color.NRGBA{0xf8, 0xfa, 0xfc, 0xff},
		C100: color.NRGBA{0xf1, 0xf5, 0xf9, 0xff},
		C200: color.NRGBA{0xe2, 0xe8, 0xf0, 0xff},
		C300: color.NRGBA{0xcb, 0xd5, 0xe1, 0xff},
		C400: color.NRGBA{0x94, 0xa3, 0xb8, 0xff},
		C500: color.NRGBA{0x64, 0x74, 0x8b, 0xff},
		C600: color.NRGBA{0x47, 0x55, 0x69, 0xff},
		C700: color.NRGBA{0x33, 0x41, 0x55, 0xff},
		C800: color.NRGBA{0x1e, 0x29, 0x3b, 0xff},
		C900: color.NRGBA{0x0f, 0x17, 0x2a, 0xff},
		C950: color.NRGBA{0x02, 0x06, 0x17, 0xff},
	}
	Blue = ColorScale{
		C50:  color.NRGBA{0xef, 0xf6, 0xff, 0xff},
		C100: color.NRGBA{0xdb, 0xea, 0xfe, 0xff},
		C200: color.NRGBA{0xbf, 0xdb, 0xfe, 0xff},
		C300: color.NRGBA{0x93, 0xc5, 0xfd, 0xff},
		C400: color.NRGBA{0x60, 0xa5, 0xfa, 0xff},
		C500: color.NRGBA{0x3b, 0x82, 0xf6, 0xff},
		C600: color.NRGBA{0x25, 0x63, 0xeb, 0xff},
		C700: color.NRGBA{0x1d, 0x4e, 0xd8, 0xff},
		C800: color.NRGBA{0x1e, 0x40, 0xaf, 0xff},
		C900: color.NRGBA{0x1e, 0x3a, 0x8a, 0xff},
		C950: color.NRGBA{0x17, 0x25, 0x54, 0xff},
	}
	Red = ColorScale{
		C50:  color.NRGBA{0xfe, 0xf2, 0xf2, 0xff},
		C100: color.NRGBA{0xfe, 0xe2, 0xe2, 0xff},
		C200: color.NRGBA{0xfe, 0xca, 0xca, 0xff},
		C300: color.NRGBA{0xfc, 0xa5, 0xa5, 0xff},
		C400: color.NRGBA{0xf8, 0x71, 0x71, 0xff},
		C500: color.NRGBA{0xef, 0x44, 0x44, 0xff},
		C600: color.NRGBA{0xdc, 0x26, 0x26, 0xff},
		C700: color.NRGBA{0xb9, 0x1c, 0x1c, 0xff},
		C800: color.NRGBA{0x99, 0x1b, 0x1b, 0xff},
		C900: color.NRGBA{0x7f, 0x1d, 0x1d, 0xff},
		C950: color.NRGBA{0x45, 0x0a, 0x0a, 0xff},
	}
)

// White is pure white. Unlike the named palettes above it is part of the
// semantic layer: FromSeed uses it as the on-colour over every light-mode
// pinned base.
var White = color.NRGBA{0xff, 0xff, 0xff, 0xff}

// Ramp is one colour role's nine-step functional ramp per ADR-007. Steps run
// 100–900 in hundreds and the step number carries the meaning: 100–300 are
// tinted fills, hovers and subtle borders, 500 is the mid-value reference,
// 700–900 are text over tinted fills and pressed states. Index i holds step
// (i+1)*100; use Step to address a ramp by its step number.
//
// Light and dark ramps are paired scales, not two role tables: the same step
// keeps the same job in both modes, so a component asking for neutral 200
// gets a light card on a light ground and a dark card on a dark one.
type Ramp [9]color.NRGBA

// Step returns the colour at step n, where n is one of 100, 200, … 900.
// Any other n is a programming error and panics.
func (r Ramp) Step(n int) color.NRGBA {
	if n < 100 || n > 900 || n%100 != 0 {
		panic(fmt.Sprintf("tokens: Ramp.Step(%d): step must be 100–900 in hundreds", n))
	}
	return r[n/100-1]
}

// RampSet holds the five colour-role ramps ADR-007 defines. Neutral carries
// every surface, border and text shade; the accent ramps carry each role's
// tints and text shades, while the role's base colour is pinned separately
// on ColorTokens (see ColorTokens.Primary).
type RampSet struct {
	Neutral   Ramp
	Primary   Ramp
	Secondary Ramp
	Tertiary  Ramp
	Error     Ramp
}

// ColorTokens holds the colour vocabulary consumed by every Prism component:
// ADR-007's nine-step functional ramps, the pinned role bases, and a thin
// semantic layer resolved from ramp steps. Each "On" field is the recommended
// text/icon colour rendered on top of its companion pinned base.
//
// The pinned bases exist because a brand seed rarely sits on the shared
// lightness scale: the pin reproduces the role's base colour exactly instead
// of reading a lightened approximation off a ramp step. Dark mode pins a
// dark-appropriate base rather than reusing the light pin.
//
// The MD3-only field names survive as deprecated aliases, resolved from ramp
// steps at construction, until F3.3 deletes the shims.
type ColorTokens struct {
	// Ramps holds the functional ramps, fully populated: nine steps per
	// role, generated on the shared lightness scale by FromSeed.
	Ramps RampSet

	// Pinned accent bases and their on-colours (ADR-007 "solid fill").
	Primary     color.NRGBA // pinned primary base — in a light scheme, the seed itself
	OnPrimary   color.NRGBA // text/icon over Primary
	Secondary   color.NRGBA // pinned secondary base
	OnSecondary color.NRGBA // text/icon over Secondary
	Tertiary    color.NRGBA // pinned tertiary base
	OnTertiary  color.NRGBA // text/icon over Tertiary
	Error       color.NRGBA // pinned error base
	OnError     color.NRGBA // text/icon over Error

	// The thin semantic layer (ADR-007's surface mapping). Background and
	// Text are pins — the "bg" and "text" bases E0.1 emits — while Surface
	// and Divider resolve from Neutral ramp steps at construction.
	Background color.NRGBA // pinned app background
	Text       color.NRGBA // pinned body text over Background
	Surface    color.NRGBA // card / raised surface — Ramps.Neutral.Step(200)
	// Divider is the subtle border / separator — Ramps.Neutral.Step(300),
	// except in the high-contrast variant, which resolves it from the
	// strong-border step 500 (see FromSeedHighContrast).
	Divider color.NRGBA

	// Deprecated: alias of Text; deleted in F3.3.
	OnBackground color.NRGBA
	// Deprecated: alias of Ramps.Neutral.Step(900), the body-text step;
	// deleted in F3.3.
	OnSurface color.NRGBA
	// Deprecated: alias of Ramps.Neutral.Step(300), the hovered/tinted
	// surface step; deleted in F3.3.
	SurfaceVariant color.NRGBA
	// Deprecated: alias of Ramps.Neutral.Step(700), the low-contrast text
	// step; deleted in F3.3.
	OnSurfaceVariant color.NRGBA
	// Deprecated: alias of Ramps.Neutral.Step(500), the strong border step;
	// deleted in F3.3. For separators use Divider.
	Outline color.NRGBA
}

// resolveAliases fills every field defined as a resolution of a ramp step —
// the semantic Surface and Divider, and the deprecated MD3 aliases — from
// the ramps and pins already set on t, and returns the completed value.
// dividerStep is the Neutral step Divider resolves from: 300 in the default
// derivation, 500 in the high-contrast variant. Constructing tokens through
// it is what keeps each alias byte-identical to its documented resolution;
// FromSeed and FromSeedHighContrast build both schemes through it.
func resolveAliases(t ColorTokens, dividerStep int) ColorTokens {
	t.Surface = t.Ramps.Neutral.Step(200)
	t.Divider = t.Ramps.Neutral.Step(dividerStep)
	t.OnBackground = t.Text
	t.OnSurface = t.Ramps.Neutral.Step(900)
	t.SurfaceVariant = t.Ramps.Neutral.Step(300)
	t.OnSurfaceVariant = t.Ramps.Neutral.Step(700)
	t.Outline = t.Ramps.Neutral.Step(500)
	return t
}
