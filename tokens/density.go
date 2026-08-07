package tokens

// Desktop density targets, measured 2026-08-05. This table is the
// justification for every number below it; later tasks (the Density token in
// E1.2, the component migrations in E1.3/E1.4) work from these values, so
// argue with the sources here rather than with the diffs there.
//
// Three-way control metrics — shadcn/ui vs MD3 vs macOS (AppKit):
//
//	metric                  shadcn/ui                    MD3                            macOS (AppKit)
//	------                  ---------                    ---                            --------------
//	button height, default  36 px (h-9)                  40 dp (filled button)          24 pt regular, 28 pt large
//	button height, small    32 px (h-8; xs is 24 px)     — (no smaller desktop size)    20 pt small, 16 pt mini
//	input height            36 px (h-9)                  56 dp (filled text field)      24 pt (rounded-bezel field)
//	base radius             10 px; controls 8 px (md)    pill (buttons), 4 dp (field)   not published
//	stacked spacing         8 px label→control,          8 dp grid                      8 pt system spacing
//	                        28 px between fields
//
// Sources (all fetched/measured 2026-08-05):
//
//   - shadcn/ui: button.tsx size variants `default: "h-9 …"`, `sm: "h-8 …"`,
//     `lg: "h-10 …"`, `xs: "h-6 …"`, base class `rounded-md`; input.tsx `"h-9 …
//     rounded-md …"`; form.tsx FormItem `"grid gap-2"` (8 px label→control);
//     field.tsx Field base `gap-3` (12 px) and FieldGroup `gap-7` (28 px
//     between stacked fields); globals.css `--radius: 0.625rem` (10 px) with
//     `--radius-md: calc(var(--radius) * 0.8)` = 8 px, the radius controls
//     actually render with. Tailwind: h-9 = 2.25rem = 36 px, h-8 = 32 px.
//     https://github.com/shadcn-ui/ui — apps/v4/registry/new-york-v4/ui/{button,input,form,field}.tsx
//     and apps/v4/app/globals.css; https://ui.shadcn.com/docs/theming.
//
//   - MD3: material-web design tokens v0.192 — md-comp-filled-button
//     'container-height': 40px, 'container-shape': corner-full;
//     md-comp-filled-text-field 'container-shape': corner-extra-small-top
//     (4 dp); md-sys-shape corner-extra-small 4 / small 8 / medium 12 /
//     large 16 / extra-large 28 px. Filled/outlined text field container
//     height is 56 dp per the m3.material.io text-field spec (the site is
//     JS-walled; the 40 dp button height cross-checks against Flutter's
//     generated token data, md.comp.filled-button.container.height = 40.0).
//     MD3's minimum touch target is 48 dp.
//     https://github.com/material-components/material-web — tokens/versions/v0_192/;
//     https://github.com/flutter/flutter — dev/tools/gen_defaults/data/button_filled.json;
//     https://m3.material.io/components/text-fields/specs.
//
//   - macOS: measured directly against AppKit on macOS (Darwin 25.5.0) via
//     fittingSize — NSButton (push bezel) mini 16 / small 20 / regular 24 /
//     large 28 pt; NSTextField (rounded bezel, regular) 24 pt; stacked-control
//     system spacing (constraint(equalToSystemSpacingBelow:multiplier:1) and
//     NSStackView default spacing) 8 pt. Note the plan's "28 pt standard
//     control" is NSControlSize.large — the size Apple uses for prominent
//     buttons since Big Sur — while regular measures 24 pt. Apple's HIG
//     publishes no per-size control heights for macOS, hence the direct
//     measurement.
//
// The picks:
//
//   - Comfortable = 36 dp. The shadcn/ui default (button and input alike),
//     sitting between macOS large (28 pt) and MD3's 40 dp — dense enough to
//     read as a desktop app, generous enough to remain the default.
//   - Compact = 28 dp. macOS's large control height and squarely between
//     shadcn's sm (32 px) and xs (24 px): a native-feeling dense mode that
//     stays above every AppKit regular-size control.
//
// Why prism's existing 44 dp was rejected as Comfortable: 44 comes from touch
// guidelines — the WCAG 2.5.5 pointer-target minimum, next to MD3's 48 dp
// touch target — and every desktop column above lands well below it (shadcn
// 36, macOS 24–28; even touch-first MD3 draws its button at 40 inside a 48 dp
// target). It is a hit-target floor, not a visual control height, and it stays
// a hit-target floor: E1.2 keeps the ≥44 dp pointer target independent of
// density, so Compact shrinks the drawn control but never the clickable area.
// A control height is a floor, not a height. This is the word the table above
// was missing, and F4.4 found it by measuring rather than reading: a Compact
// button draws 29 px against a CompactControlHeight of 28, and it does so with
// an empty label, so no amount of text is to blame. The arithmetic is simply
// that a control is as tall as its content box needs, and never shorter than
// the density says:
//
//	height = max(ControlHeight, contentHeight + 2×PaddingY)
//
// Where contentHeight is the type role's line height (see [TextStyle.LineHeight]
// and spectrum/typeset), the two terms are close enough that either can win:
//
//	control                role         line height   + 2×PaddingY   ControlHeight   drawn
//	-------                ----         -----------   ------------   -------------   -----
//	button, Comfortable    LabelLarge   20            36             36              36
//	button, Compact        LabelLarge   20            32             28              32
//	text field, Comfortable BodyLarge   24            40             36              40
//	text field, Compact    BodyLarge    24            36             28              36
//
// Comfortable's 36 dp is exactly LabelLarge's line box plus its own padding,
// which is not a coincidence — the number was picked against a button.
//
// # Compact's 28 dp is historical; the rule that made 36 would have said 32
//
// Compact's 28 dp is that same sum for a 16 dp line height — LabelMedium's,
// not LabelLarge's, and nothing in this system draws a button in LabelMedium.
// Read Comfortable's derivation as a rule ("the floor is the control's own line
// box plus its own padding") and apply it to Compact and the answer is not 28:
//
//	LabelLarge 20 + 2×6 = 32
//
// 32 is also shadcn/ui's sm button (h-8), so the evidence table above would
// have carried it without complaint. That is the figure the arithmetic wants,
// and it is written here so nobody has to re-derive it a third time. F4.4c
// found the discrepancy and documented it rather than changing the number;
// F5.6 re-opened the question and reached the same answer on purpose, for
// three reasons in descending weight:
//
//   - 28 dp does not rest on the LabelMedium arithmetic and never did. It came
//     from measurement — macOS's large control height, squarely between
//     shadcn's sm (32) and xs (24) — and the LabelMedium coincidence was
//     noticed afterwards, by F4.4. Correcting a derivation that was not the
//     source of the number corrects nothing.
//   - ControlHeight is a floor for controls but a *pin* for stacked rows.
//     prism/list rows, cadence's table rows and header cells and sidebar items
//     are ControlHeight tall exactly (see the row table below). Moving 28 to 32
//     is therefore not an arithmetic tidy-up; it is a visual change to every
//     dense list and table in the system.
//   - It would nearly erase Compact. A 32 dp Compact row against a 36 dp
//     Comfortable one is an 11% difference where today it is 22% — in exactly
//     the dense tables and lists Compact exists for.
//
// So the two densities are derived by *different rules*: Comfortable from a
// type role's line box, Compact from measured native control heights. The
// asymmetry is intended, and this paragraph exists to say so rather than let
// the next reader find 28 ≠ 32 and assume it is a typo. What is not allowed is
// calling the result a height and then measuring something else.
//
// The consequence worth saying out loud: controls in different type roles come
// out at different heights, and a Comfortable text field (40) is taller than a
// Comfortable button (36) because BodyLarge is a larger role than LabelLarge.
// Both are honest readings of the tokens. A design that wants them equal
// changes the roles or the padding, not the measurement.
//
// # Pointer targets: which WCAG level actually governs
//
// [MinHitTarget] is 44 dp, and 44 dp is not the AA requirement. Two success
// criteria are in play and they are a whole conformance level apart:
//
//	criterion                            level   threshold   applies to
//	---------                            -----   ---------   ----------
//	WCAG 2.5.5 Target Size (Enhanced)    AAA     44×44 CSS px  every pointer target
//	WCAG 2.5.8 Target Size (Minimum)     AA      24×24 CSS px  every pointer target
//
// (WCAG 2.2, https://www.w3.org/TR/WCAG22/#target-size-enhanced and
// #target-size-minimum. Both carry an inline/essential exception this system
// does not need to lean on.)
//
// E1.3 extended the pointer area to 44 dp for standalone controls — button,
// checkbox, radio, text field, the dropdown's closed trigger — and
// deliberately not for stacked rows: list rows, table rows and header cells,
// and the open dropdown's option rows. Adjacent rows tile edge to edge, so
// slop granted to one row is stolen from its neighbour; the extension would
// not enlarge anything, it would only make the boundary lie about where it is.
// Rows rely on their full row width instead.
//
// So the stacked-row targets are as tall as the row is, and F4.7 measured them
// at 1:1 rather than repeating the token (remember ControlHeight is a floor —
// max(ControlHeight, lineBox + 2×PaddingY) — so a row can draw more than the
// token says):
//
//	row                                     Comfortable   Compact   sizing
//	---                                     -----------   -------   ------
//	prism/list row (list.RowHeight)         36            28        pinned to ControlHeight
//	cadence/table body row and header cell  36            28        pinned to ControlHeight
//	cadence/sidebar item                    36            28        pinned to ControlHeight
//	prism/input dropdown option row         40            36        floor formula, BodyLarge
//
// The narrowest of these is the 28 dp Compact row. 28 ≥ 24, so every row in
// the system clears 2.5.8 at AA; none of them reaches 2.5.5's 44, and F4.7
// decided not to force them to. Flooring rows at 44 dp would erase Compact in
// exactly the dense tables and lists Compact exists for — a 44 dp "compact"
// row is 8 dp taller than a Comfortable one — which trades a real, everyday
// density benefit for a AAA criterion the system does not claim. An
// application that does target AAA sets Comfortable (36) and still does not
// reach 44 by row height alone; it needs a taller row of its own.
//
// This is a scoped promise, not a weakened one: what narrowed in F4.7 was the
// documentation, which had claimed 44 dp for everything. Nothing about the
// drawn or clickable geometry changed.
const (
	// ComfortableControlHeight is the default desktop control-height floor in
	// dp: a Comfortable control is at least this tall, and taller when its
	// content box needs it.
	ComfortableControlHeight float32 = 36
	// CompactControlHeight is the dense-mode control-height floor in dp. Every
	// control drawn in a Label or Body role clears it — see the table above —
	// so it is the floor that is least often the answer.
	//
	// The number is historical: it was measured off macOS's large control, and
	// the line-box rule that produced ComfortableControlHeight would have given
	// 32 here, not 28. Both facts are load-bearing and both are argued out
	// under "Compact's 28 dp is historical" above. Read that before changing
	// it — stacked rows are pinned to this value, not floored by it.
	CompactControlHeight float32 = 28
	// MinHitTarget is the pointer-target floor in dp for a *standalone*
	// control — one with space around it: button, checkbox, radio, text
	// field, the dropdown's closed trigger. Those extend their pointer area
	// to at least this on each axis, centred on the drawn control, whatever
	// the density.
	//
	// It is 44 dp, WCAG 2.5.5 Target Size (Enhanced), which is a AAA
	// criterion. It is not what stacked rows guarantee, and never was:
	// list rows, table rows and header cells, and open-dropdown option rows
	// are their own row height (28 dp at Compact) because extending one row
	// would steal its neighbour's slop. Those clear WCAG 2.5.8 Target Size
	// (Minimum) at 24 dp, the criterion that governs at AA. See "Pointer
	// targets: which WCAG level actually governs" above for the measured
	// per-row numbers.
	MinHitTarget float32 = 44
)

