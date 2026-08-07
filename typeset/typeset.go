// Package typeset lays a type role's text out in the line box that role names,
// rather than in the box its glyphs happen to ink.
//
// It exists because gioui.org's text layout and a design system's typography
// mean different things by "line height". [gioui.org/widget.Label] passes
// LineHeight to the shaper, and gioui.org/text's calculateYOffsets baselines
// the first line at that line's own ascent and spends the line height only on
// the gap to the next one. The consequence is exact and easy to miss: a label
// with MaxLines 1 — which nearly every control in this system is — reports the
// same size at any line height at all. Measured on prism/button's LabelLarge
// specimen at 14 dp: 17 px tall at line height 0, 20, 32 and 64 alike, and the
// rendered button byte-identical in all four.
//
// A design system means the CSS thing. `line-height: 20px` on a one-line
// button makes the line box 20 px tall whatever the glyphs measure, the extra
// space split half above and half below the ink, and that is what
// spectrum/export already writes into `--font-<role>-line-height` for the
// design-surface mirror to consume. Without this package the Gio rendering and
// the CSS it exports disagree about the same token.
//
// [Layout] is the fix, and it is a wrapper rather than a replacement: it lays
// the label out exactly as widget.Label would, then pads the result up to the
// line box and reports that. Callers keep MaxLines, Alignment, WrapPolicy and
// every other widget.Label field.
//
//	f := typeset.Font(style, font.Normal)
//	lbl := typeset.Label(style, 1)
//	dims := typeset.Layout(gtx, shaper, lbl, f, unit.Sp(style.Size), text, material)
//
// The correction is a deficit, not a floor, so it is right for wrapped text
// too: Gio already spends the line height on each gap, so adding the one
// missing line height gives n lines a box of exactly n × line height.
//
// The deficit is measured against the text being laid out, not against the
// face it names, because Gio takes a line's ascent as the maximum over that
// line's runs — a line holding a fallback run is taller than its primary face
// and needs less added, not the same. See [Layout].
package typeset

import (
	"image"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/text"
	"gioui.org/unit"
	"gioui.org/widget"

	"github.com/vibrantgio/spectrum/tokens"
)

// Font builds the font.Font a text style shapes with. The style's typeface is
// honoured always; its weight is honoured when non-zero, and a zero weight —
// which is what an unset [tokens.TextStyle] carries — falls back to fallback.
// Pass font.Normal as the fallback unless the draw site has a weight of its
// own to keep.
func Font(style tokens.TextStyle, fallback font.Weight) font.Font {
	f := font.Font{Typeface: font.Typeface(style.Typeface), Weight: fallback}
	if style.Weight != 0 {
		f.Weight = tokens.FontWeight(style.Weight)
	}
	return f
}

// Label builds the widget.Label for a style at maxLines, with the style's line
// height installed as an absolute value (LineHeightScale 1) so the role's
// number is used verbatim rather than scaled by the face's own metrics. A line
// height that is not positive leaves both fields unset, which keeps the
// shaper's default.
//
// The test is `> 0` rather than `!= 0` deliberately. [widget.Label] installs
// its LineHeight whenever it is non-zero and gioui.org/text then takes it as
// the whole line box, so a negative one baselines every line *above* the line
// before it and a wrapped label draws its lines on top of each other. There is
// no reading of a negative role line height that is better than none.
//
// Set any other field — Alignment, WrapPolicy, Truncator — on the result
// before handing it to [Layout].
func Label(style tokens.TextStyle, maxLines int) widget.Label {
	lbl := widget.Label{MaxLines: maxLines}
	if style.LineHeight > 0 {
		lbl.LineHeight = unit.Sp(style.LineHeight)
		lbl.LineHeightScale = 1
	}
	return lbl
}

