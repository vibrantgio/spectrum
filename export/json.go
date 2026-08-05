package export

import (
	"encoding/json"
	"math"
	"strconv"
	"time"

	"github.com/vibrantgio/spectrum/color"
	"github.com/vibrantgio/spectrum/tokens"
)

// Parameters is theme.json's shape: the generative parameters that
// reproduce the theme. tokens.FromSeed(Seed) regenerates every ramp and pin
// — the round-trip test asserts it — so the file alone rebuilds the
// palette; the pins, scales, fonts, radius, density settings, elevation
// model and motion set are recorded alongside so a reader (or a prototype)
// need not run the generator to know them.
type Parameters struct {
	// Seed is the brand seed as lowercase #rrggbb; Hue and Sat are its
	// OKLCh hue (degrees, 2 decimals) and chroma (4 decimals), recorded for
	// the reader — regeneration starts from the hex.
	Seed string  `json:"seed"`
	Hue  float64 `json:"hue"`
	Sat  float64 `json:"sat"`

	// Pins are the pinned role bases per mode.
	Pins ModePins `json:"pins"`

	// Fonts names the heading and body faces. Both are Roboto until a
	// heading face exists.
	Fonts Fonts `json:"fonts"`

	// Radius is the base radius in dp — tokens.RadiusScale.Base, the sheet's
	// --radius-base.
	Radius float64 `json:"radius"`

	// Scale is the shared CIELAB L* lightness scale per mode, steps
	// 100–900, measured back from the emitted neutral ramps. It documents
	// the generator's fixed scale (ADR-007); it is not itself an input —
	// FromSeed carries it.
	Scale ModeScale `json:"scale"`

	// Density records the theme's active setting by name plus both
	// published settings' metrics, and the density-invariant pointer-target
	// floor.
	Density DensityParams `json:"density"`

	// Elevation records the tonal model per level 0–5: the neutral-ramp
	// step of the surface fill (the default cue; 0 marks the bg pin, not a
	// ramp step) and the dp shadow depth (the opt-in cue for floating
	// transients).
	Elevation ElevationParams `json:"elevation"`

	// Motion records the captured MotionScale in full — duration stops,
	// easing beziers and spring presets — so the file reproduces it without
	// running the generator. Springs are Go-side physics with no CSS
	// counterpart; this is their only serialisation.
	Motion MotionParams `json:"motion"`
}

// DensityParams records the density model: the active setting's name
// ("comfortable" or "compact"), both settings' metrics, and the WCAG 2.5.5
// pointer-target minimum in dp, which no setting scales.
type DensityParams struct {
	Setting      string         `json:"setting"`
	Comfortable  DensityMetrics `json:"comfortable"`
	Compact      DensityMetrics `json:"compact"`
	MinHitTarget float64        `json:"minHitTarget"`
}

// DensityMetrics is one density setting's per-setting metrics in dp.
type DensityMetrics struct {
	ControlHeight float64 `json:"controlHeight"`
	PaddingX      float64 `json:"paddingX"`
	PaddingY      float64 `json:"paddingY"`
}

// ElevationParams pairs, indexed by level 0–5, the neutral-ramp surface
// step (0 = the bg pin sentinel; levels 4 and 5 clamp to level 3's step)
// with the dp shadow depth.
type ElevationParams struct {
	SurfaceSteps [6]int     `json:"surfaceSteps"`
	ShadowDp     [6]float64 `json:"shadowDp"`
}

// MotionParams records the motion set.
type MotionParams struct {
	Durations DurationParams `json:"durations"`
	Easings   EasingParams   `json:"easings"`
	Springs   SpringParams   `json:"springs"`
}

// DurationParams carries the five duration stops in milliseconds.
type DurationParams struct {
	XFast  float64 `json:"xFast"`
	Fast   float64 `json:"fast"`
	Normal float64 `json:"normal"`
	Slow   float64 `json:"slow"`
	XSlow  float64 `json:"xSlow"`
}

// EasingParams carries each easing preset as its cubic-bezier control
// points [x1, y1, x2, y2] — the same four numbers the sheet's --ease-*
// variables carry inside cubic-bezier().
type EasingParams struct {
	Standard             [4]float64 `json:"standard"`
	StandardAccelerate   [4]float64 `json:"standardAccelerate"`
	StandardDecelerate   [4]float64 `json:"standardDecelerate"`
	Emphasized           [4]float64 `json:"emphasized"`
	EmphasizedAccelerate [4]float64 `json:"emphasizedAccelerate"`
	EmphasizedDecelerate [4]float64 `json:"emphasizedDecelerate"`
}

// SpringParams carries the three spring presets.
type SpringParams struct {
	Default SpringParam `json:"default"`
	Snappy  SpringParam `json:"snappy"`
	Gentle  SpringParam `json:"gentle"`
}

// SpringParam is one damped-oscillator preset. Damping is recorded at the
// shortest decimal that reproduces the float32 exactly (the critical
// presets are 2·√(k·m), an irrational number), so the file reproduces the
// Go value bit-for-bit — see f64.
type SpringParam struct {
	Mass      float64 `json:"mass"`
	Stiffness float64 `json:"stiffness"`
	Damping   float64 `json:"damping"`
}

// f64 widens a float32 token value for JSON without dragging float64
// conversion noise into the file: it goes through the shortest decimal
// representation of the float32 (0.2, not 0.20000000298023224), which
// converts back to the identical float32 — the round-trip test asserts
// exactly that — while staying readable.
func f64(v float32) float64 {
	f, err := strconv.ParseFloat(strconv.FormatFloat(float64(v), 'g', -1, 32), 64)
	if err != nil {
		panic("export: f64: " + err.Error()) // unreachable: FormatFloat output always parses
	}
	return f
}

