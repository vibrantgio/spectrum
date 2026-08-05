package tokens_test

import (
	"testing"

	"gioui.org/font"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"

	"github.com/vibrantgio/spectrum/tokens"
)

func TestDefaultTypographyRolesComplete(t *testing.T) {
	roles := []struct {
		name  string
		style tokens.TextStyle
	}{
		{"DisplayLarge", tokens.DefaultTypography.DisplayLarge},
		{"DisplayMedium", tokens.DefaultTypography.DisplayMedium},
		{"DisplaySmall", tokens.DefaultTypography.DisplaySmall},
		{"HeadlineLarge", tokens.DefaultTypography.HeadlineLarge},
		{"HeadlineMedium", tokens.DefaultTypography.HeadlineMedium},
		{"HeadlineSmall", tokens.DefaultTypography.HeadlineSmall},
		{"TitleLarge", tokens.DefaultTypography.TitleLarge},
		{"TitleMedium", tokens.DefaultTypography.TitleMedium},
		{"TitleSmall", tokens.DefaultTypography.TitleSmall},
		{"LabelLarge", tokens.DefaultTypography.LabelLarge},
		{"LabelMedium", tokens.DefaultTypography.LabelMedium},
		{"LabelSmall", tokens.DefaultTypography.LabelSmall},
		{"BodyLarge", tokens.DefaultTypography.BodyLarge},
		{"BodyMedium", tokens.DefaultTypography.BodyMedium},
		{"BodySmall", tokens.DefaultTypography.BodySmall},
	}
	for _, role := range roles {
		if role.style.Size <= 0 {
			t.Errorf("%s: zero size", role.name)
		}
		if role.style.Weight <= 0 {
			t.Errorf("%s: zero weight", role.name)
		}
		if role.style.LineHeight <= 0 {
			t.Errorf("%s: zero line height", role.name)
		}
	}
}

// TestDefaultShaperResolvesRobotoEveryWeight shapes text through the default
// shaper for every distinct weight the default typography names. The shaper
// is built with system fonts excluded and only the Roboto collection loaded,
// so glyphs coming back at all proves Roboto resolved; different total
// advances between regular and medium prove the weights resolve to distinct
// faces rather than collapsing onto one.
func TestDefaultShaperResolvesRobotoEveryWeight(t *testing.T) {
	typ := tokens.DefaultTypography
	weights := map[int]bool{}
	for _, style := range []tokens.TextStyle{
		typ.DisplayLarge, typ.DisplayMedium, typ.DisplaySmall,
		typ.HeadlineLarge, typ.HeadlineMedium, typ.HeadlineSmall,
		typ.TitleLarge, typ.TitleMedium, typ.TitleSmall,
		typ.LabelLarge, typ.LabelMedium, typ.LabelSmall,
		typ.BodyLarge, typ.BodyMedium, typ.BodySmall,
	} {
		weights[style.Weight] = true
	}
	if !weights[tokens.WeightRegular] || !weights[tokens.WeightMedium] {
		t.Fatalf("default typography names weights %v, want both %d and %d",
			weights, tokens.WeightRegular, tokens.WeightMedium)
	}

	shaper := typ.Shaper()
	if again := typ.Shaper(); again != shaper {
		t.Error("Shaper() built a second shaper instead of caching the first")
	}

	advances := map[int]fixed.Int26_6{}
	for weight := range weights {
		shaper.LayoutString(text.Parameters{
			Font:     font.Font{Typeface: "Roboto", Weight: tokens.FontWeight(weight)},
			PxPerEm:  fixed.I(16),
			MaxWidth: 10000,
		}, "Weights of the world, unite")
		var advance fixed.Int26_6
		glyphs := 0
		for g, ok := shaper.NextGlyph(); ok; g, ok = shaper.NextGlyph() {
			advance += g.Advance
			glyphs++
		}
		if glyphs == 0 {
			t.Errorf("weight %d: no glyphs shaped; Roboto did not resolve", weight)
		}
		advances[weight] = advance
	}
	if advances[tokens.WeightRegular] == advances[tokens.WeightMedium] {
		t.Errorf("regular and medium shaped to identical advances (%v); "+
			"medium likely fell back to the regular face",
			advances[tokens.WeightRegular])
	}
}
