// Package transition provides theme-token-specific colour interpolation
// helpers built on [github.com/vibrantgio/pulse/tween].
//
// pulse/tween owns the generic Tween[T] machinery and the per-channel
// LerpNRGBA primitive; this package keeps the integration with
// [tokens.ColorTokens] (LerpColorTokens, ColorTokensTween) that bridges
// generic tweening to the prism theme contract.
package transition

import (
	"github.com/vibrantgio/prism/tokens"
	"github.com/vibrantgio/pulse/tween"
)

// LerpColorTokens interpolates each colour field of two
// [tokens.ColorTokens] using [tween.LerpNRGBA].
func LerpColorTokens(from, to tokens.ColorTokens, t float64) tokens.ColorTokens {
	return tokens.ColorTokens{
		Background:       tween.LerpNRGBA(from.Background, to.Background, t),
		OnBackground:     tween.LerpNRGBA(from.OnBackground, to.OnBackground, t),
		Surface:          tween.LerpNRGBA(from.Surface, to.Surface, t),
		OnSurface:        tween.LerpNRGBA(from.OnSurface, to.OnSurface, t),
		SurfaceVariant:   tween.LerpNRGBA(from.SurfaceVariant, to.SurfaceVariant, t),
		OnSurfaceVariant: tween.LerpNRGBA(from.OnSurfaceVariant, to.OnSurfaceVariant, t),
		Primary:          tween.LerpNRGBA(from.Primary, to.Primary, t),
		OnPrimary:        tween.LerpNRGBA(from.OnPrimary, to.OnPrimary, t),
		Secondary:        tween.LerpNRGBA(from.Secondary, to.Secondary, t),
		OnSecondary:      tween.LerpNRGBA(from.OnSecondary, to.OnSecondary, t),
		Error:            tween.LerpNRGBA(from.Error, to.Error, t),
		OnError:          tween.LerpNRGBA(from.OnError, to.OnError, t),
		Outline:          tween.LerpNRGBA(from.Outline, to.Outline, t),
	}
}

// ColorTokensTween constructs a [tween.Tween] interpolating from a to b
// over frames frames, using [LerpColorTokens]. This is the integration
// bridge between the generic Tween and the prism theme contract.
func ColorTokensTween(a, b tokens.ColorTokens, frames int) tween.Tween[tokens.ColorTokens] {
	return tween.Tween[tokens.ColorTokens]{
		From:   a,
		To:     b,
		Frames: frames,
		Lerp:   LerpColorTokens,
	}
}
