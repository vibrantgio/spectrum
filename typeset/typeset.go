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
// number is used verbatim rather than scaled by the face's own metrics. A zero
// line height leaves both fields unset, which keeps the shaper's default.
//
// Set any other field — Alignment, WrapPolicy, Truncator — on the result
// before handing it to [Layout].
func Label(style tokens.TextStyle, maxLines int) widget.Label {
	lbl := widget.Label{MaxLines: maxLines}
	if style.LineHeight != 0 {
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
// both. An absolute line height smaller than the face's natural line — an
// unset one included — has no leading to distribute. And a label whose
// LineHeightScale is not 1 is asking for a height relative to the face's
// metrics, which the shaper already applies to every line including the first,
// so there is nothing missing to add.
//
// The extra height is a single deficit, added once, not once per line. Gio
// already spends the line height on the gap between lines, so the only line
// short of its box is the first: adding lineHeight − naturalLine to a run of n
// lines makes it exactly n × lineHeight tall. The half above is rounded down,
// which is what keeps a centred label pixel-identical to the uncorrected one
// whenever its container was already taller than the ink.
func Layout(gtx layout.Context, sh *text.Shaper, lbl widget.Label, f font.Font, size unit.Sp, txt string, material op.CallOp) layout.Dimensions {
	box := gtx.Sp(lbl.LineHeight)
	if box <= 0 || lbl.LineHeightScale != 1 {
		return lbl.Layout(gtx, sh, f, size, txt, material)
	}

	deficit := box - naturalLine(gtx, sh, f, size, material)
	if deficit <= 0 {
		return lbl.Layout(gtx, sh, f, size, txt, material)
	}

	rec := op.Record(gtx.Ops)
	dims := lbl.Layout(gtx, sh, f, size, txt, material)
	call := rec.Stop()

	above := deficit / 2
	off := op.Offset(image.Pt(0, above)).Push(gtx.Ops)
	call.Add(gtx.Ops)
	off.Pop()

	dims.Size.Y += deficit
	// Baseline is measured up from the bottom of the dimensions, so it grows
	// by the half added below the ink, not by the whole deficit.
	dims.Baseline += deficit - above
	return dims
}

// naturalLine measures the height one line of f at size occupies without any
// absolute line height — the ascent-plus-descent box Gio gives the first line.
// It shapes the empty string, which has that height and no glyphs, under
// relaxed constraints so a tight caller cannot clamp the answer, and discards
// the recorded ops.
func naturalLine(gtx layout.Context, sh *text.Shaper, f font.Font, size unit.Sp, material op.CallOp) int {
	probeGtx := gtx
	probeGtx.Constraints = layout.Constraints{Max: image.Pt(1<<20, 1<<20)}

	probe := widget.Label{MaxLines: 1}
	rec := op.Record(gtx.Ops)
	dims := probe.Layout(probeGtx, sh, f, size, "", material)
	rec.Stop()
	return dims.Size.Y
}
