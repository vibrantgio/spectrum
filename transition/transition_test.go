package transition_test

import (
	"image/color"
	"testing"

	"github.com/vibrantgio/prism/tokens"
	"github.com/vibrantgio/spectrum/transition"
)

func TestTweenAtZeroReturnsFrom(t *testing.T) {
	tw := transition.Tween[int]{From: 100, To: 200, Frames: 10, Lerp: lerpInt}
	if got := tw.At(0); got != 100 {
		t.Errorf("At(0) = %d, want 100", got)
	}
}

func TestTweenAtFramesReturnsTo(t *testing.T) {
	tw := transition.Tween[int]{From: 100, To: 200, Frames: 10, Lerp: lerpInt}
	if got := tw.At(10); got != 200 {
		t.Errorf("At(Frames) = %d, want 200", got)
	}
}

func TestTweenAtBeyondFramesReturnsTo(t *testing.T) {
	tw := transition.Tween[int]{From: 100, To: 200, Frames: 10, Lerp: lerpInt}
	if got := tw.At(99); got != 200 {
		t.Errorf("At(>Frames) = %d, want 200", got)
	}
}

func TestTweenAtNegativeReturnsFrom(t *testing.T) {
	tw := transition.Tween[int]{From: 100, To: 200, Frames: 10, Lerp: lerpInt}
	if got := tw.At(-1); got != 100 {
		t.Errorf("At(-1) = %d, want 100", got)
	}
}

func TestTweenAtMidpoint(t *testing.T) {
	tw := transition.Tween[int]{From: 0, To: 100, Frames: 10, Lerp: lerpInt}
	// Midpoint of 10-frame tween should be ~50.
	if got := tw.At(5); got != 50 {
		t.Errorf("At(5) = %d, want 50", got)
	}
}

func TestTweenZeroFramesIsImmediate(t *testing.T) {
	// A zero-frame tween has no interpolation phase; any frame returns From.
	// (Callers should not normally construct a zero-frame tween, but the
	// boundary should not panic with a divide-by-zero.)
	tw := transition.Tween[int]{From: 100, To: 200, Frames: 0, Lerp: lerpInt}
	if got := tw.At(5); got != 100 {
		t.Errorf("At(5) on zero-frame tween = %d, want From=100", got)
	}
}

func TestLerpNRGBAEndpoints(t *testing.T) {
	from := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	to := color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	if got := transition.LerpNRGBA(from, to, 0); got != from {
		t.Errorf("LerpNRGBA(_, _, 0) = %+v, want %+v", got, from)
	}
	if got := transition.LerpNRGBA(from, to, 1); got != to {
		t.Errorf("LerpNRGBA(_, _, 1) = %+v, want %+v", got, to)
	}
}

func TestLerpNRGBAClamp(t *testing.T) {
	from := color.NRGBA{R: 10, G: 20, B: 30, A: 40}
	to := color.NRGBA{R: 100, G: 110, B: 120, A: 130}
	if got := transition.LerpNRGBA(from, to, -1); got != from {
		t.Errorf("LerpNRGBA(_, _, -1) = %+v, want %+v (clamped to 0)", got, from)
	}
	if got := transition.LerpNRGBA(from, to, 2); got != to {
		t.Errorf("LerpNRGBA(_, _, 2) = %+v, want %+v (clamped to 1)", got, to)
	}
}

func TestLerpNRGBAMidpoint(t *testing.T) {
	from := color.NRGBA{R: 0, G: 0, B: 0, A: 255}
	to := color.NRGBA{R: 100, G: 200, B: 50, A: 255}
	got := transition.LerpNRGBA(from, to, 0.5)
	want := color.NRGBA{R: 50, G: 100, B: 25, A: 255}
	if got != want {
		t.Errorf("LerpNRGBA midpoint = %+v, want %+v", got, want)
	}
}

func TestLerpNRGBAReverseDirection(t *testing.T) {
	// from > to: ensure interpolation goes the other way correctly.
	from := color.NRGBA{R: 200, G: 200, B: 200, A: 255}
	to := color.NRGBA{R: 100, G: 100, B: 100, A: 255}
	got := transition.LerpNRGBA(from, to, 0.5)
	want := color.NRGBA{R: 150, G: 150, B: 150, A: 255}
	if got != want {
		t.Errorf("LerpNRGBA reverse midpoint = %+v, want %+v", got, want)
	}
}

func TestLerpColorTokensEndpoints(t *testing.T) {
	if got := transition.LerpColorTokens(tokens.DefaultLight, tokens.DefaultDark, 0); got != tokens.DefaultLight {
		t.Errorf("LerpColorTokens(_, _, 0) != DefaultLight")
	}
	if got := transition.LerpColorTokens(tokens.DefaultLight, tokens.DefaultDark, 1); got != tokens.DefaultDark {
		t.Errorf("LerpColorTokens(_, _, 1) != DefaultDark")
	}
}

func TestColorTokensTweenSettlesAtTarget(t *testing.T) {
	// The "tween settles to target" clause from G2.3 Measurable, asserted
	// at the value-equality level (not just pixel equality).
	tw := transition.ColorTokensTween(tokens.DefaultLight, tokens.DefaultDark, 30)
	if got := tw.At(30); got != tokens.DefaultDark {
		t.Errorf("Tween.At(Frames) did not settle to target: got %+v, want %+v", got, tokens.DefaultDark)
	}
}

func lerpInt(from, to int, t float64) int {
	return from + int(float64(to-from)*t+0.5)
}
