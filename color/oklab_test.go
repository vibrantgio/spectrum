package color_test

import (
	stdcolor "image/color"
	"math"
	"testing"

	"github.com/vibrantgio/spectrum/color"
)

// TestOKLabRoundTripSRGBCube converts every sRGB triple through
// OKLab and through the polar OKLCh detour and back, and requires the
// original triple. The plan's ceiling is 1% per channel (±2.55 of 255);
// like the CIELAB chain, the OKLab chain achieves an exact round trip
// over the full 16.7M-point cube, so that is what is asserted. Under
// -short only a step-5 lattice (~140K points) is walked.
func TestOKLabRoundTripSRGBCube(t *testing.T) {
	step := 1
	if testing.Short() {
		step = 5 // 0,5,…,255: still hits both cube corners
	}
	maxDiff := 0
	var worst [3]int
	for r := 0; r <= 255; r += step {
		for g := 0; g <= 255; g += step {
			for b := 0; b <= 255; b += step {
				// sRGB → OKLab → OKLCh → OKLab → sRGB: one pass
				// exercises both round trips, since the direct
				// OKLab→sRGB leg is the tail of the polar one.
				L, la, lb := color.OKLab(uint8(r), uint8(g), uint8(b))
				lL, C, h := color.OKLChFromOKLab(L, la, lb)
				R, G, B := color.RGBFromOKLCh(lL, C, h)
				d := absDiff(R, uint8(r))
				if e := absDiff(G, uint8(g)); e > d {
					d = e
				}
				if e := absDiff(B, uint8(b)); e > d {
					d = e
				}
				if d > maxDiff {
					maxDiff = d
					worst = [3]int{r, g, b}
				}
				if h < 0 || h >= 360 {
					t.Fatalf("OKLCh hue out of [0,360): h = %v at %v,%v,%v", h, r, g, b)
				}
			}
		}
	}
	if maxDiff > 0 {
		t.Errorf("round trip not exact: max per-channel diff %d at %v (1%% ceiling is 2)", maxDiff, worst)
	}
}

// TestOKLabReferenceValues checks the chain against the XYZ→OKLab table
// Ottosson publishes in "A perceptual color space for image processing".
// The package has no XYZ entry point on the OKLab side, so each XYZ triple
// (Y=1 basis, as lab.go's XYZ functions use) is bridged to linear sRGB with
// the same Lindbloom XYZ→RGB matrix lab.go carries, and fed to
// OKLabFromLinearRGB — three of the four rows are far outside the sRGB
// gamut, which the linear entry point accepts. The table is printed to
// three decimals; the bridge reproduces it within ±0.002 (measured maximum
// deviation 0.0006, dominated by the independently rounded matrices).
func TestOKLabReferenceValues(t *testing.T) {
	cases := []struct {
		name    string
		X, Y, Z float64
		L, a, b float64
	}{
		{"D65 white", 0.950, 1.000, 1.089, 1.000, 0.000, 0.000},
		{"X primary", 1.000, 0.000, 0.000, 0.450, 1.236, -0.019},
		{"Y primary", 0.000, 1.000, 0.000, 0.922, -0.671, 0.263},
		{"Z primary", 0.000, 0.000, 1.000, 0.153, -1.415, -0.449},
	}
	for _, c := range cases {
		// XYZ → linear sRGB, lab.go's XYZ_to_RGB_D65 matrix without
		// the companding (these triples must stay linear, and three
		// go negative, which companding would destroy).
		red := 3.2404542*c.X - 1.5371385*c.Y - 0.4985314*c.Z
		green := -0.9692660*c.X + 1.8760108*c.Y + 0.0415560*c.Z
		blue := 0.0556434*c.X - 0.2040259*c.Y + 1.0572252*c.Z
		L, a, b := color.OKLabFromLinearRGB(red, green, blue)
		if math.Abs(L-c.L) > 0.002 {
			t.Errorf("%s: L = %.6f, want %.3f ±0.002", c.name, L, c.L)
		}
		if math.Abs(a-c.a) > 0.002 {
			t.Errorf("%s: a = %.6f, want %.3f ±0.002", c.name, a, c.a)
		}
		if math.Abs(b-c.b) > 0.002 {
			t.Errorf("%s: b = %.6f, want %.3f ±0.002", c.name, b, c.b)
		}
	}
	// sRGB white through the uint8 chain must be OKLab white.
	if L, a, b := color.OKLab(255, 255, 255); math.Abs(L-1) > 1e-6 || math.Abs(a) > 1e-6 || math.Abs(b) > 1e-6 {
		t.Errorf("OKLab(white) = %v,%v,%v, want 1,0,0 ±1e-6", L, a, b)
	}
	if L, a, b := color.OKLab(0, 0, 0); L != 0 || a != 0 || b != 0 {
		t.Errorf("OKLab(black) = %v,%v,%v, want exactly 0,0,0", L, a, b)
	}
}

