package color_test

import (
	stdcolor "image/color"
	"math"
	"testing"

	"github.com/vibrantgio/spectrum/color"
)

// Tolerances for asserting on the 8-bit mapped result, calibrated against
// the quantization grid rather than the solver (whose own residue is
// ~1e-14):
//
//   - toneTol: near black one 8-bit step moves L* by up to ~0.27, so the
//     result's measured tone can sit up to ~0.22 from the request while
//     still being the nearest representable colour.
//   - hueDriftFloor: the chroma floor for the ≤ 1° hue assertion. Hue is
//     held exactly by construction (every candidate sits on the requested
//     hue ray), so the only drift in a mapped result is quantization: one
//     8-bit step moves OKLab a,b by up to ~2e-3, which at chroma 0.1 is
//     already ~1° of angle. Below the floor hue is quantization noise,
//     not mapping error, and is not asserted.
const (
	toneTol       = 0.3
	hueDriftFloor = 0.1
)

// hueDiff returns the angular distance between two hues in degrees, in
// [0,180].
func hueDiff(a, b float64) float64 {
	d := math.Abs(a - b)
	if d > 180 {
		d = 360 - d
	}
	return d
}

// checkMapped asserts the invariants of a mapped result: it realizes the
// requested tone (which also proves it is genuinely in gamut — a
// per-channel clamp fakes the bytes but lands on the wrong L*, e.g. the
// naive tone-100 seed result #ffefff has L* ≈ 95, not 100), and above the
// chroma floor it keeps the requested hue within 1°.
func checkMapped(t *testing.T, tone, C, h float64) stdcolor.NRGBA {
	t.Helper()
	got := color.NRGBAFromToneChromaHue(tone, C, h)
	L, _, _ := color.LabFromNRGBA(got)
	if math.Abs(L-tone) > toneTol {
		t.Errorf("tone %g C %g h %g: mapped %v has L* %.4f, want %g ±%g",
			tone, C, h, got, L, tone, toneTol)
	}
	_, mc, mh := color.OKLChFromNRGBA(got)
	if mc >= hueDriftFloor {
		if d := hueDiff(mh, h); d > 1 {
			t.Errorf("tone %g C %g h %g: mapped %v drifted hue by %.4f° to %.3f° (>1°)",
				tone, C, h, got, d, mh)
		}
	}
	return got
}

// TestToneExtremesExactEveryHue asserts tone 100 is exactly white and
// tone 0 exactly black for every hue (1° sweep) at zero, moderate and
// impossible chroma, and that tones outside [0,100] saturate to the same
// extremes.
func TestToneExtremesExactEveryHue(t *testing.T) {
	white := stdcolor.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := stdcolor.NRGBA{R: 0, G: 0, B: 0, A: 255}
	for h := 0.0; h < 360; h++ {
		for _, C := range []float64{0, 0.13, 0.4} {
			if got := color.NRGBAFromToneChromaHue(100, C, h); got != white {
				t.Fatalf("tone 100 C %g h %g = %v, want exactly white", C, h, got)
			}
			if got := color.NRGBAFromToneChromaHue(0, C, h); got != black {
				t.Fatalf("tone 0 C %g h %g = %v, want exactly black", C, h, got)
			}
		}
	}
	if got := color.NRGBAFromToneChromaHue(120, 0.4, 264); got != white {
		t.Errorf("tone 120 = %v, want white (saturated)", got)
	}
	if got := color.NRGBAFromToneChromaHue(-7, 0.4, 264); got != black {
		t.Errorf("tone -7 = %v, want black (saturated)", got)
	}
}

// TestGamutMapHardCases works the saturated blues and purples of the plan:
// sRGB blue's own OKLCh chroma (0.31321 at hue 264.052) and an impossible
// 0.4, at the extreme tones 0, 90, 95, 99 and 100 where the gamut narrows
// to a sliver around the neutral axis, plus mid tones 30 and 50 where
// enough chroma survives for the 1° hue assertion to bite.
func TestGamutMapHardCases(t *testing.T) {
	white := stdcolor.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := stdcolor.NRGBA{R: 0, G: 0, B: 0, A: 255}
	for _, h := range []float64{264.052, 300} {
		for _, C := range []float64{0.31321, 0.4} {
			for _, tone := range []float64{0, 30, 50, 90, 95, 99, 100} {
				got := checkMapped(t, tone, C, h)
				switch tone {
				case 100:
					if got != white {
						t.Errorf("tone 100 C %g h %g = %v, want white", C, h, got)
					}
				case 0:
					if got != black {
						t.Errorf("tone 0 C %g h %g = %v, want black", C, h, got)
					}
				}
			}
		}
	}
	// Near-extreme tones must converge achromatic-consistently for every
	// hue: no panics, no hue flips into wildly wrong tones — the mapped
	// result realizes the requested tone even where chroma has collapsed.
	for h := 0.0; h < 360; h += 5 {
		checkMapped(t, 1, 0.4, h)
		checkMapped(t, 99, 0.4, h)
	}
}

