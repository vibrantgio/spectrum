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
