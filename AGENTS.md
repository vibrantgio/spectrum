# AGENTS.md — spectrum

The theme runtime of the VibrantGio design system: `system.LiveTheme`,
which polls the OS light/dark preference and publishes it as an observable
theme; `window`, which pairs that observable with an mvu window;
`preferences`, which persists the chosen theme name; and `transition`,
which interpolates theme tokens between the two.

**Layer.** Tier 1 of ADR-001's stack, `mvu → spectrum → prism → pulse →
cadence → markdown`. It imports mvu, and today also `prism/theme`,
`prism/tokens`, `prism/a11y` and `pulse/tween` — the inversion goal G-B3
corrects by moving the token and theme contract down out of prism into
spectrum, and `spectrum/transition` up into `pulse/transition`. Nothing
else in the design system imports spectrum; the workbench applications do.

**Read the canonical guide before you write code against this module.** It is
the organization's one agent guide — the module inventory with current tags,
the application skeleton, the MVU loop and rx semantics, typography, and the
pitfalls that are not guessable. It lives exactly once, in `vibrantgio/.github`,
and this file links it rather than copying it:

    https://raw.githubusercontent.com/vibrantgio/.github/master/llms.txt

**Module.** `github.com/vibrantgio/spectrum`, one module at the repository
root.

**Build and test.** From the repository root:

    go build ./... && go test ./...
