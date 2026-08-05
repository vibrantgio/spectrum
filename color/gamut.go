// Gamut mapping: chroma reduction at constant tone (CIELAB L*) and
// constant OKLCh hue. New code, not part of the reactivego/luminance lift,
// so like oklab.go it uses the org's XFromY naming.
//
// The problem it solves: a tonal palette holds OKLCh hue (and nominally
// chroma) constant while sweeping CIELAB L*. Near the tone extremes the
// requested (L*, C, h) lies outside sRGB, and clamping R, G and B
// independently — what the lifted RGB used to do — shifts hue and chroma
// badly. Real gamut mapping instead walks chroma toward the neutral axis,
// keeping tone and hue exact, until the colour fits.
//
// The algorithm (the same framing Google's HCT solver uses, implemented
// from scratch): tone is CIELAB L*, which depends only on relative
// luminance Y, so a requested tone fixes a target Y exactly. For a
// candidate chroma c the pair (c, h) fixes the OKLab a,b direction, and a
// bisection over OKLab lightness L finds the colour (L, c·cos h, c·sin h)
// whose linear sRGB has that target Y — realizeToneChroma below. A colour
// is feasible when that bisection brackets the target and the resulting
// linear sRGB lies inside the unit cube. An outer bisection then finds the
// largest feasible chroma ≤ the requested one; c = 0 (the grey with
// L* = tone) is always feasible, so the search cannot fail. In-gamut
// requests take a fast path: the candidate at full chroma is realized
// once, found inside the cube, and returned untouched — gamut mapping is
// the identity inside the gamut.
//
// Hue is held exactly by construction: every candidate's a,b sit on the
// requested hue ray, so the only hue movement in a mapped result is 8-bit
// quantization noise.
package color

import (
	stdcolor "image/color"
	"math"
)

// gamutEps is the linear-sRGB tolerance inside which a channel counts as
// in gamut. Two kinds of residue must fall inside it: the boundary
// overshoot the bisections leave, and the round-trip residue of the
// conversion chains, measured at up to ~7e-7 over the sRGB cube. It must
// also stay below half the smallest 8-bit quantization step (~1.5e-4 in
// linear near black), so that snapping the residue onto the boundary picks
// the same byte true gamut mapping would. 1e-4 sits safely between the
// two.
const gamutEps = 1e-4

// inSRGBGamut reports whether a linear sRGB triple is inside the unit
// cube, within gamutEps of numeric residue per channel.
func inSRGBGamut(red, green, blue float64) bool {
	const lo, hi = -gamutEps, 1 + gamutEps
	return red >= lo && red <= hi &&
		green >= lo && green <= hi &&
		blue >= lo && blue <= hi
}

// linearY is the D65 relative luminance of a linear sRGB triple: the Y row
// of the Lindbloom matrix in RGB_to_XYZ_D65, applied to already-linear
// channels. The coefficients must stay identical to that row, or the tone
// this file solves for would disagree with the tone Lab reports.
func linearY(red, green, blue float64) float64 {
	return 0.2126729*red + 0.7151522*green + 0.0721750*blue
}

// yFromTone returns the relative luminance Y that CIELAB L* = tone denotes
// (the neutral axis: a = b = 0).
func yFromTone(tone float64) float64 {
	_, Y, _ := Lab_to_XYZ_D65(tone, 0, 0)
	return Y
}

// quantizeLinear converts a linear sRGB triple to 8-bit channels. The
// clamp only snaps the ≤ gamutEps boundary residue the solvers leave — it
// is not the per-channel gamut clamp this file exists to replace; callers
// guarantee the triple is in gamut within gamutEps.
func quantizeLinear(red, green, blue float64) (R, G, B uint8) {
	q := func(v float64) uint8 {
		if v < 0 {
			v = 0
		} else if v > 1 {
			v = 1
		}
		return uint8(math.Round(255.0 * SRGBFromLinear(v)))
	}
	return q(red), q(green), q(blue)
}