// ModePins carries the pinned bases for both modes.
type ModePins struct {
	Light Pins `json:"light"`
	Dark  Pins `json:"dark"`
}

// Pins records one mode's pinned bases as lowercase #rrggbb hexes. Accent
// is the primary pin — the sheet's --color-accent.
type Pins struct {
	Bg        string `json:"bg"`
	Text      string `json:"text"`
	Accent    string `json:"accent"`
	Secondary string `json:"secondary"`
	Tertiary  string `json:"tertiary"`
	Error     string `json:"error"`
}

// Fonts names the typefaces.
type Fonts struct {
	Heading string `json:"heading"`
	Body    string `json:"body"`
}

// ModeScale carries the L* scale for both modes.
type ModeScale struct {
	Light [9]int `json:"light"`
	Dark  [9]int `json:"dark"`
}

// pinsOf reads one scheme's pinned bases.
func pinsOf(t tokens.ColorTokens) Pins {
	return Pins{
		Bg:        hexRGB(t.Background),
		Text:      hexRGB(t.Text),
		Accent:    hexRGB(t.Primary),
		Secondary: hexRGB(t.Secondary),
		Tertiary:  hexRGB(t.Tertiary),
		Error:     hexRGB(t.Error),
	}
}

// measuredScale reads the CIELAB L* of each ramp step back from the ramp
// itself, rounded to the nearest integer — the documented scale values are
// integral, and 8-bit quantisation keeps the measurement well within
// rounding distance.
func measuredScale(r tokens.Ramp) [9]int {
	var s [9]int
	for i, c := range r {
		L, _, _ := color.LabFromNRGBA(c)
		s[i] = int(math.Round(L))
	}
	return s
}

// densityMetricsOf reads one setting's metrics.
func densityMetricsOf(d tokens.Density) DensityMetrics {
	return DensityMetrics{
		ControlHeight: f64(d.ControlHeight),
		PaddingX:      f64(d.PaddingX),
		PaddingY:      f64(d.PaddingY),
	}
}

// bezier4 flattens a Bezier into its cubic-bezier() control points.
func bezier4(bz tokens.Bezier) [4]float64 {
	return [4]float64{f64(bz.P1[0]), f64(bz.P1[1]), f64(bz.P2[0]), f64(bz.P2[1])}
}

// springOf flattens a Spring preset.
func springOf(sp tokens.Spring) SpringParam {
	return SpringParam{Mass: f64(sp.Mass), Stiffness: f64(sp.Stiffness), Damping: f64(sp.Damping)}
}

// durMs is a duration in milliseconds.
func durMs(d time.Duration) float64 {
	return float64(d) / float64(time.Millisecond)
}

// parameters assembles the Parameters for a snapshot.
func parameters(s Snapshot) Parameters {
	_, chroma, hue := color.OKLChFromNRGBA(s.Seed)
	setting, _ := densitySetting(s.Density) // Capture already validated it
	var elev ElevationParams
	for i, level := range elevationLevels {
		elev.SurfaceSteps[i] = s.Elevation.SurfaceStep(level.level)
		elev.ShadowDp[i] = f64(s.Elevation.Dp(level.level))
	}
	return Parameters{
		Seed:   hexRGB(s.Seed),
		Hue:    math.Round(hue*100) / 100,
		Sat:    math.Round(chroma*10000) / 10000,
		Pins:   ModePins{Light: pinsOf(s.Light), Dark: pinsOf(s.Dark)},
		Fonts:  Fonts{Heading: s.Typography.HeadlineLarge.Typeface, Body: s.Typography.BodyLarge.Typeface},
		Radius: float64(s.Radius.Base),
		Scale: ModeScale{
			Light: measuredScale(s.Light.Ramps.Neutral),
			Dark:  measuredScale(s.Dark.Ramps.Neutral),
		},
		Density: DensityParams{
			Setting:      setting,
			Comfortable:  densityMetricsOf(tokens.Comfortable),
			Compact:      densityMetricsOf(tokens.Compact),
			MinHitTarget: f64(s.Density.MinHitTarget()),
		},
		Elevation: elev,
		Motion: MotionParams{
			Durations: DurationParams{
				XFast:  durMs(s.Motion.DurXFast),
				Fast:   durMs(s.Motion.DurFast),
				Normal: durMs(s.Motion.DurNormal),
				Slow:   durMs(s.Motion.DurSlow),
				XSlow:  durMs(s.Motion.DurXSlow),
			},
			Easings: EasingParams{
				Standard:             bezier4(s.Motion.EaseStandard),
				StandardAccelerate:   bezier4(s.Motion.EaseStandardAccelerate),
				StandardDecelerate:   bezier4(s.Motion.EaseStandardDecelerate),
				Emphasized:           bezier4(s.Motion.EaseEmphasized),
				EmphasizedAccelerate: bezier4(s.Motion.EaseEmphasizedAccelerate),
				EmphasizedDecelerate: bezier4(s.Motion.EaseEmphasizedDecelerate),
			},
			Springs: SpringParams{
				Default: springOf(s.Motion.SpringDefault),
				Snappy:  springOf(s.Motion.SpringSnappy),
				Gentle:  springOf(s.Motion.SpringGentle),
			},
		},
	}
}

// themeJSON renders theme.json, indented and newline-terminated.
func themeJSON(s Snapshot) ([]byte, error) {
	js, err := json.MarshalIndent(parameters(s), "", "  ")
	if err != nil {
		return nil, err
	}
	return append(js, '\n'), nil
}
