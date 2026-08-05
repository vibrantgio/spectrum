// APCA (Accessible Perceptual Contrast Algorithm), the contrast metric
// ADR-007 gates the palette on. This implements the published APCA-W3
// version 0.0.98G-4g formula — the version documented in the apca-w3
// reference implementation (github.com/Myndex/apca-w3) and used by WCAG 3
// drafts — with its standard constants, verbatim. Per ADR-007, APCA is the
// gating metric and WCAG 2 ratios (wcag.go) are reported alongside.
// New code, not part of the reactivego/luminance lift.
package color

import (
	stdcolor "image/color"
	"math"
)

// The APCA-W3 0.0.98G-4g constants, as published in apca-w3. Names follow
// the reference implementation so the two can be compared line for line.
const (
	// Estimated screen luminance: a simple 2.4-exponent power curve per
	// channel (no piecewise toe — APCA models flare separately) with
	// APCA's own sRGB coefficients.
	apcaTRC = 2.4
	apcaRco = 0.2126729
	apcaGco = 0.7151522
	apcaBco = 0.0721750

	// Soft black clamp: luminances below blkThrs are raised toward it,
	// modeling screen flare and the eye's response near black.
	apcaBlkThrs = 0.022
	apcaBlkClmp = 1.414

	// Contrast-curve exponents: normal polarity (dark text on light
	// background) uses normBG/normTXT, reverse polarity (light on dark)
	// uses revBG/revTXT.
	apcaNormBG  = 0.56
	apcaNormTXT = 0.57
	apcaRevTXT  = 0.62
	apcaRevBG   = 0.65

	// Output scaling and low-contrast handling.
	apcaScale    = 1.14
	apcaOffset   = 0.027
	apcaLoClip   = 0.1
	apcaDeltaMin = 0.0005
)

// apcaLuminance is APCA's estimated screen luminance Ys of an sRGB colour:
// per-channel 2.4-exponent linearization weighted by APCA's coefficients.
// This is deliberately not WCAG 2's RelativeLuminance — APCA specifies its
// own transfer curve and coefficients.
func apcaLuminance(c stdcolor.NRGBA) float64 {
	ch := func(v uint8) float64 {
		return math.Pow(float64(v)/255.0, apcaTRC)
	}
	return apcaRco*ch(c.R) + apcaGco*ch(c.G) + apcaBco*ch(c.B)
}

// apcaClamp applies the soft black clamp to an estimated screen luminance.
func apcaClamp(y float64) float64 {
	if y < apcaBlkThrs {
		return y + math.Pow(apcaBlkThrs-y, apcaBlkClmp)
	}
	return y
}

// APCA returns the APCA-W3 (0.0.98G-4g) lightness contrast Lc between text
// and background colours, in the published signed convention: positive for
// dark text on a light background, negative for light text on a dark
// background, with |Lc| the magnitude — body text wants |Lc| ≥ 90,
// large/secondary text |Lc| ≥ 60. Alpha is ignored (NRGBA channels are
// non-premultiplied); pairs too close to distinguish return 0.
//
// Per ADR-007 this is the palette's gating metric: step 900 must reach
// |Lc| 90 and step 700 |Lc| 60 over the step-100/200 grounds, and each
// pinned base's on-colour |Lc| 60 over the base.
func APCA(text, background stdcolor.NRGBA) float64 {
	ytxt := apcaClamp(apcaLuminance(text))
	ybg := apcaClamp(apcaLuminance(background))
	if math.Abs(ybg-ytxt) < apcaDeltaMin {
		return 0
	}
	var sapc float64
	if ybg > ytxt { // normal polarity: dark text on light background
		sapc = (math.Pow(ybg, apcaNormBG) - math.Pow(ytxt, apcaNormTXT)) * apcaScale
		if sapc < apcaLoClip {
			return 0
		}
		return (sapc - apcaOffset) * 100
	}
	// reverse polarity: light text on dark background
	sapc = (math.Pow(ybg, apcaRevBG) - math.Pow(ytxt, apcaRevTXT)) * apcaScale
	if sapc > -apcaLoClip {
		return 0
	}
	return (sapc + apcaOffset) * 100
}