// TestGamutMapSweep asserts the two mapping invariants — result in gamut
// at the requested tone, hue within 1° above the chroma floor — across a
// broad grid of hues, tones and chromas, in and out of gamut.
func TestGamutMapSweep(t *testing.T) {
	for h := 0.0; h < 360; h += 5 {
		for _, tone := range []float64{1, 5, 10, 20, 30, 40, 50, 60, 70, 80, 90, 95, 99} {
			for _, C := range []float64{0.05, 0.1, 0.2, 0.3, 0.4} {
				checkMapped(t, tone, C, h)
			}
		}
	}
}

// TestGamutMapIdentity asserts gamut mapping is the identity inside the
// gamut: requesting exactly the tone, chroma and hue of a representable
// sRGB colour returns that colour byte-for-byte.
func TestGamutMapIdentity(t *testing.T) {
	for _, c := range []stdcolor.NRGBA{
		{R: 0x67, G: 0x50, B: 0xa4, A: 0xff}, // the MD3 default seed
		{R: 128, G: 128, B: 128, A: 0xff},
		{R: 200, G: 100, B: 50, A: 0xff},
		{R: 10, G: 200, B: 250, A: 0xff},
		{R: 1, G: 1, B: 1, A: 0xff},
		{R: 254, G: 254, B: 254, A: 0xff},
	} {
		tone, _, _ := color.LabFromNRGBA(c)
		_, C, h := color.OKLChFromNRGBA(c)
		if got := color.NRGBAFromToneChromaHue(tone, C, h); got != c {
			t.Errorf("identity broken: (tone %.4f, C %.5f, h %.3f) = %v, want %v",
				tone, C, h, got, c)
		}
	}
}

