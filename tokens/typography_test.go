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

// TestDefaultTypographyCode pins the code style, which is not an MD3 role —
// the 5×3 grid has no code slot — but a sixteenth style outside the grid:
// BodyMedium's metrics on the mono face (G-F0).
func TestDefaultTypographyCode(t *testing.T) {
	code, body := tokens.DefaultTypography.Code, tokens.DefaultTypography.BodyMedium
	if code.Typeface != "Roboto Mono" {
		t.Errorf("Code.Typeface = %q, want %q", code.Typeface, "Roboto Mono")
	}
	if code.Size != body.Size || code.LineHeight != body.LineHeight ||
		code.Tracking != body.Tracking || code.Weight != body.Weight {
		t.Errorf("Code metrics = %+v, want BodyMedium's %+v on the mono face", code, body)
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

// shapeRun shapes one string through the default shaper in the given font and
// returns its total advance and glyph IDs. A Gio GlyphID packs the face index
// the glyph resolved to, so identical strings shaped by different faces yield
// different ID sequences — face identity, not just metrics.
func shapeRun(t *testing.T, f font.Font) (fixed.Int26_6, []text.GlyphID) {
	t.Helper()
	shaper := tokens.DefaultTypography.Shaper()
	shaper.LayoutString(text.Parameters{
		Font:     f,
		PxPerEm:  fixed.I(16),
		MaxWidth: 100000,
	}, "wiiim... {mono[0] != prose}")
	var advance fixed.Int26_6
	var ids []text.GlyphID
	for g, ok := shaper.NextGlyph(); ok; g, ok = shaper.NextGlyph() {
		advance += g.Advance
		ids = append(ids, g.ID)
	}
	if len(ids) == 0 {
		t.Fatalf("font %+v: no glyphs shaped; the face did not resolve", f)
	}
	return advance, ids
}

func idsEqual(a, b []text.GlyphID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestDefaultShaperResolvesRobotoMono asserts the default shaper resolves the
// mono face at every weight and style the markdown/highlight path shapes —
// normal and bold, upright and italic. The shaper is built with system fonts
// excluded, so glyphs coming back at all proves the collection resolved the
// request; a mono advance differing from proportional Roboto's for the same
// string proves "Roboto Mono" did not fall back to Roboto (C1.2 precedent);
// and pairwise-distinct glyph-ID sequences prove the four requests resolve to
// four distinct faces — a mono italic keeps the upright's fixed pitch, so
// advances alone could not tell them apart.
func TestDefaultShaperResolvesRobotoMono(t *testing.T) {
	combos := []struct {
		name string
		font font.Font
	}{
		{"regular-normal", font.Font{Typeface: "Roboto Mono", Style: font.Regular, Weight: font.Normal}},
		{"regular-bold", font.Font{Typeface: "Roboto Mono", Style: font.Regular, Weight: font.Bold}},
		{"italic-normal", font.Font{Typeface: "Roboto Mono", Style: font.Italic, Weight: font.Normal}},
		{"italic-bold", font.Font{Typeface: "Roboto Mono", Style: font.Italic, Weight: font.Bold}},
	}
	ids := map[string][]text.GlyphID{}
	for _, c := range combos {
		monoAdvance, monoIDs := shapeRun(t, c.font)
		ids[c.name] = monoIDs

		// The same string in proportional Roboto at the same weight and style
		// must measure differently: 'w', 'i', 'm', '.' collapse to one width
		// only under the mono face.
		robotoAdvance, _ := shapeRun(t, font.Font{Typeface: "Roboto", Style: c.font.Style, Weight: c.font.Weight})
		if monoAdvance == robotoAdvance {
			t.Errorf("%s: mono advance %v equals proportional Roboto's; %q likely fell back to Roboto",
				c.name, monoAdvance, c.font.Typeface)
		}
	}
	for i, a := range combos {
		for _, b := range combos[i+1:] {
			if idsEqual(ids[a.name], ids[b.name]) {
				t.Errorf("%s and %s shaped to identical glyph IDs; the two requests collapsed onto one face",
					a.name, b.name)
			}
		}
	}
}
