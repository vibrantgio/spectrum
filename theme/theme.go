package theme

import (
	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/tokens"
)

// Theme carries an rx.Observable for each token category. Consumers subscribe
// to the fields they need and react to updates without polling.
type Theme struct {
	Color      rx.Observable[tokens.ColorTokens]
	Type       rx.Observable[tokens.TypeScale]
	Typography rx.Observable[tokens.Typography]
	Density    rx.Observable[tokens.Density]
	Motion     rx.Observable[tokens.MotionScale]
	Spacing    rx.Observable[tokens.SpacingScale]
	Radius     rx.Observable[tokens.RadiusScale]
	Elevation  rx.Observable[tokens.ElevationScale]
}

// Default returns a Theme whose every field emits the package-level default
// value once and then completes. It is the canonical starting point for static
// or test scenarios that do not need live token switching.
func Default() Theme {
	return Theme{
		Color:      rx.Of(tokens.DefaultLight),
		Type:       rx.Of(tokens.DefaultTypeScale),
		Typography: rx.Of(tokens.DefaultTypography),
		Density:    rx.Of(tokens.Comfortable),
		Motion:     rx.Of(tokens.Motion),
		Spacing:    rx.Of(tokens.Spacing),
		Radius:     rx.Of(tokens.Radius),
		Elevation:  rx.Of(tokens.Elevation),
	}
}
