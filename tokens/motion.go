package tokens

import (
	"math"
	"time"
)

// Bezier holds the two inner control points of a cubic-bezier easing curve,
// equivalent to CSS cubic-bezier(P1.X, P1.Y, P2.X, P2.Y).
type Bezier struct {
	P1, P2 [2]float32
}

// Spring holds the parameters of the textbook damped harmonic oscillator
// (m·ẍ = −k·x − c·ẋ) used by the pulse physics path: Mass is m, Stiffness
// is k, Damping is c. The three values are only meaningful together —
// critical damping (fastest settle, no overshoot) is c = 2·√(k·m), see
// [CriticalDamping] — so use a preset from [MotionScale] or set all three,
// never one.
type Spring struct {
	Mass, Stiffness, Damping float32
}

// CriticalDamping returns the critical damping coefficient c = 2·√(k·m)
// for the given stiffness and mass: the smallest damping at which a spring
// reaches its target without overshooting.
func CriticalDamping(stiffness, mass float32) float32 {
	return 2 * float32(math.Sqrt(float64(stiffness)*float64(mass)))
}

// MD3 motion semantics on a desktop scale, mapped 2026-08-05 (E3.1,
// per ADR-005: MD3's system, not MD3's look). This table is the
// justification for every value below it; argue with the sources here
// rather than with the consumer diffs.
//
// Easing. MD3 defines two easing families: standard for utilitarian
// transitions and emphasized for expressive, attention-drawing ones, each
// with accelerate (exit) and decelerate (enter) variants. The canonical
// cubic-beziers, verified against material-web design tokens v0.192
// (tokens/versions/v0_192/_md-sys-motion.scss, fetched 2026-08-05):
//
//	token                          cubic-bezier
//	-----                          ------------
//	easing-standard                (0.2, 0, 0, 1)
//	easing-standard-accelerate     (0.3, 0, 1, 1)
//	easing-standard-decelerate     (0, 0, 0, 1)
//	easing-emphasized              (0.2, 0, 0, 1)
//	easing-emphasized-accelerate   (0.3, 0, 0.8, 0.15)
//	easing-emphasized-decelerate   (0.05, 0.7, 0.1, 1)
//
// MD3's "full" emphasized ease is a two-segment path (an accelerate
// segment chained into a decelerate one) that a single cubic-bezier cannot
// represent. EaseEmphasized carries the documented single-bezier stand-in,
// which is material-web's own easing-emphasized value: the standard curve
// (0.2, 0, 0, 1). That is an honest approximation, not the spec path —
// the expressive character lives in the accelerate/decelerate pair, which
// single beziers do represent exactly.
//
// Durations. MD3 publishes sixteen duration roles — short1-4 (50-200 ms),
// medium1-4 (250-400), long1-4 (450-600), extra-long1-4 (700-1000), same
// source as above. Desktop keeps the five existing stops rather than
// adopting all sixteen: a pointer-driven desktop app wants fewer, faster
// stops (ADR-005's reading), and five is what every consumer already
// works from. Each stop is defined as exactly one MD3 role:
//
//	stop       was     MD3 role   now     used for
//	----       ---     --------   ---     --------
//	DurXFast   75 ms   short1     50 ms   state layers, hover feedback
//	DurFast    150 ms  short3     150 ms  small component transitions
//	DurNormal  250 ms  medium1    250 ms  standard enter/transition
//	DurSlow    400 ms  medium4    400 ms  emphasized/large transitions, fades
//	DurXSlow   500 ms  long2      500 ms  the slowest desktop motion; delays
//
// The reasoning: the three middle stops already sat on MD3 role values and
// do not move. DurXFast drops 75 → 50 (short1) because instant-feeling
// pointer feedback is the faster lean the desktop cut exists for. DurXSlow
// drops 700 → 500 (long2): the extra-long family (700-1000 ms) is
// full-screen, touch-scale choreography that a desktop window never needs,
// so the scale's ceiling is long2 — which is also exactly the 500 ms
// tooltip delay and the 30-frame (500 ms at 60 Hz) pulse/motion default
// already shipping, so the ceiling change costs no consumer a behaviour
// change.
//
// Springs. The pulse physics path animates with damped springs rather than
// beziers, so the scale carries spring presets alongside the curves. FX.2
// (the pulse/spring defaults fix) is not landed as of 2026-08-05 and owns
// the decision of what a usable default is; per that coordination these
// presets encode pulse's CURRENT working values rather than pre-empting
// FX.2 — it may retune them when it lands:
//
//	preset         mass  stiffness  damping        source
//	------         ----  ---------  -------        ------
//	SpringDefault  1     80         2·√80  ≈17.9   pulse/motion DefaultSpring (critical)
//	SpringSnappy   1     300        22     ζ≈0.64  pulse/springbutton defaults (slight overshoot, "pop")
//	SpringGentle   1     20         2·√20  ≈8.94   pulse/spring doc example (critical, soft)
//
// pulse/spring's own zero-Options fallback (k=0.4, c=0.7 — the ~873-frame
// settle FX.2 exists to fix) is deliberately NOT a preset.

