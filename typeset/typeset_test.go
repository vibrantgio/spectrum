package typeset_test

import (
	"image"
	"testing"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/vibrantgio/font/notosansmono"
	"github.com/vibrantgio/spectrum/tokens"
	"github.com/vibrantgio/spectrum/typeset"
)

// specimen carries an ascender, a descender and digits, so the ink box is the
// widest a Latin run gets and the correction cannot be an artefact of a string
// that happens to sit inside x-height.
const specimen = "Il1 Wm gj 018"

// arrowed is the specimen plus two runes no Roboto face carries, so the line
// it shapes to holds a run from a second face. That is the case the natural
// line has to be measured from the text for: Gio takes a line's ascent as the
// maximum over its runs, so this line is taller than the primary face alone.
const arrowed = "Il1 Wm gj 018 →←"

// pinned is the shaper every test here draws with: the default faces, system
// fonts off, so a machine with a different font set cannot change a
// measurement.
func pinned() *text.Shaper { return tokens.DefaultTypography.DeterministicShaper() }

// fallbackShaper is pinned's counterpart for the mixed-face measurements: the
// default faces plus the optional symbol face, still with system fonts off, so
// an arrow resolves to a real face on every machine rather than to whatever
// the platform happens to own. It stands in for the fallback that applications
// get from the system.
//
// It returns the style to draw with as well, because that style must ask for a
// weight the fallback face has. Gio matches a fallback by weight and Noto Sans
// Mono ships Regular alone, so LabelLarge's Medium does not reach it: at
// Medium the arrow shapes to Roboto's missing-glyph box and the line stays the
// primary face's height, which would make these tests measure nothing.
func fallbackShaper() (*text.Shaper, tokens.TextStyle) {
	typ := tokens.DefaultTypography.WithFaces(notosansmono.FontFace())
	style := typ.LabelLarge
	style.Weight = 0
	return typ.DeterministicShaper(), style
}

// gtx returns a layout context at 1 px per dp and per sp, so every measured px
// below is also the dp figure a role names.
func gtx(ops *op.Ops, maxX int) layout.Context {
	return layout.Context{
		Ops:         ops,
		Metric:      unit.Metric{PxPerDp: 1, PxPerSp: 1},
		Constraints: layout.Constraints{Max: image.Pt(maxX, 1<<20)},
	}
}

// measure lays txt out in style on one line and returns the box height.
func measure(t *testing.T, g layout.Context, sh *text.Shaper, style tokens.TextStyle, txt string) int {
	t.Helper()
	return typeset.Layout(g, sh, typeset.Label(style, 1), typeset.Font(style, font.Normal),
		unit.Sp(style.Size), txt, op.CallOp{}).Size.Y
}

func styleAt(lineHeight float32) tokens.TextStyle {
	s := tokens.DefaultTypography.LabelLarge
	s.LineHeight = lineHeight
	return s
}

// TestLineHeightChangesSingleLineHeight is the assertion the organization did
// not have. tokens.TextStyle.LineHeight is documented to reach the shaper, and
// it does — but through widget.Label alone it cannot move a single-line label
// by one pixel, because gioui.org/text spends the line height on the gap to
// the next line and a MaxLines:1 label has no next line. Measured before
// typeset existed: 17 px at line height 0, 20, 32 and 64 alike.
//
// This test fails against that behaviour, which is the point of it. If someone
// unwires typeset.Layout back to widget.Label.Layout, every height below
// collapses to the same number and this test says so before a golden does.
func TestLineHeightChangesSingleLineHeight(t *testing.T) {
	var ops op.Ops
	g := gtx(&ops, 1<<20)
	sh := pinned()

	heights := map[float32]int{}
	for _, lh := range []float32{20, 32, 64} {
		style := styleAt(lh)
		dims := typeset.Layout(g, sh, typeset.Label(style, 1), typeset.Font(style, font.Normal),
			unit.Sp(style.Size), specimen, op.CallOp{})
		if got, want := dims.Size.Y, int(lh); got != want {
			t.Errorf("line height %v: single-line box %d px, want %d", lh, got, want)
		}
		heights[lh] = dims.Size.Y
	}
	if heights[20] == heights[32] {
		t.Errorf("line heights 20 and 32 both lay out %d px tall: the role's line height reaches the shaper and changes nothing", heights[20])
	}
}

