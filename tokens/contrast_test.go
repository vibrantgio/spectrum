package tokens_test

import (
	stdcolor "image/color"
	"math"
	"testing"

	"github.com/vibrantgio/spectrum/color"
	"github.com/vibrantgio/spectrum/tokens"
)

// TestAPCAContrastGate is ADR-007's contrast gate over the default palette,
// in both modes: in every role ramp, step 900 must reach |Lc| ≥ 90 and step
// 700 |Lc| ≥ 60 over the step-100 and step-200 grounds, and each pinned
// base's on-colour |Lc| ≥ 60 over the base.
//
// Reading: ADR-007's sentence — "in both ramps, step 900 must reach Lc 90
// and step 700 Lc 60 over the step-100 and step-200 grounds" — is read with
// the grounds taken from the SAME role's ramp, because the ADR assigns
// 700–900 the job "text over tinted fills and pressed states" and the
// tinted fills 100–300 come from the ramp being read. Since every ramp
// shares one lightness scale, the neutral-grounds reading differs only by
// hue-induced luminance wiggle; the same-role reading covers neutral anyway
// (neutral is one of the five gated ramps).
//
// WCAG 2 ratios for the same pairs are logged alongside — conformance
// claims cite them per ADR-007 — but they do not gate: only APCA failures
// fail this test, so a WCAG regression shows up in the log, never as a
// verdict.
func TestAPCAContrastGate(t *testing.T) {
	for _, s := range []struct {
		name string
		tok  tokens.ColorTokens
	}{
		{"DefaultLight", tokens.DefaultLight},
		{"DefaultDark", tokens.DefaultDark},
	} {
		t.Run(s.name, func(t *testing.T) {
			for _, r := range namedRamps(s.tok) {
				for _, tc := range []struct {
					textStep int
					minLc    float64
				}{
					{900, 90},
					{700, 60},
				} {
					for _, groundStep := range []int{100, 200} {
						text := r.ramp.Step(tc.textStep)
						ground := r.ramp.Step(groundStep)
						lc := color.APCA(text, ground)
						wcag := color.ContrastRatio(text, ground)
						t.Logf("%s %d on %d: Lc %.2f (gate ≥ %.0f), WCAG %.2f:1 (AA %.1f:1: %s, cited not gating)",
							r.name, tc.textStep, groundStep, lc, tc.minLc, wcag, wcagAA, wcagVerdict(wcag))
						if math.Abs(lc) < tc.minLc {
							t.Errorf("%s: step %d on step-%d ground: |Lc| %.2f < %.0f",
								r.name, tc.textStep, groundStep, math.Abs(lc), tc.minLc)
						}
					}
				}
			}
			for _, p := range []struct {
				name     string
				base, on stdcolor.NRGBA
			}{
				{"Primary", s.tok.Primary, s.tok.OnPrimary},
				{"Secondary", s.tok.Secondary, s.tok.OnSecondary},
				{"Tertiary", s.tok.Tertiary, s.tok.OnTertiary},
				{"Error", s.tok.Error, s.tok.OnError},
			} {
				lc := color.APCA(p.on, p.base)
				wcag := color.ContrastRatio(p.on, p.base)
				t.Logf("pin %s: on-colour Lc %.2f (gate ≥ 60), WCAG %.2f:1 (AA %.1f:1: %s, cited not gating)",
					p.name, lc, wcag, wcagAA, wcagVerdict(wcag))
				if math.Abs(lc) < 60 {
					t.Errorf("pin %s: on-colour |Lc| %.2f < 60", p.name, math.Abs(lc))
				}
			}
		})
	}
}

// wcagVerdict renders a WCAG AA pass/fail for the gate test's log lines —
// reported, never gated on.
func wcagVerdict(ratio float64) string {
	if ratio >= wcagAA {
		return "pass"
	}
	return "fail"
}

// TestAPCAContrastGateHighContrast records the E3.3 gap: ADR-007's gate
// must also hold for the high-contrast variant, but that variant does not
// exist yet — E3.3 (high-contrast palette) has not landed. When it does,
// extend TestAPCAContrastGate's scheme table with the variant and delete
// this skip.
func TestAPCAContrastGateHighContrast(t *testing.T) {
	t.Skip("E3.3 (high-contrast palette) has not landed; gate its variant here when it does")
}
