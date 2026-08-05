package color_test

import (
	stdcolor "image/color"
	"math"
	"testing"

	"github.com/vibrantgio/spectrum/color"
)

// md3Stops is Material Design 3's thirteen standard tone stops.
var md3Stops = [13]int{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 95, 99, 100}

// TestToneMonotonicLuminance asserts that WCAG relative luminance strictly
// increases across the thirteen MD3 stops — tone is CIELAB L*, monotone in
// luminance, and 8-bit quantization must not flatten or invert any
// adjacent pair. Checked for the default seed's hue/chroma, a saturated
// blue at sRGB blue's own chroma (where gamut mapping collapses most of
// the chroma near the extremes), and a low-chroma neutral.
func TestToneMonotonicLuminance(t *testing.T) {
	for _, palette := range []struct {
		name   string
		hue, C float64
	}{
		{"seed #6750A4", 293.709, 0.13046},
		{"saturated blue", 264.052, 0.31321},
		{"low-chroma neutral", 90, 0.02},
	} {
		prev := math.Inf(-1)
		prevTone := -1
		for _, tone := range md3Stops {
			c := color.Tone(palette.hue, palette.C, tone)
			lum := color.RelativeLuminance(c)
			if lum <= prev {
				t.Errorf("%s: luminance not strictly increasing: tone %d has %.6f, tone %d has %.6f",
					palette.name, prevTone, prev, tone, lum)
			}
			prev, prevTone = lum, tone
		}
	}
}

// TestToneSeedRegression pins the MD3 default seed #6750A4: MD3 palettes
// place the seed at tone 40 by construction, and although the seed's exact
// L* is 40.0827 rather than 40.0, the integer stop must still reproduce
// the seed byte-exactly (the ~0.08 L* difference vanishes in 8-bit
// quantization — and gamut_test.go separately proves the solver exact at
// the seed's own L*).
func TestToneSeedRegression(t *testing.T) {
	seed := stdcolor.NRGBA{R: 0x67, G: 0x50, B: 0xa4, A: 0xff}
	seedL, _, _ := color.LabFromNRGBA(seed)
	_, seedC, seedH := color.OKLChFromNRGBA(seed)
	if math.Abs(seedL-40.0827) > 0.001 {
		t.Errorf("seed L* = %.4f, want the measured 40.0827 (tone 40 by construction)", seedL)
	}
	if got := color.Tone(seedH, seedC, 40); got != seed {
		t.Errorf("Tone(%.3f, %.5f, 40) = %v, want the seed %v byte-exactly",
			seedH, seedC, got, seed)
	}
}

// TestToneClampsOutOfRange asserts the documented clamp: tones below 0 and
// above 100 saturate to exact black and exact white, matching tones 0 and
// 100 themselves.
func TestToneClampsOutOfRange(t *testing.T) {
	white := stdcolor.NRGBA{R: 255, G: 255, B: 255, A: 255}
	black := stdcolor.NRGBA{R: 0, G: 0, B: 0, A: 255}
	for _, tone := range []int{100, 101, 120, 1000} {
		if got := color.Tone(293.709, 0.13046, tone); got != white {
			t.Errorf("Tone(…, %d) = %v, want exactly white", tone, got)
		}
	}
	for _, tone := range []int{0, -1, -40} {
		if got := color.Tone(293.709, 0.13046, tone); got != black {
			t.Errorf("Tone(…, %d) = %v, want exactly black", tone, got)
		}
	}
}