// TestLineHeightBelowNaturalLineIsFloored pins the one case where two line
// heights legitimately agree, with the reason beside it: a line height smaller
// than the face's own ascent-plus-descent has no leading to distribute, and
// typeset does not shrink a line box below its glyphs. LabelLarge at 14 dp
// inks 17 px, so 8 and 12 both come back as 17 — not because the property was
// dropped, but because there is nothing to add.
func TestLineHeightBelowNaturalLineIsFloored(t *testing.T) {
	var ops op.Ops
	g := gtx(&ops, 1<<20)
	sh := pinned()

	natural := measure(t, g, sh, styleAt(0), specimen)
	for _, lh := range []float32{8, 12} {
		if got := measure(t, g, sh, styleAt(lh), specimen); got != natural {
			t.Errorf("line height %v: box %d px, want the natural line %d px", lh, got, natural)
		}
	}
	if natural >= 20 {
		t.Fatalf("natural LabelLarge line is %d px; this test assumed it below 20", natural)
	}
}

// TestWrappedLinesGetWholeLineBoxes checks the correction is a deficit added
// once rather than a floor applied per line: wrapped text must come back at a
// whole multiple of the line height, which is what a CSS engine gives the same
// text and what Gio alone does not — it leaves the first line short by
// lineHeight − naturalLine and so lands between multiples.
func TestWrappedLinesGetWholeLineBoxes(t *testing.T) {
	var ops op.Ops
	g := gtx(&ops, 60) // narrow enough to wrap the specimen
	sh := pinned()

	const lh = 20
	style := styleAt(lh)
	for _, maxLines := range []int{2, 3, 4} {
		lbl := typeset.Label(style, maxLines)
		dims := typeset.Layout(g, sh, lbl, typeset.Font(style, font.Normal),
			unit.Sp(style.Size), specimen, op.CallOp{})
		if dims.Size.Y < 2*lh {
			t.Fatalf("MaxLines %d: %d px — the specimen did not wrap, so this proves nothing", maxLines, dims.Size.Y)
		}
		if dims.Size.Y%lh != 0 {
			t.Errorf("MaxLines %d: %d px is not a whole multiple of the %d px line height", maxLines, dims.Size.Y, lh)
		}
	}
}

// TestUncorrectedLabelStillReportsInk records the behaviour typeset wraps, so
// that the reason this package exists stays measured rather than remembered.
// If a future gioui.org gives the first line its whole line box, this test is
// the one that fails, and typeset.Layout's deficit becomes zero on its own.
func TestUncorrectedLabelStillReportsInk(t *testing.T) {
	var ops op.Ops
	g := gtx(&ops, 1<<20)
	sh := pinned()

	var seen []int
	for _, lh := range []float32{0, 20, 32, 64} {
		style := styleAt(lh)
		lbl := typeset.Label(style, 1)
		dims := lbl.Layout(g, sh, typeset.Font(style, font.Normal), unit.Sp(style.Size), specimen, op.CallOp{})
		seen = append(seen, dims.Size.Y)
	}
	for i, got := range seen {
		if got != seen[0] {
			t.Fatalf("gioui.org/widget.Label now honours line height on a single line: %v; typeset.Layout's deficit is stale (case %d)", seen, i)
		}
	}
}

