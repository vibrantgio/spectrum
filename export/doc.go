// Package export serialises a theme.Theme emission into the project layout
// claude.ai/design consumes: theme.json, the machine-readable generative
// parameters, and styles.css, the token sheet. Capture collects the first
// emission of each Theme observable into a Snapshot; Write renders the pair
// into a target directory. cmd/vg-tokens is the command-line front door.
//
// # The token sheet
//
// styles.css carries one :root block (the light scheme plus every
// mode-invariant scale) and one .dark override block (the paired dark
// colours only). ADR-007 records the reference project's token families but
// not its dark-mode selector, so the sheet uses a .dark class override —
// chosen here, not ADR-recorded.
//
// Colour variables follow ADR-007's families exactly:
//
//   - --color-<role>-100 … --color-<role>-900 — the nine-step functional
//     ramps, roles neutral, primary, secondary, tertiary and error.
//   - Pinned bases and the semantic layer: --color-accent is the Primary
//     pin (the reference project's .btn-primary consumes --color-accent, per
//     ADR-007), with --color-on-accent its on-colour; --color-secondary,
//     --color-tertiary and --color-error are the other role pins with their
//     --color-on-* companions; --color-bg, --color-text are the pinned
//     background and body text; --color-surface and --color-divider are the
//     semantic layer's ramp-resolved card and separator colours.
//
// The remaining families, all emitted in :root only because they do not
// change with the scheme:
//
//   - --font-family, plus --font-<role>-size, -line-height, -weight and
//     -tracking per type role (display-large … body-small): sizes, line
//     heights and tracking in px, weights as CSS numeric weights.
//   - --space-<key> from tokens.SpacingScale, keys as the Go scale names
//     them (0, 1, 2, … 24), in px.
//   - --radius-<key> from tokens.RadiusScale in Tailwind naming (none, sm,
//     base, md, lg, xl, 2xl, 3xl, full), in px; Base is also theme.json's
//     base radius parameter.
//   - --shadow-<level> (0–5): CSS box-shadow approximations of
//     tokens.ElevationScale. Each level's dp depth d becomes
//     "0 <d>px <2d>px 0 rgba(0, 0, 0, 0.2)" — y-offset the depth, blur
//     twice it, no spread, black at 20% — and level 0 is "none". E2.1
//     remaps elevation to surface roles and E5.1 re-emits it.
//
// # The generative parameters
//
// theme.json records what reproduces the theme: the seed (hex plus its
// OKLCh hue and sat), the pinned role hexes per mode, the heading and body
// faces, the base radius, and the shared CIELAB L* scales measured back
// from the emitted neutral ramps. tokens.FromSeed(seed) regenerates the
// full palette from the seed alone — the round-trip test asserts it.
// Density and the motion set are E5.1's; they are deliberately absent, as
// is Theme.Motion from the sheet. Theme.Type is not consumed either: the
// per-role --font-*-size tokens come from Typography, which carries the
// same sizes.
package export