// TestSeedPaletteRegression pins the plan's measured facts about the MD3
// default seed #6750A4. naive reproduces the pre-gamut-mapping palette:
// realize (tone, C, h) by the same lightness bisection the solver uses,
// then clamp each linear channel independently. Tones 10-70 were already
// exact under that clamp, so gamut mapping must leave them unchanged;
// tones 100 and 0 were #ffefff and #01003f (via the clamped CIELAB path,
// asserted separately below) and must now be exactly white and black.
func TestSeedPaletteRegression(t *testing.T) {
	seed := stdcolor.NRGBA{R: 0x67, G: 0x50, B: 0xa4, A: 0xff}
	_, seedC, seedH := color.OKLChFromNRGBA(seed)
	naive := func(tone float64) stdcolor.NRGBA {
		_, a, b := color.OKLabFromOKLCh(0, seedC, seedH)
		_, targetY, _ := color.Lab_to_XYZ_D65(tone, 0, 0)
		y := func(L float64) float64 {
			red, green, blue := color.LinearRGBFromOKLab(L, a, b)
			return 0.2126729*red + 0.7151522*green + 0.0721750*blue
		}
		lo, hi := -0.5, 1.5
		for i := 0; i < 60; i++ {
			if mid := (lo + hi) / 2; y(mid) < targetY {
				lo = mid
			} else {
				hi = mid
			}
		}
		red, green, blue := color.LinearRGBFromOKLab((lo+hi)/2, a, b)
		q := func(v float64) uint8 {
			return uint8(math.Round(255 * color.SRGBFromLinear(math.Max(0, math.Min(v, 1)))))
		}
		return stdcolor.NRGBA{R: q(red), G: q(green), B: q(blue), A: 0xff}
	}
	for _, tone := range []float64{10, 20, 30, 40, 50, 60, 70} {
		want := naive(tone)
		if got := color.NRGBAFromToneChromaHue(tone, seedC, seedH); got != want {
			t.Errorf("seed tone %g: mapped %v, want naive-identical %v (tones 10-70 were already exact)",
				tone, got, want)
		}
	}
	white := stdcolor.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := stdcolor.NRGBA{R: 0, G: 0, B: 0, A: 255}
	if got := color.NRGBAFromToneChromaHue(100, seedC, seedH); got != white {
		t.Errorf("seed tone 100 = %v, want white", got)
	}
	if got := color.NRGBAFromToneChromaHue(0, seedC, seedH); got != black {
		t.Errorf("seed tone 0 = %v, want black", got)
	}

	// The plan's measured values #ffefff and #01003f come from the old
	// per-channel clamp on the CIELAB path, holding the seed's a,b: first
	// reproduce them with an inline copy of the retired clamp, proving
	// the reproduction measures the same defect the plan did, then assert
	// the replaced RGB maps the same inputs to white and black.
	seedL, seedA, seedB := color.LabFromNRGBA(seed)
	oldClamp := func(L, a, b float64) stdcolor.NRGBA {
		X, Y, Z := color.Lab_to_XYZ_D65(L, a, b)
		red, green, blue := color.XYZ_to_RGB_D65(X, Y, Z)
		q := func(v float64) uint8 {
			return uint8(math.Round(255 * math.Max(0, math.Min(v, 1))))
		}
		return stdcolor.NRGBA{R: q(red), G: q(green), B: q(blue), A: 0xff}
	}
	if got, want := oldClamp(100, seedA, seedB), (stdcolor.NRGBA{R: 0xff, G: 0xef, B: 0xff, A: 0xff}); got != want {
		t.Errorf("old clamp at tone 100 = %v, want the plan's measured %v", got, want)
	}
	if got, want := oldClamp(0, seedA, seedB), (stdcolor.NRGBA{R: 0x01, G: 0x00, B: 0x3f, A: 0xff}); got != want {
		t.Errorf("old clamp at tone 0 = %v, want the plan's measured %v", got, want)
	}
	if got := color.NRGBAFromLab(100, seedA, seedB); got != white {
		t.Errorf("RGB at tone 100 with seed a,b = %v, want white", got)
	}
	if got := color.NRGBAFromLab(0, seedA, seedB); got != black {
		t.Errorf("RGB at tone 0 with seed a,b = %v, want black", got)
	}
	// The seed's own tone must reproduce the seed exactly (tone 40 in the
	// palette; its exact L* is 40.08).
	if got := color.NRGBAFromToneChromaHue(seedL, seedC, seedH); got != seed {
		t.Errorf("seed at its own tone %.4f = %v, want %v", seedL, got, seed)
	}
}

// TestReplacedClampsMapNotClamp asserts the replaced converters gamut map
// instead of clamping: wildly out-of-gamut input comes back at the
// requested lightness (CIELAB L* for RGB, OKLab L for RGBFromOKLab) with
// the hue held — invariants the per-channel clamp broke.
func TestReplacedClampsMapNotClamp(t *testing.T) {
	for _, c := range []struct{ L, a, b float64 }{
		{50, -128, 128},
		{40.0827, 80, -100},
		{90, 60, -80},
	} {
		r, g, b := color.RGB(c.L, c.a, c.b)
		L, _, _ := color.Lab(r, g, b)
		if math.Abs(L-c.L) > toneTol {
			t.Errorf("RGB(%g,%g,%g) = %d,%d,%d with L* %.4f, want %g ±%g",
				c.L, c.a, c.b, r, g, b, L, c.L, toneTol)
		}
	}
	for _, c := range []struct{ L, a, b float64 }{
		{0.5, -0.4, 0.4},
		{0.9, 0.3, -0.3},
	} {
		wantH := math.Atan2(c.b, c.a) * 180 / math.Pi
		if wantH < 0 {
			wantH += 360
		}
		r, g, b := color.RGBFromOKLab(c.L, c.a, c.b)
		L, _, _ := color.OKLab(r, g, b)
		if math.Abs(L-c.L) > 0.01 {
			t.Errorf("RGBFromOKLab(%g,%g,%g) = %d,%d,%d with OKLab L %.5f, want %g ±0.01",
				c.L, c.a, c.b, r, g, b, L, c.L)
		}
		if _, mc, mh := color.OKLCh(r, g, b); mc >= hueDriftFloor && hueDiff(mh, wantH) > 1 {
			t.Errorf("RGBFromOKLab(%g,%g,%g) drifted hue to %.3f°, want %.3f° ±1°",
				c.L, c.a, c.b, mh, wantH)
		}
	}
}
