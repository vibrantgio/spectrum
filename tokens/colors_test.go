package tokens_test

import (
	"image/color"
	"math"
	"testing"

	"github.com/vibrantgio/spectrum/tokens"
)

// contrastPair is a named foreground/background pair for WCAG AA verification.
type contrastPair struct {
	name string
	bg   color.NRGBA
	fg   color.NRGBA
}

// tokenPairs returns the foreground/background pairs defined by the "On" naming
// convention in t. Outline has no On counterpart and is excluded.
func tokenPairs(t tokens.ColorTokens) []contrastPair {
	return []contrastPair{
		{"Background/OnBackground", t.Background, t.OnBackground},
		{"Surface/OnSurface", t.Surface, t.OnSurface},
		{"SurfaceVariant/OnSurfaceVariant", t.SurfaceVariant, t.OnSurfaceVariant},
		{"Primary/OnPrimary", t.Primary, t.OnPrimary},
		{"Secondary/OnSecondary", t.Secondary, t.OnSecondary},
		{"Error/OnError", t.Error, t.OnError},
	}
}

func linearize(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func relativeLuminance(c color.NRGBA) float64 {
	r := linearize(float64(c.R) / 255.0)
	g := linearize(float64(c.G) / 255.0)
	b := linearize(float64(c.B) / 255.0)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func contrastRatio(c1, c2 color.NRGBA) float64 {
	l1 := relativeLuminance(c1)
	l2 := relativeLuminance(c2)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

const wcagAA = 4.5

// TestRampStepAddressing verifies Step's 100–900 addressing over the
// backing array and that out-of-vocabulary steps panic.
func TestRampStepAddressing(t *testing.T) {
	var r tokens.Ramp
	for i := range r {
		r[i] = color.NRGBA{R: uint8(i + 1), A: 0xff}
	}
	for n := 100; n <= 900; n += 100 {
		want := color.NRGBA{R: uint8(n / 100), A: 0xff}
		if got := r.Step(n); got != want {
			t.Errorf("Step(%d) = %v, want %v", n, got, want)
		}
	}
	for _, bad := range []int{0, 50, 99, 150, 901, 1000, -100} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Step(%d): expected panic", bad)
				}
			}()
			r.Step(bad)
		}()
	}
}

// TestAliasesResolveFromRamps verifies every deprecated MD3 alias equals its
// documented ramp-step or pin resolution, and that the semantic layer
// resolves per ADR-007's surface mapping, in both default schemes.
func TestAliasesResolveFromRamps(t *testing.T) {
	for _, s := range []struct {
		name string
		tok  tokens.ColorTokens
	}{
		{"DefaultLight", tokens.DefaultLight},
		{"DefaultDark", tokens.DefaultDark},
	} {
		n := s.tok.Ramps.Neutral
		checks := []struct {
			name string
			got  color.NRGBA
			want color.NRGBA
		}{
			{"Surface = Neutral.Step(200)", s.tok.Surface, n.Step(200)},
			{"Divider = Neutral.Step(300)", s.tok.Divider, n.Step(300)},
			{"OnBackground = Text", s.tok.OnBackground, s.tok.Text},
			{"OnSurface = Neutral.Step(900)", s.tok.OnSurface, n.Step(900)},
			{"SurfaceVariant = Neutral.Step(300)", s.tok.SurfaceVariant, n.Step(300)},
			{"OnSurfaceVariant = Neutral.Step(700)", s.tok.OnSurfaceVariant, n.Step(700)},
			{"Outline = Neutral.Step(500)", s.tok.Outline, n.Step(500)},
		}
		for _, c := range checks {
			if c.got != c.want {
				t.Errorf("%s: %s: got %v, want %v", s.name, c.name, c.got, c.want)
			}
		}
	}
}

// TestNeutralStepsPopulated verifies the Neutral steps the aliases resolve
// from are present and opaque in both default schemes. The remaining steps
// and the accent ramps are deliberately zero until D2.2 generates them.
func TestNeutralStepsPopulated(t *testing.T) {
	for _, s := range []struct {
		name string
		tok  tokens.ColorTokens
	}{
		{"DefaultLight", tokens.DefaultLight},
		{"DefaultDark", tokens.DefaultDark},
	} {
		for _, step := range []int{200, 300, 500, 700, 900} {
			if c := s.tok.Ramps.Neutral.Step(step); c.A != 0xff {
				t.Errorf("%s: Neutral.Step(%d) = %v, want an opaque colour", s.name, step, c)
			}
		}
	}
}

