// Package transition animates colour interpolation across theme switches
// using a generic frame-indexed [Tween].
//
// The package is a pre-Phase 3 stub: the [Tween] type is the seed of what
// will become pulse/tween/ in Phase 3. spectrum/transition keeps the
// integration with [tokens.ColorTokens] (LerpColorTokens, ColorTokensTween)
// that bridges generic tweening to the prism theme contract; when
// pulse/tween/ lands, [Tween] moves there and this package keeps only the
// theme-specific helpers.
//
// Tween is intentionally minimal: a value type with no internal state, no
// easing, and no clock. Callers drive it by passing a frame index to
// [Tween.At]. The frame index is whatever the caller chooses to use as
// "time" — render-loop frames, real-time samples, or test fixtures.
package transition

import (
	"image/color"

	"github.com/vibrantgio/prism/tokens"
)

// Tween is a frame-indexed interpolator. Construct one with From, To,
// Frames, and a Lerp function; call [Tween.At] to read the value at a
// given frame.
//
// Frame semantics: At(0) returns From, At(n >= Frames) returns To, and
// frames in between are interpolated by Lerp at parameter n/Frames.
type Tween[T any] struct {
	From, To T
	Frames   int
	Lerp     func(from, to T, t float64) T
}

// At returns the tweened value at frame n. n is clamped to [0, Frames]:
// n <= 0 returns From; n >= Frames returns To.
func (tw Tween[T]) At(n int) T {
	if n <= 0 || tw.Frames <= 0 {
		return tw.From
	}
	if n >= tw.Frames {
		return tw.To
	}
	return tw.Lerp(tw.From, tw.To, float64(n)/float64(tw.Frames))
}

// LerpNRGBA linearly interpolates each channel of two NRGBA colours at
// parameter t. t is clamped to [0, 1]. Inputs are treated as straight
// alpha (no premultiplication, no gamma correction); perceptual blending
// is a Phase 3 concern.
func LerpNRGBA(from, to color.NRGBA, t float64) color.NRGBA {
	if t <= 0 {
		return from
	}
	if t >= 1 {
		return to
	}
	return color.NRGBA{
		R: lerpByte(from.R, to.R, t),
		G: lerpByte(from.G, to.G, t),
		B: lerpByte(from.B, to.B, t),
		A: lerpByte(from.A, to.A, t),
	}
}

// lerpByte interpolates a single byte channel and rounds to nearest.
// The result is in [min(from,to), max(from,to)] for t in [0,1], so the
// "+0.5 then truncate" round-half-up trick is safe.
func lerpByte(from, to uint8, t float64) uint8 {
	return uint8(float64(from) + (float64(to)-float64(from))*t + 0.5)
}

// LerpColorTokens interpolates each [color.NRGBA] field of two
// [tokens.ColorTokens] using [LerpNRGBA].
func LerpColorTokens(from, to tokens.ColorTokens, t float64) tokens.ColorTokens {
	return tokens.ColorTokens{
		Background:       LerpNRGBA(from.Background, to.Background, t),
		OnBackground:     LerpNRGBA(from.OnBackground, to.OnBackground, t),
		Surface:          LerpNRGBA(from.Surface, to.Surface, t),
		OnSurface:        LerpNRGBA(from.OnSurface, to.OnSurface, t),
		SurfaceVariant:   LerpNRGBA(from.SurfaceVariant, to.SurfaceVariant, t),
		OnSurfaceVariant: LerpNRGBA(from.OnSurfaceVariant, to.OnSurfaceVariant, t),
		Primary:          LerpNRGBA(from.Primary, to.Primary, t),
		OnPrimary:        LerpNRGBA(from.OnPrimary, to.OnPrimary, t),
		Secondary:        LerpNRGBA(from.Secondary, to.Secondary, t),
		OnSecondary:      LerpNRGBA(from.OnSecondary, to.OnSecondary, t),
		Error:            LerpNRGBA(from.Error, to.Error, t),
		OnError:          LerpNRGBA(from.OnError, to.OnError, t),
		Outline:          LerpNRGBA(from.Outline, to.Outline, t),
	}
}

// ColorTokensTween constructs a [Tween] interpolating from a to b over
// frames frames, using [LerpColorTokens]. This is the integration bridge
// between the generic [Tween] and the prism theme contract.
func ColorTokensTween(a, b tokens.ColorTokens, frames int) Tween[tokens.ColorTokens] {
	return Tween[tokens.ColorTokens]{
		From:   a,
		To:     b,
		Frames: frames,
		Lerp:   LerpColorTokens,
	}
}
