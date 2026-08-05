package color_test

import (
	stdcolor "image/color"
	"math"
	"testing"

	"github.com/vibrantgio/spectrum/color"
)

// TestRoundTripSRGBCube converts every sRGB triple to CIELAB and back and
// requires the original triple. The plan's ceiling is 1% per channel
// (±2.55 of 255); the chain actually achieves an exact round trip over the
// full 16.7M-point cube, so that is what is asserted. Under -short only a
// step-5 lattice (~140K points) is walked.
func TestRoundTripSRGBCube(t *testing.T) {
	step := 1
	if testing.Short() {
		step = 5 // 0,5,…,255: still hits both cube corners
	}
	maxDiff := 0
	var worst [3]int
	for r := 0; r <= 255; r += step {
		for g := 0; g <= 255; g += step {
			for b := 0; b <= 255; b += step {
				L, la, lb := color.Lab(uint8(r), uint8(g), uint8(b))
				R, G, B := color.RGB(L, la, lb)
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
			}
		}
	}
	if maxDiff > 0 {
		t.Errorf("round trip not exact: max per-channel diff %d at %v (1%% ceiling is 2)", maxDiff, worst)
	}
}

func absDiff(a, b uint8) int {
	if a > b {
		return int(a) - int(b)
	}
	return int(b) - int(a)
}

// TestLabReferenceValues checks the chain against published CIELAB D65/2°
// values for the sRGB primaries, secondaries and greys (Lindbloom's
// calculator, four decimals). Tolerance ±0.1 on L*, ±0.15 on a*/b*.
func TestLabReferenceValues(t *testing.T) {
	cases := []struct {
		name    string
		R, G, B uint8
		L, a, b float64
	}{
		{"red #ff0000", 255, 0, 0, 53.2408, 80.0925, 67.2032},
		{"green #00ff00", 0, 255, 0, 87.7347, -86.1827, 83.1793},
		{"blue #0000ff", 0, 0, 255, 32.2970, 79.1875, -107.8602},
		{"yellow #ffff00", 255, 255, 0, 97.1393, -21.5537, 94.4780},
		{"cyan #00ffff", 0, 255, 255, 91.1132, -48.0875, -14.1312},
		{"magenta #ff00ff", 255, 0, 255, 60.3242, 98.2343, -60.8249},
		{"white #ffffff", 255, 255, 255, 100.0000, 0.0000, 0.0000},
		{"black #000000", 0, 0, 0, 0.0000, 0.0000, 0.0000},
		{"mid grey #808080", 128, 128, 128, 53.5850, 0.0000, 0.0000},
	}
	for _, c := range cases {
		L, a, b := color.Lab(c.R, c.G, c.B)
		if math.Abs(L-c.L) > 0.1 {
			t.Errorf("%s: L* = %.4f, want %.4f ±0.1", c.name, L, c.L)
		}
		if math.Abs(a-c.a) > 0.15 {
			t.Errorf("%s: a* = %.4f, want %.4f ±0.15", c.name, a, c.a)
		}
		if math.Abs(b-c.b) > 0.15 {
			t.Errorf("%s: b* = %.4f, want %.4f ±0.15", c.name, b, c.b)
		}
	}
}

// TestRGBIsTotal checks that RGB accepts any L*a*b* input: out-of-range L*
// is clamped to [0,100] and out-of-gamut a,b still produce a triple.
func TestRGBIsTotal(t *testing.T) {
	if r, g, b := color.RGB(120, 0, 0); [3]uint8{r, g, b} != [3]uint8{255, 255, 255} {
		t.Errorf("RGB(120,0,0) = %v,%v,%v, want white (L* clamped to 100)", r, g, b)
	}
	if r, g, b := color.RGB(-7, 0, 0); [3]uint8{r, g, b} != [3]uint8{0, 0, 0} {
		t.Errorf("RGB(-7,0,0) = %v,%v,%v, want black (L* clamped to 0)", r, g, b)
	}
	// Far out of gamut: a vivid green at near-black tone. Gamut mapping
	// keeps this total; it lands on the most chromatic sRGB green that
	// still has L* = 1 (mapping semantics are tested in gamut_test.go).
	r, g, b := color.RGB(1, -128, 128)
	_ = [3]uint8{r, g, b}
}

// TestNRGBAHelpers checks the image/color adapters agree with Lab and RGB.
func TestNRGBAHelpers(t *testing.T) {
	in := stdcolor.NRGBA{R: 0x67, G: 0x50, B: 0xa4, A: 0xff} // the MD3 default seed
	L, a, b := color.LabFromNRGBA(in)
	wl, wa, wb := color.Lab(in.R, in.G, in.B)
	if L != wl || a != wa || b != wb {
		t.Errorf("LabFromNRGBA(%v) = %v,%v,%v, want %v,%v,%v", in, L, a, b, wl, wa, wb)
	}
	out := color.NRGBAFromLab(L, a, b)
	if out != in {
		t.Errorf("NRGBAFromLab round trip = %v, want %v", out, in)
	}
	// Alpha is ignored on the way in and opaque on the way out.
	translucent := in
	translucent.A = 0x40
	if l2, a2, b2 := color.LabFromNRGBA(translucent); l2 != L || a2 != a || b2 != b {
		t.Errorf("LabFromNRGBA should ignore alpha: got %v,%v,%v", l2, a2, b2)
	}
}
