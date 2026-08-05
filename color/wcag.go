// WCAG 2.x relative luminance and contrast ratio, after the definitions
// published in the WCAG 2 specification (w3.org/TR/WCAG21, "relative
// luminance" and "contrast ratio"). Per ADR-007 these are the
// reported-but-not-gating contrast metric in the design system: contrast
// gating uses APCA (arriving with the contrast engine, D2.4), while WCAG 2
// ratios are computed for reporting and for tooling that still speaks
// them. New code, not part of the reactivego/luminance lift.
package color

import (
	stdcolor "image/color"
	"math"
)

// RelativeLuminance returns the WCAG 2.x relative luminance of an
// image/color NRGBA value, in [0,1] — 0 for black, 1 for white. The alpha
// channel is ignored: NRGBA is non-premultiplied, so the colour channels
// are the colour regardless of coverage.
//
// The linearization deliberately uses WCAG 2's own published piecewise
// threshold, V ≤ 0.03928, rather than the 0.04045 of the standard sRGB
// EOTF (LinearFromSRGB): the WCAG 2 definition carries the older sRGB
// draft constant, and conformance tooling expects results computed with
// the spec's own numbers. For 8-bit channels the two thresholds never
// straddle a representable value, so the results are identical in
// practice — the spec's numbers are used so this code matches its source
// line for line.
func RelativeLuminance(c stdcolor.NRGBA) float64 {
	lin := func(V8 uint8) float64 {
		V := float64(V8) / 255.0
		if V <= 0.03928 {
			return V / 12.92
		}
		return math.Pow((V+0.055)/1.055, 2.4)
	}
	// WCAG 2's published luminance coefficients (four significant digits,
	// unlike the seven-digit Lindbloom row in linearY — again the spec's
	// own numbers, kept verbatim).
	return 0.2126*lin(c.R) + 0.7152*lin(c.G) + 0.0722*lin(c.B)
}

// ContrastRatio returns the WCAG 2.x contrast ratio between two
// image/color NRGBA values: (L1 + 0.05) / (L2 + 0.05) with L1 the lighter
// of the two relative luminances, so the result is order-independent and
// ranges from 1 (identical luminance) to 21 (white on black). Per ADR-007
// this ratio is reported, not gated on — APCA is the gating metric.
func ContrastRatio(a, b stdcolor.NRGBA) float64 {
	la, lb := RelativeLuminance(a), RelativeLuminance(b)
	if la < lb {
		la, lb = lb, la
	}
	return (la + 0.05) / (lb + 0.05)
}