// Density is one density setting: the drawn control height and its inner
// padding, all in dp. It is a comparable value struct like the other token
// types. The standalone-control pointer-target floor is deliberately a method,
// not a field — see [Density.MinHitTarget] — so no Density value can carry a
// shrunken hit target.
type Density struct {
	// ControlHeight is the minimum visual control height in dp
	// ([ComfortableControlHeight] or [CompactControlHeight]). It is a floor:
	// a control draws max(ControlHeight, contentHeight + 2×PaddingY), so a
	// content box taller than this makes the control taller. See the table
	// above the constants for which controls clear it and by how much.
	ControlHeight float32
	// PaddingX is the horizontal inner padding of a control in dp.
	PaddingX float32
	// PaddingY is the vertical inner padding of a control in dp.
	PaddingY float32
}

// MinHitTarget returns the standalone-control pointer-target floor in dp —
// the package const [MinHitTarget], 44 dp, WCAG 2.5.5 Target Size (Enhanced).
// It is a method rather than a struct field, so it is structurally identical
// across every density: Compact shrinks the drawn control, never the clickable
// area of a control that has room to grow into.
//
// It does not describe stacked rows. Read [MinHitTarget] before wiring this
// into anything that tiles.
func (Density) MinHitTarget() float32 { return MinHitTarget }

// The padding picks come from the same measured world as the control
// heights above:
//
//   - Comfortable: shadcn/ui's default button is h-9 px-4 py-2 → 16 dp
//     horizontal, 8 dp vertical, pairing with the 36 dp height.
//   - Compact: shadcn's sm button drops to px-3 (12 dp); vertical scales
//     with it on the 2:1 ratio the default keeps → 12 dp / 6 dp, pairing
//     with the 28 dp height.
var (
	// Comfortable is the default desktop density.
	Comfortable = Density{ControlHeight: ComfortableControlHeight, PaddingX: 16, PaddingY: 8}
	// Compact is the dense mode: smaller drawn controls, same hit target.
	Compact = Density{ControlHeight: CompactControlHeight, PaddingX: 12, PaddingY: 6}
)