// realizeToneChroma finds the linear sRGB triple with OKLCh chroma C and
// hue h whose relative luminance is targetY, by bisecting OKLab lightness
// over [0,1]. ok reports whether targetY was bracketed at all; when it was
// not, no OKLab lightness in [0,1] reaches that luminance at this chroma,
// which the chroma search treats as infeasible. The returned triple may
// still be outside the sRGB cube — the caller checks.
func realizeToneChroma(C, h, targetY float64) (red, green, blue float64, ok bool) {
	_, a, b := OKLabFromOKLCh(0, C, h)
	yAt := func(L float64) float64 {
		return linearY(LinearRGBFromOKLab(L, a, b))
	}
	lo, hi := 0.0, 1.0
	if yLo, yHi := yAt(lo), yAt(hi); targetY < yLo || targetY > yHi {
		return 0, 0, 0, false
	}
	// Y is not perfectly monotone in L far outside the gamut (it can dip
	// slightly near L = 0 at saturated blue hues), but the bisection
	// invariant yAt(lo) ≤ targetY ≤ yAt(hi) still converges on a
	// crossing, which is all the search needs.
	for i := 0; i < 48; i++ {
		if mid := (lo + hi) / 2; yAt(mid) < targetY {
			lo = mid
		} else {
			hi = mid
		}
	}
	red, green, blue = LinearRGBFromOKLab((lo+hi)/2, a, b)
	return red, green, blue, true
}

// linearRGBFromToneChromaHue is the solver behind RGBFromToneChromaHue,
// returning the mapped colour as a linear sRGB triple (each channel in
// [0,1] within gamutEps). Callers must handle tone ≤ 0 and tone ≥ 100
// themselves; those tones denote exact black and white.
func linearRGBFromToneChromaHue(tone, C, h float64) (red, green, blue float64) {
	if C < 0 {
		C = 0
	}
	targetY := yFromTone(tone)
	// Fast path: the requested chroma already fits — mapping is the
	// identity inside the gamut.
	if r, g, b, ok := realizeToneChroma(C, h, targetY); ok && inSRGBGamut(r, g, b) {
		return r, g, b
	}
	// Bisect chroma over [0, C]. Chroma 0 — the grey with L* = tone — is
	// always feasible for tone in (0,100), so the invariant "lo feasible,
	// hi infeasible" holds throughout and the loop converges on the gamut
	// boundary from inside.
	red, green, blue, _ = realizeToneChroma(0, h, targetY)
	lo, hi := 0.0, C
	for i := 0; i < 32; i++ {
		mid := (lo + hi) / 2
		if r, g, b, ok := realizeToneChroma(mid, h, targetY); ok && inSRGBGamut(r, g, b) {
			red, green, blue = r, g, b
			lo = mid
		} else {
			hi = mid
		}
	}
	return red, green, blue
}

// RGBFromToneChromaHue returns the sRGB triple in range [0,255] for the
// colour with CIELAB L* equal to tone (range [0,100]), OKLCh hue h in
// degrees, and OKLCh chroma C — reduced toward 0 at constant tone and hue
// until the colour fits sRGB, per the algorithm in the file header. In
// short: you always get the requested tone and hue, and as much of the
// requested chroma as sRGB can hold.
//
// Tone 100 is exactly white and tone 0 exactly black for every hue and
// chroma (tones outside [0,100] saturate to those); at both extremes the
// gamut narrows to the neutral axis, so this is the mapping's own limit,
// short-circuited to keep the extremes byte-exact.
func RGBFromToneChromaHue(tone, C, h float64) (R, G, B uint8) {
	if tone >= 100 {
		return 255, 255, 255
	}
	if tone <= 0 {
		return 0, 0, 0
	}
	return quantizeLinear(linearRGBFromToneChromaHue(tone, C, h))
}

// NRGBAFromToneChromaHue returns the fully opaque image/color NRGBA value
// for RGBFromToneChromaHue of the same tone, chroma and hue.
func NRGBAFromToneChromaHue(tone, C, h float64) stdcolor.NRGBA {
	R, G, B := RGBFromToneChromaHue(tone, C, h)
	return stdcolor.NRGBA{R: R, G: G, B: B, A: 0xff}
}

// mapOKLabChroma reduces OKLCh chroma at constant OKLab lightness L and
// constant hue until the colour is inside sRGB, for out-of-gamut OKLab
// input: the OKLab-lightness analogue of the tone solver, used by
// RGBFromOKLab where the caller's lightness axis is OKLab L rather than
// tone. Scaling (a,b) toward the origin holds the hue ray exactly; the
// neutral axis (scale 0) is in gamut for every L in [0,1], so the
// bisection converges on the gamut boundary from inside.
func mapOKLabChroma(L, a, b float64) (red, green, blue float64) {
	red, green, blue = LinearRGBFromOKLab(L, 0, 0)
	lo, hi := 0.0, 1.0
	for i := 0; i < 48; i++ {
		mid := (lo + hi) / 2
		if r, g, bl := LinearRGBFromOKLab(L, mid*a, mid*b); inSRGBGamut(r, g, bl) {
			red, green, blue = r, g, bl
			lo = mid
		} else {
			hi = mid
		}
	}
	return red, green, blue
}
