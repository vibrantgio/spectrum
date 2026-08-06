package export

import (
	"fmt"
	stdcolor "image/color"
	"strconv"
	"strings"
	"time"

	"github.com/vibrantgio/spectrum/tokens"
)

// cssVar is one custom property, emitted in declaration order.
type cssVar struct {
	name, value string
}

// hexRGB formats a colour as lowercase #rrggbb. Every token colour is fully
// opaque, so alpha is never written.
func hexRGB(c stdcolor.NRGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

// fnum formats a float32 with no trailing zeros: 16 → "16", 0.15 → "0.15".
func fnum(v float32) string {
	return strconv.FormatFloat(float64(v), 'f', -1, 32)
}

// px formats a dp value as a CSS px length. Device-independent pixels map
// 1:1 onto CSS px, both being density-abstract logical pixels.
func px(v float32) string {
	return fnum(v) + "px"
}

// rampRoles orders ADR-007's five colour roles under their CSS names.
var rampRoles = []struct {
	name string
	ramp func(tokens.RampSet) tokens.Ramp
}{
	{"neutral", func(r tokens.RampSet) tokens.Ramp { return r.Neutral }},
	{"primary", func(r tokens.RampSet) tokens.Ramp { return r.Primary }},
	{"secondary", func(r tokens.RampSet) tokens.Ramp { return r.Secondary }},
	{"tertiary", func(r tokens.RampSet) tokens.Ramp { return r.Tertiary }},
	{"error", func(r tokens.RampSet) tokens.Ramp { return r.Error }},
}

// pinRoles orders the pinned bases and the semantic layer under their CSS
// names; the doc comment on this package records the mapping.
var pinRoles = []struct {
	name string
	pick func(tokens.ColorTokens) stdcolor.NRGBA
}{
	{"bg", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.Background }},
	{"surface", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.Surface }},
	{"text", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.Text }},
	{"divider", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.Divider }},
	{"accent", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.Primary }},
	{"on-accent", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.OnPrimary }},
	{"secondary", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.Secondary }},
	{"on-secondary", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.OnSecondary }},
	{"tertiary", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.Tertiary }},
	{"on-tertiary", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.OnTertiary }},
	{"error", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.Error }},
	{"on-error", func(t tokens.ColorTokens) stdcolor.NRGBA { return t.OnError }},
}

// typeRoles orders the fifteen MD3 type roles under their CSS names, plus
// code — the sixteenth style outside the MD3 grid, the mono face at
// body-medium's metrics (G-F0) — emitted last.
var typeRoles = []struct {
	name string
	pick func(tokens.Typography) tokens.TextStyle
}{
	{"display-large", func(t tokens.Typography) tokens.TextStyle { return t.DisplayLarge }},
	{"display-medium", func(t tokens.Typography) tokens.TextStyle { return t.DisplayMedium }},
	{"display-small", func(t tokens.Typography) tokens.TextStyle { return t.DisplaySmall }},
	{"headline-large", func(t tokens.Typography) tokens.TextStyle { return t.HeadlineLarge }},
	{"headline-medium", func(t tokens.Typography) tokens.TextStyle { return t.HeadlineMedium }},
	{"headline-small", func(t tokens.Typography) tokens.TextStyle { return t.HeadlineSmall }},
	{"title-large", func(t tokens.Typography) tokens.TextStyle { return t.TitleLarge }},
	{"title-medium", func(t tokens.Typography) tokens.TextStyle { return t.TitleMedium }},
	{"title-small", func(t tokens.Typography) tokens.TextStyle { return t.TitleSmall }},
	{"label-large", func(t tokens.Typography) tokens.TextStyle { return t.LabelLarge }},
	{"label-medium", func(t tokens.Typography) tokens.TextStyle { return t.LabelMedium }},
	{"label-small", func(t tokens.Typography) tokens.TextStyle { return t.LabelSmall }},
	{"body-large", func(t tokens.Typography) tokens.TextStyle { return t.BodyLarge }},
	{"body-medium", func(t tokens.Typography) tokens.TextStyle { return t.BodyMedium }},
	{"body-small", func(t tokens.Typography) tokens.TextStyle { return t.BodySmall }},
	{"code", func(t tokens.Typography) tokens.TextStyle { return t.Code }},
}

// spaceKeys orders the spacing stops under the Go scale's own key names.
var spaceKeys = []struct {
	name string
	pick func(tokens.SpacingScale) float32
}{
	{"0", func(s tokens.SpacingScale) float32 { return s.S0 }},
	{"1", func(s tokens.SpacingScale) float32 { return s.S1 }},
	{"2", func(s tokens.SpacingScale) float32 { return s.S2 }},
	{"3", func(s tokens.SpacingScale) float32 { return s.S3 }},
	{"4", func(s tokens.SpacingScale) float32 { return s.S4 }},
	{"5", func(s tokens.SpacingScale) float32 { return s.S5 }},
	{"6", func(s tokens.SpacingScale) float32 { return s.S6 }},
	{"8", func(s tokens.SpacingScale) float32 { return s.S8 }},
	{"10", func(s tokens.SpacingScale) float32 { return s.S10 }},
	{"12", func(s tokens.SpacingScale) float32 { return s.S12 }},
	{"16", func(s tokens.SpacingScale) float32 { return s.S16 }},
	{"20", func(s tokens.SpacingScale) float32 { return s.S20 }},
	{"24", func(s tokens.SpacingScale) float32 { return s.S24 }},
}

