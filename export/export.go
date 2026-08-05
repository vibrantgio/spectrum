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
	Density    tokens.Density
	Motion     tokens.MotionScale
	Spacing    tokens.SpacingScale
	Radius     tokens.RadiusScale
	Elevation  tokens.ElevationScale
}

// Capture collects the first emission of each observable a serialisation
// needs — Color, Typography, Density, Motion, Spacing, Radius and Elevation
// — into a Snapshot. (Type is not consumed: it duplicates Typography's
// sizes.)
//
// The colour emission must be a seed-derived light scheme: FromSeed pins
// the light primary base to the seed exactly, so Capture recovers the seed
// from the emission's Primary and regenerates the pair. An emission
// FromSeed cannot reproduce — a dark scheme, or hand-assembled tokens — is
// an error, because theme.json could not honestly claim to reproduce it.
// The density emission must likewise be one of the two published settings —
// tokens.Comfortable or tokens.Compact — because theme.json records density
// as a named setting plus both settings' metrics, not as free-form numbers.
func Capture(th theme.Theme) (Snapshot, error) {
	var s Snapshot
	if th.Color == nil || th.Typography == nil || th.Density == nil || th.Motion == nil || th.Spacing == nil || th.Radius == nil || th.Elevation == nil {
		return s, fmt.Errorf("export: Capture: theme has nil observables; every consumed field of theme.Theme must be set")
	}
	var err error
	if s.Light, err = th.Color.First(); err != nil {
		return s, fmt.Errorf("export: Capture: Color: %w", err)
	}
	if s.Typography, err = th.Typography.First(); err != nil {
		return s, fmt.Errorf("export: Capture: Typography: %w", err)
	}
	if s.Density, err = th.Density.First(); err != nil {
		return s, fmt.Errorf("export: Capture: Density: %w", err)
	}
	if s.Motion, err = th.Motion.First(); err != nil {
		return s, fmt.Errorf("export: Capture: Motion: %w", err)
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

	if _, ok := densitySetting(s.Density); !ok {
		return s, fmt.Errorf("export: Capture: the density emission is neither tokens.Comfortable nor tokens.Compact; theme.json records density as a named setting, so only the published settings are reproducible")
	}
	return s, nil
}

// densitySetting names a density emission: the setting string theme.json
// records, and whether the emission is one of the two published settings.
func densitySetting(d tokens.Density) (string, bool) {
	switch d {
	case tokens.Comfortable:
		return "comfortable", true
	case tokens.Compact:
		return "compact", true
	}
	return "", false
}

// Write renders s into dir as the full Claude Design project layout —
// theme.json, styles.css, readme.md and the foundation pages under
// foundations/ — creating directories as needed. Existing files are
// overwritten: the tree is generated output, regenerated whole.
func Write(dir string, s Snapshot) error {
	if err := os.MkdirAll(filepath.Join(dir, "foundations"), 0o755); err != nil {
		return fmt.Errorf("export: Write: %w", err)
	}
	js, err := themeJSON(s)
	if err != nil {
		return fmt.Errorf("export: Write: %w", err)
	}
	files := []struct {
		name    string
		content []byte
	}{
		{"theme.json", js},
		{"styles.css", []byte(stylesCSS(s))},
		{"readme.md", []byte(readmeMD(s))},
		{filepath.Join("foundations", "color.html"), []byte(colorHTML(s))},
		{filepath.Join("foundations", "type.html"), []byte(typeHTML(s))},
		{filepath.Join("foundations", "layout.html"), []byte(layoutHTML(s))},
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(dir, f.name), f.content, 0o644); err != nil {
			return fmt.Errorf("export: Write: %w", err)
		}
	}
	return nil
}
