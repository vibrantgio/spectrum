# AGENTS.md — spectrum

The theme runtime of the Vibrant Gio design system: `system.LiveTheme`,
which polls the OS light/dark preference and publishes it as an observable
theme; `window`, which pairs that observable with an mvu window; and
`preferences`, which persists the chosen theme name. Interpolating theme
tokens between the two lives above, in `pulse/transition`.

**Layer.** Tier 1 of ADR-001's stack, `mvu → spectrum → prism → pulse →
cadence → markdown`. It imports mvu; the G-B3 inversion moved the token and
theme contract down out of prism into spectrum, E3.2 moved `a11y` down the
same way, and `spectrum/transition` moved up into `pulse/transition`. F3.3
deleted that last shim, so spectrum has no upward edge at all: check-layers
records zero transitional edges for it. The design-system layers above
(prism, pulse, cadence, markdown) and the workbench applications import
spectrum.

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

**Two shapers, and the choice is yours to make.** `tokens.Typography` builds
both, cached in separate fields so neither can hand back the other's:

- `Shaper()` — the system fallback is **on**. This is what applications and
  library components draw with. The embedded faces answer first; the platform's
  fonts answer for everything they lack, which is every arrow, box-drawing
  character and dingbat, because Roboto and Roboto Mono carry none of them.
  Never disable system fonts here to make an output stable — that is the F4.2
  defect, and it ships tofu to every user.
- `DeterministicShaper()` — system fonts **off**, the collection pinned. This
  is what a golden or pixel test draws with, and the reason it exists: a test
  that says which faces it wants cannot drift when the default changes.

Widen the collection with `WithFaces`, which copies and clears both caches:
`tokens.DefaultTypography.WithFaces(notosansmono.FontFace())`. That is how a
test that legitimately draws an arrow stays deterministic, and how an
application that cannot rely on system fonts — a container, a kiosk — ships its
own symbol coverage. The face is optional and is not in
`DefaultTypography.Faces`; see ADR-003.

**Golden images.** None. Spectrum stores no rendered output — it computes
colour, type and layout values and asserts on numbers. The golden-image
harness lives in `prism/golden`, and the repos that render use it from
there.