// radiusKeys orders the radius stops under Tailwind's names, which the Go
// field names mirror (Xl2 and Xl3 are Go spellings of 2xl and 3xl).
var radiusKeys = []struct {
	name string
	pick func(tokens.RadiusScale) float32
}{
	{"none", func(r tokens.RadiusScale) float32 { return r.None }},
	{"sm", func(r tokens.RadiusScale) float32 { return r.Sm }},
	{"base", func(r tokens.RadiusScale) float32 { return r.Base }},
	{"md", func(r tokens.RadiusScale) float32 { return r.Md }},
	{"lg", func(r tokens.RadiusScale) float32 { return r.Lg }},
	{"xl", func(r tokens.RadiusScale) float32 { return r.Xl }},
	{"2xl", func(r tokens.RadiusScale) float32 { return r.Xl2 }},
	{"3xl", func(r tokens.RadiusScale) float32 { return r.Xl3 }},
	{"full", func(r tokens.RadiusScale) float32 { return r.Full }},
}

// elevationLevels orders the elevation levels under their level numbers;
// each level's surface step and shadow dp are read off the snapshot's
// ElevationScale through its accessors.
var elevationLevels = []struct {
	name  string
	level tokens.ElevationLevel
}{
	{"0", tokens.Level0},
	{"1", tokens.Level1},
	{"2", tokens.Level2},
	{"3", tokens.Level3},
}

// densityMetrics orders the per-setting density metrics under their CSS
// names. The WCAG pointer-target floor is not here: it is not a per-setting
// metric — see densityVars.
var densityMetrics = []struct {
	name string
	pick func(tokens.Density) float32
}{
	{"control-height", func(d tokens.Density) float32 { return d.ControlHeight }},
	{"padding-x", func(d tokens.Density) float32 { return d.PaddingX }},
	{"padding-y", func(d tokens.Density) float32 { return d.PaddingY }},
}

// easeRoles orders the MD3 easing presets under their CSS names.
var easeRoles = []struct {
	name string
	pick func(tokens.MotionScale) tokens.Bezier
}{
	{"standard", func(m tokens.MotionScale) tokens.Bezier { return m.EaseStandard }},
	{"standard-accelerate", func(m tokens.MotionScale) tokens.Bezier { return m.EaseStandardAccelerate }},
	{"standard-decelerate", func(m tokens.MotionScale) tokens.Bezier { return m.EaseStandardDecelerate }},
	{"emphasized", func(m tokens.MotionScale) tokens.Bezier { return m.EaseEmphasized }},
	{"emphasized-accelerate", func(m tokens.MotionScale) tokens.Bezier { return m.EaseEmphasizedAccelerate }},
	{"emphasized-decelerate", func(m tokens.MotionScale) tokens.Bezier { return m.EaseEmphasizedDecelerate }},
}

// durationStops orders the duration stops under their CSS names.
var durationStops = []struct {
	name string
	pick func(tokens.MotionScale) time.Duration
}{
	{"x-fast", func(m tokens.MotionScale) time.Duration { return m.DurXFast }},
	{"fast", func(m tokens.MotionScale) time.Duration { return m.DurFast }},
	{"normal", func(m tokens.MotionScale) time.Duration { return m.DurNormal }},
	{"slow", func(m tokens.MotionScale) time.Duration { return m.DurSlow }},
	{"x-slow", func(m tokens.MotionScale) time.Duration { return m.DurXSlow }},
}

// boxShadow approximates an elevation depth as a CSS box-shadow: y-offset
// the level's dp, blur twice it, no spread, black at 20%. Depth 0 casts no
// shadow at all, so it is "none" rather than an invisible shadow.
func boxShadow(dp float32) string {
	if dp == 0 {
		return "none"
	}
	return fmt.Sprintf("0 %s %s 0 rgba(0, 0, 0, 0.2)", px(dp), px(2*dp))
}

// surfaceVarRef renders an elevation level's surface fill as a reference
// into the colour families — var(--color-bg) for the step-0 sentinel (the
// Background pin), var(--color-neutral-<step>) otherwise. Emitting a
// reference rather than a resolved hex keeps --elevation-* mode-invariant:
// the .dark block overrides the colour variables and every elevation
// surface flips with them, which is exactly the tonal model (an elevation
// level IS a neutral-ramp step).
func surfaceVarRef(step int) string {
	if step == 0 {
		return "var(--color-bg)"
	}
	return fmt.Sprintf("var(--color-neutral-%d)", step)
}

