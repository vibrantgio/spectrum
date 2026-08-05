package color_test

import (
	stdcolor "image/color"
	"math"
	"testing"

	"github.com/vibrantgio/spectrum/color"
)

// TestRelativeLuminanceAnchors pins WCAG relative luminance at the values
// the spec's own numbers force: black 0, white 1, and each pure primary
// exactly its luminance coefficient (channel 1.0 linearizes to 1.0 in
// both piecewise branches' meeting of the formula).
func TestRelativeLuminanceAnchors(t *testing.T) {
	for _, c := range []struct {
		name string
		in   stdcolor.NRGBA
		want float64
	}{
		{"black", stdcolor.NRGBA{A: 255}, 0},
		{"white", stdcolor.NRGBA{R: 255, G: 255, B: 255, A: 255}, 1},
		{"red", stdcolor.NRGBA{R: 255, A: 255}, 0.2126},
		{"green", stdcolor.NRGBA{G: 255, A: 255}, 0.7152},
		{"blue", stdcolor.NRGBA{B: 255, A: 255}, 0.0722},
	} {
		if got := color.RelativeLuminance(c.in); math.Abs(got-c.want) > 1e-12 {
			t.Errorf("RelativeLuminance(%s) = %.13f, want %g", c.name, got, c.want)
		}
	}
	// Alpha is documented as ignored.
	opaque := stdcolor.NRGBA{R: 0x67, G: 0x50, B: 0xa4, A: 0xff}
	translucent := stdcolor.NRGBA{R: 0x67, G: 0x50, B: 0xa4, A: 0x40}
	if color.RelativeLuminance(opaque) != color.RelativeLuminance(translucent) {
		t.Error("RelativeLuminance depends on alpha; it must ignore it")
	}
}

// TestContrastRatio pins the ratio's range and symmetry: white on black is
// exactly 21 ((1+0.05)/(0+0.05)), a colour against itself is exactly 1,
// the argument order never matters, and #767676 on white sits just above
// the familiar 4.5:1 AA threshold (the classic "lightest AA grey").
func TestContrastRatio(t *testing.T) {
	white := stdcolor.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := stdcolor.NRGBA{A: 255}
	if got := color.ContrastRatio(white, black); math.Abs(got-21) > 1e-12 {
		t.Errorf("ContrastRatio(white, black) = %.13f, want 21", got)
	}
	seed := stdcolor.NRGBA{R: 0x67, G: 0x50, B: 0xa4, A: 0xff}
	if got := color.ContrastRatio(seed, seed); got != 1 {
		t.Errorf("ContrastRatio(seed, seed) = %v, want exactly 1", got)
	}
	if ab, ba := color.ContrastRatio(seed, white), color.ContrastRatio(white, seed); ab != ba {
		t.Errorf("ContrastRatio is order-dependent: %v vs %v", ab, ba)
	}
	grey := stdcolor.NRGBA{R: 0x76, G: 0x76, B: 0x76, A: 0xff}
	if got := color.ContrastRatio(grey, white); got < 4.5 || got > 4.6 {
		t.Errorf("ContrastRatio(#767676, white) = %.4f, want in (4.5, 4.6)", got)
	}
}