// TestOKLChSpotValues checks sRGB→OKLCh against independently published
// conversions of the sRGB primaries (the values oklch.com and the CSS
// Color 4 tooling report, reproduced to five figures across
// implementations). Tolerance ±0.001 on L, ±0.002 on C, ±0.5° on h.
func TestOKLChSpotValues(t *testing.T) {
	cases := []struct {
		name    string
		R, G, B uint8
		L, C, h float64
	}{
		{"red #ff0000", 255, 0, 0, 0.62796, 0.25768, 29.234},
		{"lime #00ff00", 0, 255, 0, 0.86644, 0.29483, 142.495},
		{"blue #0000ff", 0, 0, 255, 0.45201, 0.31321, 264.052},
	}
	for _, c := range cases {
		L, C, h := color.OKLCh(c.R, c.G, c.B)
		if math.Abs(L-c.L) > 0.001 {
			t.Errorf("%s: L = %.5f, want %.5f ±0.001", c.name, L, c.L)
		}
		if math.Abs(C-c.C) > 0.002 {
			t.Errorf("%s: C = %.5f, want %.5f ±0.002", c.name, C, c.C)
		}
		if math.Abs(h-c.h) > 0.5 {
			t.Errorf("%s: h = %.3f, want %.3f ±0.5", c.name, h, c.h)
		}
	}
}

// TestOKLChNeutralHue checks that hue is reported as 0 where chroma is
// numeric noise: greys have no hue, and the polar form must not manufacture
// one from the rounded matrices' ~4e-8 residual.
func TestOKLChNeutralHue(t *testing.T) {
	for _, v := range []uint8{0, 1, 128, 254, 255} {
		if _, C, h := color.OKLCh(v, v, v); h != 0 {
			t.Errorf("OKLCh(grey %d): h = %v (C = %v), want 0 for neutral", v, h, C)
		}
	}
	// A chromatic colour must keep its hue.
	if _, C, h := color.OKLCh(0x67, 0x50, 0xa4); C < 0.1 || h == 0 {
		t.Errorf("OKLCh(#6750a4) = C %v, h %v: expected real chroma and hue", C, h)
	}
}

// TestRGBFromOKLabIsTotal checks that the inverse accepts any OKLab input:
// out-of-range L is clamped to [0,1] and out-of-gamut a,b still produce a
// triple (per-channel clamp, approximate by design until gamut mapping).
func TestRGBFromOKLabIsTotal(t *testing.T) {
	if r, g, b := color.RGBFromOKLab(1.2, 0, 0); [3]uint8{r, g, b} != [3]uint8{255, 255, 255} {
		t.Errorf("RGBFromOKLab(1.2,0,0) = %v,%v,%v, want white (L clamped to 1)", r, g, b)
	}
	if r, g, b := color.RGBFromOKLab(-0.1, 0, 0); [3]uint8{r, g, b} != [3]uint8{0, 0, 0} {
		t.Errorf("RGBFromOKLab(-0.1,0,0) = %v,%v,%v, want black (L clamped to 0)", r, g, b)
	}
	r, g, b := color.RGBFromOKLab(0.05, -0.5, 0.5)
	_ = [3]uint8{r, g, b}
}

// TestOKLabNRGBAHelpers checks the image/color adapters agree with OKLab,
// OKLCh and their inverses.
func TestOKLabNRGBAHelpers(t *testing.T) {
	in := stdcolor.NRGBA{R: 0x67, G: 0x50, B: 0xa4, A: 0xff} // the MD3 default seed
	L, a, b := color.OKLabFromNRGBA(in)
	wl, wa, wb := color.OKLab(in.R, in.G, in.B)
	if L != wl || a != wa || b != wb {
		t.Errorf("OKLabFromNRGBA(%v) = %v,%v,%v, want %v,%v,%v", in, L, a, b, wl, wa, wb)
	}
	if out := color.NRGBAFromOKLab(L, a, b); out != in {
		t.Errorf("NRGBAFromOKLab round trip = %v, want %v", out, in)
	}
	lL, C, h := color.OKLChFromNRGBA(in)
	if wL, wC, wh := color.OKLCh(in.R, in.G, in.B); lL != wL || C != wC || h != wh {
		t.Errorf("OKLChFromNRGBA(%v) = %v,%v,%v, want %v,%v,%v", in, lL, C, h, wL, wC, wh)
	}
	if out := color.NRGBAFromOKLCh(lL, C, h); out != in {
		t.Errorf("NRGBAFromOKLCh round trip = %v, want %v", out, in)
	}
	// Alpha is ignored on the way in and opaque on the way out.
	translucent := in
	translucent.A = 0x40
	if l2, a2, b2 := color.OKLabFromNRGBA(translucent); l2 != L || a2 != a || b2 != b {
		t.Errorf("OKLabFromNRGBA should ignore alpha: got %v,%v,%v", l2, a2, b2)
	}
}