// Layout lays txt out as lbl would and returns it in its line box: the same
// pixels, in dimensions tall enough for the line height lbl carries, with the
// leading split evenly above and below the ink and the baseline moved to
// match.
//
// It is a no-op in two cases, and returns lbl.Layout's own result unchanged in
// both. An absolute line height smaller than the natural line of this text —
// an unset or negative one included — has no leading to distribute. And a
// label whose LineHeightScale is not 1 is asking for a height relative to the
// face's metrics, which the shaper already applies to every line including the
// first, so there is nothing missing to add.
//
// The extra height is a single deficit, added once, not once per line. Gio
// already spends the line height on the gap between lines, so the only line
// short of its box is the first: adding lineHeight − naturalLine to a run of n
// lines makes it exactly n × lineHeight tall. The half above is rounded down,
// which is what keeps a centred label pixel-identical to the uncorrected one
// whenever its container was already taller than the ink.
//
// # The natural line is this text's, not this face's
//
// naturalLine is measured from txt itself. Gio takes a line's ascent as the
// maximum over that line's runs, so a line carrying a fallback run — an arrow,
// a box-drawing character, anything the primary face has no glyph for — is
// taller than the primary face alone. Measuring a probe string instead made
// the deficit too large for exactly those lines: under the fallback shaper
// applications draw with, "arrows →←" came back 25 px tall where LabelLarge
// declares 20 and spectrum/export writes `line-height: 20`.
//
// # Constraints are applied once, to the corrected height
//
// widget.Label constrains its own result, so adding the deficit on top of that
// would double-count: a label handed Min.Y == Max.Y — which is every Flexed
// child of a vertical layout.Flex — would report more than its slot. Layout
// therefore lays out under a relaxed Min.Y, corrects, and constrains the
// corrected size once with the caller's own constraints. The result fits the
// constraints it was given, which is what every other Gio widget promises, and
// callers no longer have to zero Constraints.Min to be told the truth.
func Layout(gtx layout.Context, sh *text.Shaper, lbl widget.Label, f font.Font, size unit.Sp, txt string, material op.CallOp) layout.Dimensions {
	if gtx.Sp(lbl.LineHeight) < 0 {
		// widget.Label installs any non-zero LineHeight; this function bails
		// at <= 0. Without this the two disagree and a negative reaches the
		// shaper with scale 1, which baselines each line above the last.
		lbl.LineHeight = 0
	}
	box := gtx.Sp(lbl.LineHeight)
	if box <= 0 || lbl.LineHeightScale != 1 {
		return lbl.Layout(gtx, sh, f, size, txt, material)
	}

	deficit := box - naturalLine(gtx, sh, lbl, f, size, txt, material, box)
	if deficit <= 0 {
		return lbl.Layout(gtx, sh, f, size, txt, material)
	}

	// Min.Y is dropped for the inner layout so widget.Label reports the ink it
	// actually drew rather than the caller's floor; the floor is re-applied to
	// the corrected size below.
	inner := gtx
	inner.Constraints.Min.Y = 0

	rec := op.Record(gtx.Ops)
	dims := lbl.Layout(inner, sh, f, size, txt, material)
	call := rec.Stop()

	above := deficit / 2
	off := op.Offset(image.Pt(0, above)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	off.Pop()

	dims.Size.Y += deficit
	// Baseline is measured up from the bottom of the dimensions, so it grows
	// by the half added below the ink, not by the whole deficit.
	dims.Baseline += deficit - above

	// The ink is anchored to the top of the box, so whichever way the caller's
	// constraints move the bottom edge, the baseline moves with it.
	boxed := dims.Size.Y
	dims.Size = gtx.Constraints.Constrain(dims.Size)
	dims.Baseline += dims.Size.Y - boxed
	return dims
}

// naturalLine measures the ascent-plus-descent box Gio gives txt's own lines
// at box — the height a run of them has before the missing first line box is
// added — rather than the box it would give a probe string in f alone.
//
// It reads the number off two measurements of txt, because Gio spends an
// absolute line height only on the gaps between lines: laying n lines out at L
// gives naturalLine + (n−1)×L. Laying the same text out at a second, larger L′
// therefore moves the height by (n−1)×(L′−L) and nothing else — wrapping and
// truncation do not depend on the line height — so the gap count divides out
// and naturalLine falls out of the first measurement.
//
// The first measurement shapes the same text.Parameters the caller's own
// layout will, so it is a hit in the shaper's cache rather than a second
// shaping. The second is skipped whenever the text is known to be one line,
// which is the common case in this system: MaxLines 1, or a measurement that
// already fits inside a single box.
//
// Both measurements run with Min.Y dropped and Max.Y opened, so neither the
// caller's floor nor its viewport — which clips the glyphs widget.Label
// measures — can distort the answer. Max.X is left alone: it decides where the
// text wraps, and a measurement that wrapped differently would measure
// different text. The recorded ops are discarded.
func naturalLine(gtx layout.Context, sh *text.Shaper, lbl widget.Label, f font.Font, size unit.Sp, txt string, material op.CallOp, box int) int {
	m := gtx
	m.Constraints.Min.Y = 0
	m.Constraints.Max.Y = 1 << 20

	h := measureHeight(m, sh, lbl, f, size, txt, material)
	if lbl.MaxLines == 1 || h <= box {
		// One line, so there are no gaps to subtract: h is the natural line.
		// A height within a single box cannot hold two, since a second line
		// would add a whole box to it.
		return h
	}

	wide := lbl
	wide.LineHeight = lbl.LineHeight * 2
	wideBox := gtx.Sp(wide.LineHeight)
	if wideBox <= box {
		return h
	}
	gaps := (measureHeight(m, sh, wide, f, size, txt, material) - h) / (wideBox - box)
	if gaps <= 0 {
		return h
	}
	if natural := h - gaps*box; natural > 0 {
		return natural
	}
	return h
}

// measureHeight lays lbl out for its height alone and throws the drawing away.
func measureHeight(gtx layout.Context, sh *text.Shaper, lbl widget.Label, f font.Font, size unit.Sp, txt string, material op.CallOp) int {
	rec := op.Record(gtx.Ops)
	dims := lbl.Layout(gtx, sh, f, size, txt, material)
	rec.Stop()
	return dims.Size.Y
}