// TestPreADR007ValuesUnchanged pins every pre-ADR-007 field to its exact
// former value, byte for byte, so the ramp restructuring cannot move a
// colour any consumer already renders (goldens downstream depend on these).
func TestPreADR007ValuesUnchanged(t *testing.T) {
	hex := func(r, g, b uint8) color.NRGBA { return color.NRGBA{r, g, b, 0xff} }
	for _, s := range []struct {
		name string
		tok  tokens.ColorTokens
		want map[string]color.NRGBA
	}{
		{"DefaultLight", tokens.DefaultLight, map[string]color.NRGBA{
			"Background":       hex(0xff, 0xff, 0xff),
			"OnBackground":     hex(0x0f, 0x17, 0x2a),
			"Surface":          hex(0xf8, 0xfa, 0xfc),
			"OnSurface":        hex(0x0f, 0x17, 0x2a),
			"SurfaceVariant":   hex(0xf1, 0xf5, 0xf9),
			"OnSurfaceVariant": hex(0x33, 0x41, 0x55),
			"Primary":          hex(0x1d, 0x4e, 0xd8),
			"OnPrimary":        hex(0xff, 0xff, 0xff),
			"Secondary":        hex(0x47, 0x55, 0x69),
			"OnSecondary":      hex(0xff, 0xff, 0xff),
			"Error":            hex(0xb9, 0x1c, 0x1c),
			"OnError":          hex(0xff, 0xff, 0xff),
			"Outline":          hex(0xcb, 0xd5, 0xe1),
		}},
		{"DefaultDark", tokens.DefaultDark, map[string]color.NRGBA{
			"Background":       hex(0x02, 0x06, 0x17),
			"OnBackground":     hex(0xf8, 0xfa, 0xfc),
			"Surface":          hex(0x0f, 0x17, 0x2a),
			"OnSurface":        hex(0xf1, 0xf5, 0xf9),
			"SurfaceVariant":   hex(0x1e, 0x29, 0x3b),
			"OnSurfaceVariant": hex(0xcb, 0xd5, 0xe1),
			"Primary":          hex(0x60, 0xa5, 0xfa),
			"OnPrimary":        hex(0x0f, 0x17, 0x2a),
			"Secondary":        hex(0x94, 0xa3, 0xb8),
			"OnSecondary":      hex(0x0f, 0x17, 0x2a),
			"Error":            hex(0xf8, 0x71, 0x71),
			"OnError":          hex(0x0f, 0x17, 0x2a),
			"Outline":          hex(0x33, 0x41, 0x55),
		}},
	} {
		got := map[string]color.NRGBA{
			"Background":       s.tok.Background,
			"OnBackground":     s.tok.OnBackground,
			"Surface":          s.tok.Surface,
			"OnSurface":        s.tok.OnSurface,
			"SurfaceVariant":   s.tok.SurfaceVariant,
			"OnSurfaceVariant": s.tok.OnSurfaceVariant,
			"Primary":          s.tok.Primary,
			"OnPrimary":        s.tok.OnPrimary,
			"Secondary":        s.tok.Secondary,
			"OnSecondary":      s.tok.OnSecondary,
			"Error":            s.tok.Error,
			"OnError":          s.tok.OnError,
			"Outline":          s.tok.Outline,
		}
		for name, want := range s.want {
			if got[name] != want {
				t.Errorf("%s.%s = %v, want %v", s.name, name, got[name], want)
			}
		}
	}
}

func TestWCAGAAContrast(t *testing.T) {
	schemes := []struct {
		name   string
		tokens tokens.ColorTokens
	}{
		{"DefaultLight", tokens.DefaultLight},
		{"DefaultDark", tokens.DefaultDark},
	}
	for _, s := range schemes {
		for _, p := range tokenPairs(s.tokens) {
			cr := contrastRatio(p.bg, p.fg)
			if cr < wcagAA {
				t.Errorf("%s %s: contrast ratio %.2f:1 < %.1f:1 (WCAG AA)",
					s.name, p.name, cr, wcagAA)
			}
		}
	}
}