// MotionScale holds duration stops, MD3 easing presets, and spring presets
// for animation tokens. See the mapping table above for where every value
// comes from.
type MotionScale struct {
	// Duration stops — strictly increasing fastest → slowest, each pinned
	// to one MD3 duration role (see the table above).
	DurXFast  time.Duration // MD3 short1, 50 ms
	DurFast   time.Duration // MD3 short3, 150 ms
	DurNormal time.Duration // MD3 medium1, 250 ms
	DurSlow   time.Duration // MD3 medium4, 400 ms
	DurXSlow  time.Duration // MD3 long2, 500 ms

	// MD3 standard easing family: utilitarian transitions.
	// Accelerate is for exits, Decelerate for enters.
	EaseStandard           Bezier
	EaseStandardAccelerate Bezier
	EaseStandardDecelerate Bezier

	// MD3 emphasized easing family: expressive transitions. EaseEmphasized
	// is the documented single-bezier stand-in for MD3's two-segment path
	// (see above); the accelerate/decelerate pair is exact.
	EaseEmphasized           Bezier
	EaseEmphasizedAccelerate Bezier
	EaseEmphasizedDecelerate Bezier

	// Spring presets for the pulse physics path. Values track pulse's
	// current working springs until FX.2 retunes them (see above).
	SpringDefault Spring // critically damped, brisk: enter/exit scale
	SpringSnappy  Spring // slightly underdamped: button-press "pop"
	SpringGentle  Spring // critically damped, soft: large soft reveals
}

// Reduced returns the reduce-motion variant of the scale: every duration
// stop is zero, everything else — easings and spring presets — is carried
// unchanged. It is what the theme's Motion field emits while the OS
// "Reduce Motion" preference is on (E3.2).
//
// Zero durations are the whole contract. A duration-driven animation of
// zero duration is complete the moment it starts — pulse/motion's
// FramesAt(0, fps) is 0 frames — so a component that derives its frame
// count from the scale reaches its target on the first frame it draws:
// it snaps. (Watch one pulse edge: pulse/motion's Options treats a zero
// Frames as "use the default", so a caller that snaps must skip the
// primitive on a zero duration rather than construct one with Frames 0.)
//
// The spring presets are deliberately NOT retuned. No finite spring
// completes in one frame, and a stiffness large enough to fake it
// (ω ≈ 600 rad/s for a 16 ms settle) is far outside the stability range
// of pulse/spring's explicit integrator at 60 Hz — it would oscillate or
// diverge, the opposite of reduced motion. A spring-driven component
// honours reduce-motion the same way a duration-driven one does: it reads
// the zero durations as the signal and jumps to its target instead of
// animating.
func (m MotionScale) Reduced() MotionScale {
	m.DurXFast = 0
	m.DurFast = 0
	m.DurNormal = 0
	m.DurSlow = 0
	m.DurXSlow = 0
	return m
}

// Motion is the default scale instance.
var Motion = MotionScale{
	DurXFast:  50 * time.Millisecond,
	DurFast:   150 * time.Millisecond,
	DurNormal: 250 * time.Millisecond,
	DurSlow:   400 * time.Millisecond,
	DurXSlow:  500 * time.Millisecond,

	EaseStandard:           Bezier{P1: [2]float32{0.2, 0}, P2: [2]float32{0, 1}},
	EaseStandardAccelerate: Bezier{P1: [2]float32{0.3, 0}, P2: [2]float32{1, 1}},
	EaseStandardDecelerate: Bezier{P1: [2]float32{0, 0}, P2: [2]float32{0, 1}},

	EaseEmphasized:           Bezier{P1: [2]float32{0.2, 0}, P2: [2]float32{0, 1}},
	EaseEmphasizedAccelerate: Bezier{P1: [2]float32{0.3, 0}, P2: [2]float32{0.8, 0.15}},
	EaseEmphasizedDecelerate: Bezier{P1: [2]float32{0.05, 0.7}, P2: [2]float32{0.1, 1}},

	SpringDefault: Spring{Mass: 1, Stiffness: 80, Damping: criticalDamping80},
	SpringSnappy:  Spring{Mass: 1, Stiffness: 300, Damping: 22},
	SpringGentle:  Spring{Mass: 1, Stiffness: 20, Damping: criticalDamping20},
}

// Critical damping coefficients for the preset springs, spelled as
// variables because math.Sqrt is not a constant expression.
var (
	criticalDamping80 = CriticalDamping(80, 1)
	criticalDamping20 = CriticalDamping(20, 1)
)
