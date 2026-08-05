package export

import (
	"fmt"
	stdcolor "image/color"
	"os"
	"path/filepath"

	"github.com/vibrantgio/spectrum/theme"
	"github.com/vibrantgio/spectrum/tokens"
)

// Snapshot is one resolved theme: the first emission of each theme.Theme
// observable, with the paired dark colour scheme and the seed recovered
// from the light scheme's primary pin. It is the input Write serialises.
type Snapshot struct {
	// Seed is the brand seed the colour schemes derive from — the light
	// scheme's pinned Primary, which FromSeed guarantees is the seed
	// byte-for-byte.
	Seed stdcolor.NRGBA

	// Light is the colour scheme the theme emitted; Dark is its paired
	// scheme, FromSeed(Seed)'s dark half.
	Light, Dark tokens.ColorTokens

	Typography tokens.Typography
	Spacing    tokens.SpacingScale
	Radius     tokens.RadiusScale
	Elevation  tokens.ElevationScale
}

// Capture collects the first emission of each observable a serialisation
// needs — Color, Typography, Spacing, Radius and Elevation — into a
// Snapshot. (Type duplicates Typography's sizes and Motion is E5.1's, so
// neither is consumed.)
//
// The colour emission must be a seed-derived light scheme: FromSeed pins
// the light primary base to the seed exactly, so Capture recovers the seed
// from the emission's Primary and regenerates the pair. An emission
// FromSeed cannot reproduce — a dark scheme, or hand-assembled tokens — is
// an error, because theme.json could not honestly claim to reproduce it.
func Capture(th theme.Theme) (Snapshot, error) {
	var s Snapshot
	if th.Color == nil || th.Typography == nil || th.Spacing == nil || th.Radius == nil || th.Elevation == nil {
		return s, fmt.Errorf("export: Capture: theme has nil observables; every consumed field of theme.Theme must be set")
	}
	var err error
	if s.Light, err = th.Color.First(); err != nil {
		return s, fmt.Errorf("export: Capture: Color: %w", err)
	}
	if s.Typography, err = th.Typography.First(); err != nil {
		return s, fmt.Errorf("export: Capture: Typography: %w", err)
	}
	if s.Spacing, err = th.Spacing.First(); err != nil {
		return s, fmt.Errorf("export: Capture: Spacing: %w", err)
	}
	if s.Radius, err = th.Radius.First(); err != nil {
		return s, fmt.Errorf("export: Capture: Radius: %w", err)
	}
	if s.Elevation, err = th.Elevation.First(); err != nil {
		return s, fmt.Errorf("export: Capture: Elevation: %w", err)
	}

	s.Seed = s.Light.Primary
	light, dark := tokens.FromSeed(s.Seed)
	if light != s.Light {
		return s, fmt.Errorf("export: Capture: the colour emission is not FromSeed(%s)'s light scheme; only seed-derived light schemes are reproducible from theme.json", hexRGB(s.Seed))
	}
	s.Dark = dark
	return s, nil
}

// Write renders s into dir as theme.json and styles.css, creating dir if
// needed. Existing files are overwritten: the pair is generated output,
// regenerated whole.
func Write(dir string, s Snapshot) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("export: Write: %w", err)
	}
	css := stylesCSS(s)
	if err := os.WriteFile(filepath.Join(dir, "styles.css"), []byte(css), 0o644); err != nil {
		return fmt.Errorf("export: Write: %w", err)
	}
	js, err := themeJSON(s)
	if err != nil {
		return fmt.Errorf("export: Write: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "theme.json"), js, 0o644); err != nil {
		return fmt.Errorf("export: Write: %w", err)
	}
	return nil
}
