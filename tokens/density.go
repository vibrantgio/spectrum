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
const (
	// ComfortableControlHeight is the default desktop control height in dp.
	ComfortableControlHeight float32 = 36
	// CompactControlHeight is the dense-mode control height in dp.
	CompactControlHeight float32 = 28
	// MinHitTarget is the WCAG 2.5.5 pointer-target minimum in dp. It does
	// not scale with density; prism's current hardcoded 44 dp is this value,
	// not a control height.
	MinHitTarget float32 = 44
)