// cubicBezier renders a Bezier as the CSS cubic-bezier() function.
func cubicBezier(bz tokens.Bezier) string {
	return fmt.Sprintf("cubic-bezier(%s, %s, %s, %s)",
		fnum(bz.P1[0]), fnum(bz.P1[1]), fnum(bz.P2[0]), fnum(bz.P2[1]))
}

// ms formats a duration as CSS milliseconds.
func ms(d time.Duration) string {
	return strconv.FormatFloat(float64(d)/float64(time.Millisecond), 'f', -1, 64) + "ms"
}

// colorVars renders one colour scheme as its ramp and pin variables.
func colorVars(t tokens.ColorTokens) []cssVar {
	var vars []cssVar
	for _, role := range rampRoles {
		ramp := role.ramp(t.Ramps)
		for step := 100; step <= 900; step += 100 {
			vars = append(vars, cssVar{
				name:  fmt.Sprintf("--color-%s-%d", role.name, step),
				value: hexRGB(ramp.Step(step)),
			})
		}
	}
	for _, pin := range pinRoles {
		vars = append(vars, cssVar{"--color-" + pin.name, hexRGB(pin.pick(t))})
	}
	return vars
}

// scaleVars renders the mode-invariant families: fonts, density
// (comfortable — the :root setting), spacing, radius, the tonal elevation
// surfaces (the default cue), the dp shadows (the opt-in cue for floating
// transients, per E2.2), and the motion set.
func scaleVars(s Snapshot) []cssVar {
	vars := []cssVar{
		{"--font-family", strconv.Quote(s.Typography.BodyLarge.Typeface)},
		{"--font-family-code", strconv.Quote(s.Typography.Code.Typeface)},
	}
	for _, role := range typeRoles {
		style := role.pick(s.Typography)
		vars = append(vars,
			cssVar{"--font-" + role.name + "-size", px(style.Size)},
			cssVar{"--font-" + role.name + "-line-height", px(style.LineHeight)},
			cssVar{"--font-" + role.name + "-weight", strconv.Itoa(style.Weight)},
			cssVar{"--font-" + role.name + "-tracking", px(style.Tracking)},
		)
	}
	vars = append(vars, densityVars(tokens.Comfortable)...)
	vars = append(vars, cssVar{"--density-min-hit-target", px(tokens.Comfortable.MinHitTarget())})
	for _, key := range spaceKeys {
		vars = append(vars, cssVar{"--space-" + key.name, px(key.pick(s.Spacing))})
	}
	for _, key := range radiusKeys {
		vars = append(vars, cssVar{"--radius-" + key.name, px(key.pick(s.Radius))})
	}
	for _, level := range elevationLevels {
		vars = append(vars, cssVar{"--elevation-" + level.name, surfaceVarRef(s.Elevation.SurfaceStep(level.level))})
	}
	for _, level := range elevationLevels {
		vars = append(vars, cssVar{"--shadow-" + level.name, boxShadow(s.Elevation.Dp(level.level))})
	}
	for _, role := range easeRoles {
		vars = append(vars, cssVar{"--ease-" + role.name, cubicBezier(role.pick(s.Motion))})
	}
	for _, stop := range durationStops {
		vars = append(vars, cssVar{"--duration-" + stop.name, ms(stop.pick(s.Motion))})
	}
	return vars
}

// densityVars renders one density setting's per-setting metrics. The :root
// block carries tokens.Comfortable's; the .compact override block carries
// tokens.Compact's. --density-min-hit-target is deliberately not among
// them: the WCAG 2.5.5 pointer-target floor does not scale with density, so
// it is emitted once in :root and never overridden — the CSS mirror of
// Density.MinHitTarget being a method, not a field.
func densityVars(d tokens.Density) []cssVar {
	var vars []cssVar
	for _, m := range densityMetrics {
		vars = append(vars, cssVar{"--density-" + m.name, px(m.pick(d))})
	}
	return vars
}

// block renders one selector's declarations.
func block(b *strings.Builder, selector string, vars []cssVar) {
	b.WriteString(selector)
	b.WriteString(" {\n")
	for _, v := range vars {
		fmt.Fprintf(b, "  %s: %s;\n", v.name, v.value)
	}
	b.WriteString("}\n")
}

// stylesCSS renders the full token sheet: the light scheme and every
// mode-invariant scale under :root, the paired dark colours under .dark,
// and the compact density metrics under .compact. The two class blocks are
// orthogonal switches — .dark flips the colours (and with them the
// var()-chained --elevation-* surfaces), .compact flips the per-setting
// density metrics — so a surface can be any of the four combinations.
func stylesCSS(s Snapshot) string {
	var b strings.Builder
	b.WriteString("/* Generated by spectrum/export (cmd/vg-tokens). Do not edit. */\n\n")
	block(&b, ":root", append(colorVars(s.Light), scaleVars(s)...))
	b.WriteString("\n")
	block(&b, ".dark", colorVars(s.Dark))
	b.WriteString("\n")
	block(&b, ".compact", densityVars(tokens.Compact))
	return b.String()
}