// TestLabelKeepsCallerFields checks typeset.Label only fills in the line
// height: everything else a caller sets on the returned value must survive,
// since the call sites set Alignment and MaxLines on it.
func TestLabelKeepsCallerFields(t *testing.T) {
	style := styleAt(20)
	lbl := typeset.Label(style, 3)
	lbl.Alignment = text.Middle
	if lbl.MaxLines != 3 || lbl.Alignment != text.Middle {
		t.Errorf("Label(…, 3) = %+v, want MaxLines 3 and the caller's alignment", lbl)
	}
	if lbl.LineHeight != unit.Sp(style.LineHeight) || lbl.LineHeightScale != 1 {
		t.Errorf("Label installed line height %v scale %v, want %v scale 1", lbl.LineHeight, lbl.LineHeightScale, style.LineHeight)
	}
	if zero := typeset.Label(styleAt(0), 1); zero.LineHeight != 0 || zero.LineHeightScale != 0 {
		t.Errorf("a zero line height installed %v scale %v, want both unset", zero.LineHeight, zero.LineHeightScale)
	}
}

// TestFontFallbackWeight pins Font's two-way weight rule: the style wins when
// it names a weight, the draw site's fallback wins when it does not.
func TestFontFallbackWeight(t *testing.T) {
	styled := typeset.Font(tokens.DefaultTypography.LabelLarge, font.Normal)
	if want := tokens.FontWeight(tokens.WeightMedium); styled.Weight != want {
		t.Errorf("LabelLarge weight = %v, want %v", styled.Weight, want)
	}
	bare := typeset.Font(tokens.TextStyle{Typeface: "Roboto"}, font.Bold)
	if bare.Weight != font.Bold {
		t.Errorf("unset weight = %v, want the fallback font.Bold", bare.Weight)
	}
	if bare.Typeface != "Roboto" {
		t.Errorf("typeface = %q, want Roboto", bare.Typeface)
	}
}

// TestBaselineMovesWithTheLowerHalf checks the baseline is reported against
// the new bottom edge. Dimensions.Baseline is measured up from the bottom, so
// it must grow by the leading added below the ink and not by the whole
// deficit — otherwise every layout.Flex aligned on Baseline would drift by
// half a line.
func TestBaselineMovesWithTheLowerHalf(t *testing.T) {
	var ops op.Ops
	g := gtx(&ops, 1<<20)
	sh := pinned()

	style := styleAt(0)
	bare := widget.Label{MaxLines: 1}
	plain := bare.Layout(g, sh, typeset.Font(style, font.Normal), unit.Sp(style.Size), specimen, op.CallOp{})

	boxed := styleAt(32)
	dims := typeset.Layout(g, sh, typeset.Label(boxed, 1), typeset.Font(boxed, font.Normal),
		unit.Sp(boxed.Size), specimen, op.CallOp{})

	deficit := dims.Size.Y - plain.Size.Y
	if want := plain.Baseline + deficit - deficit/2; dims.Baseline != want {
		t.Errorf("baseline = %d, want %d (plain %d + %d below the ink)",
			dims.Baseline, want, plain.Baseline, deficit-deficit/2)
	}
}

// TestMixedFaceLineLandsOnItsLineHeight is the assertion the empty-string probe
// could not make. The probe measured the primary face's ascent and descent, but
// Gio takes a line's ascent as the maximum over that line's runs, so a line
// carrying a fallback run is taller than the probe and the deficit was computed
// against the wrong baseline. Measured before this was fixed: 23 px on a role
// that declares 20 — and spectrum/export writes `line-height: 20` for the same
// role, so the two surfaces disagreed for exactly the characters the fallback
// exists to serve.
func TestMixedFaceLineLandsOnItsLineHeight(t *testing.T) {
	var ops op.Ops
	g := gtx(&ops, 1<<20)
	sh, style := fallbackShaper()

	// The premise first: without this, the test could pass because the arrow
	// never reached the second face and every line was the primary's height.
	f := typeset.Font(style, font.Normal)
	bare := widget.Label{MaxLines: 1}
	latin := bare.Layout(g, sh, f, unit.Sp(style.Size), specimen, op.CallOp{}).Size.Y
	mixed := bare.Layout(g, sh, f, unit.Sp(style.Size), arrowed, op.CallOp{}).Size.Y
	if mixed <= latin {
		t.Fatalf("uncorrected: Latin %d px, mixed %d px — the arrows did not pull in a taller face, so this test measures nothing", latin, mixed)
	}

	for _, txt := range []string{specimen, arrowed} {
		if got, want := measure(t, g, sh, style, txt), int(style.LineHeight); got != want {
			t.Errorf("%q: box %d px, want the role's %d px line height", txt, got, want)
		}
	}
}

