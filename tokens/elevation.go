// Tonal elevation (goal G-E2, task E2.1): elevation is a surface step.
//
// Per ADR-007, a raised surface separates from its ground primarily by
// colour — each elevation level fills with a step of the neutral ramp, one
// step deeper per storey — and only secondarily by a cast shadow. Because
// the light and dark ramps are paired scales (same step, same job), the
// same level reads as "raised" in both modes: a level-1 card is a light
// card on a lighter ground in light mode and a dark card on a darker
// ground in dark mode, with no mode-specific rule. The dp shadow survives
// as the secondary cue and is what pulse/depth still renders.
package tokens

import (
	"fmt"
	"image/color"
)

// ElevationScale pairs, per level, the neutral-ramp step of the level's
// surface fill (the primary cue) with its shadow depth in dp (the
// secondary cue).
//
// The LevelN fields keep their original names and dp meaning so existing
// readers — pulse/depth's dp lookup, spectrum/export's --shadow-* table —
// compile unchanged; the StepN fields carry the paired surface steps.
// Prefer the Dp and SurfaceStep accessors over field access in new code.
type ElevationScale struct {
	// Shadow depths in device-independent pixels, following Material
	// Design 3 elevation levels 0–5. The secondary cue.
	Level0 float32 // 0 dp
	Level1 float32 // 1 dp
	Level2 float32 // 3 dp
	Level3 float32 // 6 dp
	Level4 float32 // 8 dp
	Level5 float32 // 12 dp

	// Surface-fill steps on the neutral ramp. Step0 is not a ramp step:
	// its zero value marks the Background pin — a level-0 surface is the
	// app's bg pin sitting over the step-100 ground.
	Step0 int // 0 — sentinel: the Background pin, not a ramp step
	Step1 int // 200
	Step2 int // 300
	Step3 int // 400

	// Step4 and Step5 clamp to level 3's step, exactly D2.3's step-walk
	// clamp: desktop has no six-storey stack. The levels survive only so
	// existing call sites keep compiling.
	//
	// Deprecated: levels 4 and 5 are shims; marked for F3.3's shim sweep.
	Step4 int // 400 — clamped to Step3
	Step5 int // 400 — clamped to Step3
}

// Elevation is the default scale instance.
var Elevation = ElevationScale{
	Level0: 0,
	Level1: 1,
	Level2: 3,
	Level3: 6,
	Level4: 8,
	Level5: 12,

	Step0: 0, // Background pin
	Step1: 200,
	Step2: 300,
	Step3: 400,
	Step4: 400, // clamped to Step3 — F3.3 shim
	Step5: 400, // clamped to Step3 — F3.3 shim
}

// ElevationLevel selects an entry on the [ElevationScale] by name.
// The dp and step values for a given level are read from the [Elevation]
// instance.
type ElevationLevel int

const (
	Level0 ElevationLevel = iota
	Level1
	Level2
	Level3
	Level4
	Level5
)

// Dp returns level's shadow depth in device-independent pixels. An
// out-of-vocabulary level panics, matching [Ramp.Step].
func (e ElevationScale) Dp(level ElevationLevel) float32 {
	switch level {
	case Level0:
		return e.Level0
	case Level1:
		return e.Level1
	case Level2:
		return e.Level2
	case Level3:
		return e.Level3
	case Level4:
		return e.Level4
	case Level5:
		return e.Level5
	}
	panic(fmt.Sprintf("tokens: unknown ElevationLevel %d", level))
}

// SurfaceStep returns the neutral-ramp step of level's surface fill, or 0
// for a level whose fill is the Background pin rather than a ramp step
// (level 0 on the default scale). An out-of-vocabulary level panics,
// matching [Ramp.Step].
func (e ElevationScale) SurfaceStep(level ElevationLevel) int {
	switch level {
	case Level0:
		return e.Step0
	case Level1:
		return e.Step1
	case Level2:
		return e.Step2
	case Level3:
		return e.Step3
	case Level4:
		return e.Step4
	case Level5:
		return e.Step5
	}
	panic(fmt.Sprintf("tokens: unknown ElevationLevel %d", level))
}

// SurfaceAt resolves the surface colour of an elevated component: the fill
// of the given elevation level on t, per the default [Elevation] scale's
// step mapping. Level 0 is the Background pin over the step-100 ground;
// levels 1–3 fill with Neutral steps 200, 300 and 400; levels 4 and 5
// clamp to level 3's step (see [ElevationScale]).
//
// D2.3's state walks compose on top with the level's step as the ground:
// hover on a level-1 surface is StateColor(RoleNeutral, 200, StateHover),
// i.e. Neutral step 300 in both modes, courtesy of the paired scales. A
// level-0 surface is the app background, which has no ramp ground; treat
// interactive regions on it as level-1 surfaces instead.
func (t ColorTokens) SurfaceAt(level ElevationLevel) color.NRGBA {
	step := Elevation.SurfaceStep(level) // validates level
	if step == 0 {
		return t.Background
	}
	return t.Ramps.Neutral.Step(step)
}
