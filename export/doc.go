// Package export serialises a theme.Theme emission into the project layout
// claude.ai/design consumes: theme.json, the machine-readable generative
// parameters; styles.css, the token sheet; readme.md, the project's front
// door; and the foundation pages under foundations/. Capture collects the
// first emission of each Theme observable into a Snapshot; Write renders
// the whole tree into a target directory. cmd/vg-tokens is the command-line
// front door.
//
// # The token sheet
//
// styles.css carries one :root block (the light scheme plus every
// mode-invariant scale, comfortable density), one .dark override block (the
// paired dark colours only) and one .compact override block (the compact
// density metrics only). ADR-007 records the reference project's token
// families but not its dark-mode selector, so the sheet uses class
// overrides — chosen here, not ADR-recorded — and .compact follows the same
// pattern; the two switches are orthogonal.
//
// Colour variables follow ADR-007's families exactly:
//
//   - --color-<role>-100 … --color-<role>-900 — the nine-step functional
//     ramps, roles neutral, primary, secondary, tertiary, error, success
//     and warning. The last three are the status roles: hue-fixed rather
//     than seed-derived, so a re-brand never rotates them.
//   - Pinned bases and the semantic layer: --color-accent is the Primary
//     pin (the reference project's .btn-primary consumes --color-accent, per
//     ADR-007), with --color-on-accent its on-colour; --color-secondary,
//     --color-tertiary, --color-error, --color-success and --color-warning
//     are the other role pins with their --color-on-* companions; --color-bg, --color-text are the pinned
//     background and body text; --color-surface and --color-divider are the
//     semantic layer's ramp-resolved card and separator colours.
//
// The remaining families, all emitted in :root only because they do not
// change with the scheme:
//
//   - --font-family and --font-family-code (the code style's mono family),
//     plus --font-<role>-size, -line-height, -weight and -tracking per type
//     role (display-large … body-small, and code — the mono style outside
//     the MD3 grid, at body-medium's metrics): sizes, line heights and
//     tracking in px, weights as CSS numeric weights.
//   - --density-control-height, --density-padding-x and --density-padding-y
//     from tokens.Density: :root carries tokens.Comfortable, the .compact
//     block overrides with tokens.Compact. --density-min-hit-target is the
//     WCAG 2.5.5 pointer-target floor, emitted once and never overridden —
//     density scales the drawn control, never the clickable area.
//   - --space-<key> from tokens.SpacingScale, keys as the Go scale names
//     them (0, 1, 2, … 24), in px.
//   - --radius-<key> from tokens.RadiusScale in Tailwind naming (none, sm,
//     base, md, lg, xl, 2xl, 3xl, full), in px; Base is also theme.json's
//     base radius parameter.
//   - --elevation-<level> (0–3): the tonal surface fills, the DEFAULT
//     elevation cue (E2.1). Each level is emitted as a var() reference —
//     var(--color-bg) for level 0's bg-pin sentinel, var(--color-neutral-N)
//     for the ramp steps — so the surfaces flip with .dark through the
//     colour overrides and the sheet itself states that an elevation level
//     is a neutral-ramp step.
//   - --shadow-<level> (0–3): CSS box-shadow approximations of the dp
//     depths, the OPT-IN cue floating transients (menus, dialogs, tooltips)
//     layer over their tonal fill (E2.2). Each level's dp depth d becomes
//     "0 <d>px <2d>px 0 rgba(0, 0, 0, 0.2)" — y-offset the depth, blur
//     twice it, no spread, black at 20% — and level 0 is "none".
//   - --ease-<name> from tokens.MotionScale: the six MD3 easing presets as
//     cubic-bezier() strings (standard and emphasized families, each plain
//     / -accelerate / -decelerate).
//   - --duration-<stop> (x-fast, fast, normal, slow, x-slow): the five
//     MD3-pinned duration stops in ms. The spring presets have no CSS
//     counterpart — springs are Go-side physics — and are serialised only
//     in theme.json's motion parameters.
//
// # The foundation pages
//
// foundations/color.html, type.html and layout.html render the scales at
// real sizes, and readme.md is the file a human or an agent reads first.
// The pages are static HTML that reads only from the emitted sheet: every
// styled colour, size, radius, shadow and font value is a var() reference
// into ../styles.css, so regenerating the sheet from another seed reflows
// every page. Literal token values appear only as annotation text — hexes,
// px numbers, and the measured APCA Lc and WCAG 2 ratio of each text pair —
// printed for both modes (labelled L and D) because text cannot flip with a
// class the way painted specimens do. Each page carries a light/dark toggle
// flipping the .dark class on the root element. The page test enforces the
// no-hard-coded-values rule: no hex colours or px lengths in any style
// context, and every referenced variable declared by the sheet.
//
// # The generative parameters
//
// theme.json records what reproduces the theme: the seed (hex plus its
// OKLCh hue and sat), the pinned role hexes per mode, the heading, body and
// mono faces, the base radius, the shared CIELAB L* scales measured back from
// the emitted neutral ramps, the density model (the active setting by name,
// both settings' metrics and the invariant hit-target floor), the elevation
// model (surface step and shadow dp per level) and the motion set
// (durations in ms, easing control points, spring presets).
// tokens.FromSeed(seed) regenerates the full palette from the seed alone —
// the round-trip test asserts it. The per-role --font-*-size tokens come from
// Typography, which is the theme's only type source since v0.3.0 dropped the
// size-only Theme.Type stream.
package export