// TestMixedFaceWrappedLinesGetWholeLineBoxes is TestWrappedLinesGetWholeLineBoxes
// under the fallback, where the deficit measured off a probe went wrong by the
// difference between two faces: the wrapped mixed-face run came back 41 px on a
// 20 px line height, which is not a multiple of anything.
func TestMixedFaceWrappedLinesGetWholeLineBoxes(t *testing.T) {
	var ops op.Ops
	g := gtx(&ops, 60) // narrow enough to wrap
	sh, style := fallbackShaper()
	lh := int(style.LineHeight)

	for _, maxLines := range []int{2, 3, 4} {
		lbl := typeset.Label(style, maxLines)
		dims := typeset.Layout(g, sh, lbl, typeset.Font(style, font.Normal),
			unit.Sp(style.Size), arrowed, op.CallOp{})
		if dims.Size.Y < 2*lh {
			t.Fatalf("MaxLines %d: %d px — the text did not wrap, so this proves nothing", maxLines, dims.Size.Y)
		}
		if dims.Size.Y%lh != 0 {
			t.Errorf("MaxLines %d: %d px is not a whole multiple of the %d px line height", maxLines, dims.Size.Y, lh)
		}
	}
}

// TestResultFitsTheCallersConstraints pins the resolution of the constraint
// double-count. widget.Label constrains its own result, so adding the deficit
// on top of that reported more than the caller's slot whenever Min.Y was set —
// which is every Flexed child of a vertical layout.Flex. The org's components
// dodged it by zeroing Constraints.Min first, which was convention and not
// contract. Layout now constrains the corrected size instead, so the contract
// is the one every other Gio widget keeps.
func TestResultFitsTheCallersConstraints(t *testing.T) {
	var ops op.Ops
	sh := pinned()
	style := styleAt(20)

	for _, slot := range []int{12, 20, 40} {
		g := gtx(&ops, 1<<20)
		g.Constraints.Min.Y, g.Constraints.Max.Y = slot, slot
		dims := typeset.Layout(g, sh, typeset.Label(style, 1), typeset.Font(style, font.Normal),
			unit.Sp(style.Size), specimen, op.CallOp{})
		if dims.Size.Y != slot {
			t.Errorf("Min.Y == Max.Y == %d: reported %d px", slot, dims.Size.Y)
		}
	}
}

// TestNegativeLineHeightNeverReachesTheShaper covers the mismatch between the
// two guards. widget.Label installs any LineHeight that is != 0 where Layout
// bails at <= 0, so a negative one used to pass straight through with
// LineHeightScale 1 — and gioui.org/text then baselines each line above the one
// before it, stacking a wrapped label's lines on top of each other and
// reporting the height of one.
func TestNegativeLineHeightNeverReachesTheShaper(t *testing.T) {
	if lbl := typeset.Label(styleAt(-20), 1); lbl.LineHeight != 0 || lbl.LineHeightScale != 0 {
		t.Errorf("Label at line height -20 installed %v scale %v, want both unset", lbl.LineHeight, lbl.LineHeightScale)
	}

	var ops op.Ops
	g := gtx(&ops, 60) // narrow enough to wrap
	sh := pinned()
	style := styleAt(0)
	f := typeset.Font(style, font.Normal)

	one := widget.Label{MaxLines: 1}.Layout(g, sh, f, unit.Sp(style.Size), specimen, op.CallOp{}).Size.Y
	hand := widget.Label{LineHeight: -20, LineHeightScale: 1}
	if got := typeset.Layout(g, sh, hand, f, unit.Sp(style.Size), specimen, op.CallOp{}).Size.Y; got <= one {
		t.Errorf("a hand-built negative line height laid %d px out where one line alone is %d: the wrapped lines collapsed onto each other", got, one)
	}
}
