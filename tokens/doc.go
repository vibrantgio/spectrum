// Package tokens holds the typed design values the whole system is styled
// from: the ColorTokens colour vocabulary — ADR-007's nine-step functional
// ramps, pinned role bases and thin semantic layer, derived from a brand
// seed by FromSeed — the Material Design 3 type roles, and the 4-pt spacing,
// radius, elevation and motion scales. The Tailwind eleven-stop
// scales survive only as optional named palettes, and the MD3-only colour
// names, deprecated aliases through v0.1.x, are gone as of v0.2.0: each was
// a fixed ramp step, named on ColorTokens.
//
// Reach for it when you draw something yourself and want a value that matches
// the components around it — a pane background, a gap, a corner radius, an
// animation duration — instead of inventing a number. Components do not import
// this package for their values: they read the observables on a theme.Theme,
// and DefaultLight, DefaultDark, DefaultTypography, Comfortable, Spacing,
// Radius, Elevation and Motion are what those observables carry by default.
//
// Every scale is a plain comparable struct of float32 device-independent
// pixels, except MotionScale, whose duration stops are time.Duration, and
// Typography, which also carries the font faces and the lazily built shaper
// its roles are shaped with. The
// package-level instances are variables rather than constants, so treat them
// as read-only: mutating one changes it for every consumer in the process.
// Copy and edit a value instead, which is how a custom theme is built.
package tokens
