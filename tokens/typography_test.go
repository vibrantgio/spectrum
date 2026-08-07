package tokens_test

import (
	"testing"

	"gioui.org/font"
	"gioui.org/text"
	"golang.org/x/image/math/fixed"

	"github.com/vibrantgio/font/notosansmono"
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

// TestDefaultShaperResolvesRobotoEveryWeight shapes text through the pinned
// shaper for every distinct weight the default typography names. It uses
// DeterministicShaper rather than Shaper deliberately: with system fonts off,
// only the Roboto collection can answer, so glyphs coming back at all proves
// Roboto resolved rather than proving this machine owns some other face.
// Different total advances between regular and medium then prove the weights
// resolve to distinct faces rather than collapsing onto one.
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

	shaper := typ.DeterministicShaper()

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

// shapeRun shapes one string through the pinned shaper in the given font and
// returns its total advance and glyph IDs. A Gio GlyphID packs the face index
// the glyph resolved to, so identical strings shaped by different faces yield
// different ID sequences — face identity, not just metrics.
func shapeRun(t *testing.T, f font.Font) (fixed.Int26_6, []text.GlyphID) {
	t.Helper()
	shaper := tokens.DefaultTypography.DeterministicShaper()
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

// TestDefaultShaperResolvesRobotoMono asserts the pinned shaper resolves the
// mono face at every weight and style the markdown/highlight path shapes —
// normal and bold, upright and italic. System fonts are off, so glyphs coming
// back at all proves the collection resolved the
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

// symbolProbes are characters outside Roboto's and Roboto Mono's coverage that
// real text genuinely contains — U+2193 is the one that exposed the defect,
// emitted by a language model into mindchat and drawn as tofu in both
// appearances. They are deliberately never rendered into a golden image: the
// face that serves them is machine-dependent under the default shaper, which
// is exactly what a golden cannot pin.
var symbolProbes = []struct {
	name string
	r    rune
}{
	{"down arrow", '↓'},
	{"right arrow", '→'},
	{"left-right arrow", '↔'},
	{"double right arrow", '⇒'},
	{"box drawing light horizontal", '─'},
	{"box drawing light vertical", '│'},
	{"box drawing light down and right", '┌'},
	{"full block", '█'},
	{"black circle", '●'},
	{"element of", '∈'},
	{"circled plus", '⊕'},
}

// Roboto itself carries ≠ ≤ ≈ ∞ − × · … ° ± — measured, not assumed — so those
// are not probes: they resolve with the collection pinned and prove nothing
// about fallback either way.

// resolvedGlyph shapes exactly one rune and reports the glyph it resolved to.
//
// A Gio GlyphID packs the face index, the ppem and the font's own glyph ID
// into one 64-bit value, glyph ID in the low 32 bits — see newGlyphID in
// gioui.org/text. Glyph ID 0 is .notdef, the missing-glyph box: a shaper that
// found no face for a rune still returns a glyph, and this is how it says so.
// Asking for the glyph rather than for pixels is the whole point — it answers
// "did this resolve to a real face?" without depending on what that face draws.
func resolvedGlyph(t *testing.T, shaper *text.Shaper, r rune) (gid uint32, faceIdx int) {
	t.Helper()
	shaper.LayoutString(text.Parameters{
		Font:     font.Font{Typeface: "Roboto"},
		PxPerEm:  fixed.I(16),
		MaxWidth: 1000,
	}, string(r))
	g, ok := shaper.NextGlyph()
	if !ok {
		t.Fatalf("U+%04X %q: shaper produced no glyph at all", r, r)
	}
	return uint32(g.ID), int(uint64(g.ID) >> 48)
}

// TestShapersAreCachedApart asserts the two configurations do not hand back
// each other's shaper. The cache was a single field before F4.2, so this is
// the assertion that keeps a second configuration from quietly aliasing the
// first — a golden test would then inherit the system fallback it exists to
// avoid, and pass everywhere until the machine changed.
func TestShapersAreCachedApart(t *testing.T) {
	// WithFaces() with nothing to add is the copy whose caches are known
	// empty; a plain copy of the package variable may already carry shapers
	// another test in this package built through it.
	typ := tokens.DefaultTypography.WithFaces()

	fallback, pinned := typ.Shaper(), typ.DeterministicShaper()
	if fallback == pinned {
		t.Fatal("Shaper() and DeterministicShaper() returned the same shaper")
	}
	if again := typ.Shaper(); again != fallback {
		t.Error("Shaper() built a second shaper instead of caching the first")
	}
	if again := typ.DeterministicShaper(); again != pinned {
		t.Error("DeterministicShaper() built a second shaper instead of caching the first")
	}

	// Built in the other order, they must still differ: the caches are two
	// fields, not one field plus a flag.
	reversed := tokens.DefaultTypography.WithFaces()
	if reversed.DeterministicShaper() == reversed.Shaper() {
		t.Error("built pinned-first, the two configurations aliased")
	}

	// And a copy with empty caches builds its own pair, sharing neither with
	// the value it came from.
	fresh := typ.WithFaces()
	if fresh.Shaper() == fallback || fresh.DeterministicShaper() == pinned {
		t.Error("a copy with cleared caches inherited the original's shapers")
	}

	// The two configurations must also disagree about what they resolve, or
	// they are separate objects doing the same thing. U+2193 is the case.
	if gid, _ := resolvedGlyph(t, pinned, '↓'); gid != 0 {
		t.Error("the pinned shaper resolved U+2193; it is not pinned to the collection")
	}
}

// TestShaperCacheSurvivesCopying is the F5.1 regression. It reproduces what an
// rx emission does to a Typography and asserts the cache survives it.
//
// Every component in this organization reaches the theme's shaper the same
// way: the theme emits a tokens.Typography, the component's map function pulls
// it out of the tuple into a local — `typ := n.Second` — and calls
// typ.Shaper(). Both accessors take pointer receivers and cache into the
// receiver, so before F5.1 that cache was written into a local that died at
// the end of the map function and was rebuilt, from sixteen embedded faces
// plus the platform's font list, on the very next emission. The pre-F5.1 tests
// missed it because they all held their Typography in a variable and called
// twice, which is the one shape production never has.
//
// The copies below are made from one source, exactly as rx.Of(…) hands the
// same value to every subscriber, and every one of them must name the same
// shaper.
func TestShaperCacheSurvivesCopying(t *testing.T) {
	source := tokens.DefaultTypography.WithFaces()

	// Copies taken before the first shaper call: nothing is built yet, so
	// this is the emission ordering that used to produce N shapers for N
	// subscribers.
	copies := make([]tokens.Typography, 4)
	for i := range copies {
		copies[i] = source
	}

	want := copies[0].Shaper()
	if want == nil {
		t.Fatal("Shaper() returned nil")
	}
	for i, c := range copies[1:] {
		if got := c.Shaper(); got != want {
			t.Errorf("copy %d built its own fallback shaper; the cache did not survive the copy", i+1)
		}
	}

	// The source itself must see the shaper its copy built — the cache belongs
	// to the value they all came from, not to whichever copy got there first.
	if got := source.Shaper(); got != want {
		t.Error("the source did not see the shaper its copy built")
	}

	// A copy taken *after* the shaper exists must see it too.
	if late := source; late.Shaper() != want {
		t.Error("a copy taken after the first call built a second fallback shaper")
	}

	// The same for the pinned configuration, built here in the other order:
	// a copy first, then the source, so neither ordering is special.
	wantPinned := copies[1].DeterministicShaper()
	if wantPinned == nil {
		t.Fatal("DeterministicShaper() returned nil")
	}
	for i, c := range copies {
		if got := c.DeterministicShaper(); got != wantPinned {
			t.Errorf("copy %d built its own pinned shaper; the cache did not survive the copy", i)
		}
	}
	if got := source.DeterministicShaper(); got != wantPinned {
		t.Error("the source did not see the pinned shaper its copy built")
	}

	// F4.2's separation must survive the shared cache: one holder, still two
	// distinct shapers. Collapsing them would make every golden in the
	// organization inherit the system fallback it exists to avoid.
	if want == wantPinned {
		t.Fatal("Shaper() and DeterministicShaper() collapsed to one shaper across copies")
	}
	if gid, _ := resolvedGlyph(t, wantPinned, '↓'); gid != 0 {
		t.Error("the shared pinned shaper resolved U+2193; it is not pinned to the collection")
	}

	// And WithFaces still detaches: a different collection is a different
	// shaper, so its copy must share nothing with the value it came from.
	wide := source.WithFaces(notosansmono.FontFace())
	if wide.Shaper() == want || wide.DeterministicShaper() == wantPinned {
		t.Error("WithFaces handed back the receiver's shapers instead of allocating a fresh cache")
	}
	if source.Shaper() != want || source.DeterministicShaper() != wantPinned {
		t.Error("WithFaces disturbed the receiver's shared cache")
	}
}

// TestDefaultTypographySharesOneShaper asserts the property the whole fix
// exists for, at the value the whole organization actually uses: every
// component reading rx.Of(tokens.DefaultTypography) gets one process-wide
// shaper per configuration, not one per emission.
func TestDefaultTypographySharesOneShaper(t *testing.T) {
	// Two independent snapshots of the package value, as two components'
	// theme wiring would take them.
	first := tokens.DefaultTypography
	second := tokens.DefaultTypography
	if first.Shaper() != second.Shaper() {
		t.Error("two snapshots of DefaultTypography built two fallback shapers")
	}
	if first.DeterministicShaper() != second.DeterministicShaper() {
		t.Error("two snapshots of DefaultTypography built two pinned shapers")
	}
	if tokens.DefaultTypography.Shaper() != first.Shaper() {
		t.Error("the package value does not share the shaper its snapshots use")
	}
}

// TestDeterministicShaperPinsTheCollection asserts the pinned shaper really is
// pinned: every symbol probe resolves to .notdef, because no face in the
// default collection carries it and system fonts are off. If this test starts
// passing glyphs back, the deterministic configuration has stopped being
// deterministic and every golden in the organization is machine-dependent.
func TestDeterministicShaperPinsTheCollection(t *testing.T) {
	typ := tokens.DefaultTypography
	shaper := typ.DeterministicShaper()
	for _, p := range symbolProbes {
		if gid, faceIdx := resolvedGlyph(t, shaper, p.r); gid != 0 {
			t.Errorf("%s U+%04X: pinned shaper resolved glyph %d on face %d; "+
				"the default collection should not carry it and system fonts should be off",
				p.name, p.r, gid, faceIdx)
		}
	}
	// The control: Latin text the collection does carry must still resolve,
	// or the assertion above would pass on a broken shaper.
	if gid, _ := resolvedGlyph(t, shaper, 'A'); gid == 0 {
		t.Error("pinned shaper failed to resolve 'A' from Roboto")
	}
}

// TestSymbolFaceResolvesSymbols is the resolution assertion the symbol face
// exists for: pinned collection, system fonts off, optional face appended, and
// every probe comes back as a real glyph from that face. It is machine-
// independent by construction — nothing here reads a font off the host — which
// is what lets a component that legitimately draws an arrow stay testable.
//
// The import is test-only on purpose. spectrum must not link 596 KB of symbol
// face into every application that imports tokens, which is why the face is
// not in DefaultTypography.Faces and why WithFaces exists instead.
func TestSymbolFaceResolvesSymbols(t *testing.T) {
	typ := tokens.DefaultTypography.WithFaces(notosansmono.FontFace())
	shaper := typ.DeterministicShaper()

	symbolFace := len(typ.Faces) - 1
	for _, p := range symbolProbes {
		gid, faceIdx := resolvedGlyph(t, shaper, p.r)
		if gid == 0 {
			t.Errorf("%s U+%04X: resolved to the missing-glyph glyph; the symbol face did not serve it", p.name, p.r)
			continue
		}
		if faceIdx != symbolFace {
			t.Errorf("%s U+%04X: resolved on face %d, want the appended symbol face %d",
				p.name, p.r, faceIdx, symbolFace)
		}
	}

	// Appending must not displace the default family: Latin text still comes
	// from Roboto, which is face 0, not from the mono face at the end.
	if _, faceIdx := resolvedGlyph(t, shaper, 'A'); faceIdx != 0 {
		t.Errorf("'A' resolved on face %d, want Roboto at 0; the appended face displaced the default family", faceIdx)
	}
}

// TestDefaultShaperFallsBackToSystemFonts asserts the reversal itself: the
// default shaper resolves characters no embedded face carries, by reaching the
// platform's fonts. This is the one test here that depends on the host, so it
// says so — and it separates "system fallback is disabled", which is the
// defect, from "this host has no fonts to fall back to", which is not our bug
// and is what a scratch container looks like.
func TestDefaultShaperFallsBackToSystemFonts(t *testing.T) {
	// The probe for the host itself: a shaper with no collection at all can
	// only draw what the system provides.
	if gid, _ := resolvedGlyph(t, text.NewShaper(), 'A'); gid == 0 {
		t.Skip("host has no system fonts at all; nothing to fall back to")
	}

	typ := tokens.DefaultTypography
	shaper := typ.Shaper()
	resolved := 0
	for _, p := range symbolProbes {
		if gid, _ := resolvedGlyph(t, shaper, p.r); gid != 0 {
			resolved++
		} else {
			t.Logf("%s U+%04X: not covered by this host's fonts", p.name, p.r)
		}
	}
	if resolved == 0 {
		t.Errorf("the default shaper resolved none of the %d symbol probes on a host that has fonts; "+
			"system fallback is off and every application draws tofu", len(symbolProbes))
	}
	// U+2193 is the character that exposed the defect. A host with fonts at
	// all carries a down arrow.
	if gid, faceIdx := resolvedGlyph(t, shaper, '↓'); gid == 0 {
		t.Error("U+2193 ↓ still shapes to the missing-glyph glyph through the default shaper")
	} else if faceIdx < len(typ.Faces) {
		t.Errorf("U+2193 ↓ resolved on face %d, inside the %d-face collection; expected a system face beyond it",
			faceIdx, len(typ.Faces))
	}
}

// TestWithFacesCopies asserts WithFaces leaves the receiver alone — including
// its already built shapers — and hands back a value whose caches are its own.
// A copy that inherited the narrower shaper would append a face and change
// nothing, silently.
func TestWithFacesCopies(t *testing.T) {
	base := tokens.DefaultTypography
	basePinned := base.DeterministicShaper()
	baseFaces := len(base.Faces)

	wide := base.WithFaces(notosansmono.FontFace())
	if len(base.Faces) != baseFaces {
		t.Errorf("receiver's Faces grew to %d from %d; WithFaces mutated it", len(base.Faces), baseFaces)
	}
	if len(wide.Faces) != baseFaces+1 {
		t.Fatalf("copy has %d faces, want %d", len(wide.Faces), baseFaces+1)
	}
	if wide.DeterministicShaper() == basePinned {
		t.Error("the copy handed back the receiver's shaper, built from the narrower collection")
	}
	if base.DeterministicShaper() != basePinned {
		t.Error("WithFaces disturbed the receiver's cached shaper")
	}

	// No aliasing of the backing array: appending to one must not be visible
	// in the other.
	if &base.Faces[0] == &wide.Faces[0] && cap(base.Faces) > baseFaces {
		t.Error("the copy shares the receiver's backing array")
	}

	// WithFaces with nothing to add is a legal copy, not a no-op alias.
	if plain := base.WithFaces(); len(plain.Faces) != baseFaces {
		t.Errorf("WithFaces() with no faces gave %d faces, want %d", len(plain.Faces), baseFaces)
	}
}
