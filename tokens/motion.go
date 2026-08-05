package tokens

import "time"

// Bezier holds the two inner control points of a cubic-bezier easing curve,
// equivalent to CSS cubic-bezier(P1.X, P1.Y, P2.X, P2.Y).
type Bezier struct {
	P1, P2 [2]float32
}

// MotionScale holds duration stops and easing presets for animation tokens.
type MotionScale struct {
	// Duration stops — strictly increasing fastest → slowest.
	DurXFast  time.Duration // 75 ms
	DurFast   time.Duration // 150 ms
	DurNormal time.Duration // 250 ms
	DurSlow   time.Duration // 400 ms
	DurXSlow  time.Duration // 700 ms

	// Easing presets (CSS standard curves).
	EaseLinear Bezier
	EaseIn     Bezier
	EaseOut    Bezier
	EaseInOut  Bezier
}

// Motion is the default scale instance.
var Motion = MotionScale{
	DurXFast:  75 * time.Millisecond,
	DurFast:   150 * time.Millisecond,
	DurNormal: 250 * time.Millisecond,
	DurSlow:   400 * time.Millisecond,
	DurXSlow:  700 * time.Millisecond,

	EaseLinear: Bezier{P1: [2]float32{0, 0}, P2: [2]float32{1, 1}},
	EaseIn:     Bezier{P1: [2]float32{0.4, 0}, P2: [2]float32{1, 1}},
	EaseOut:    Bezier{P1: [2]float32{0, 0}, P2: [2]float32{0.2, 1}},
	EaseInOut:  Bezier{P1: [2]float32{0.4, 0}, P2: [2]float32{0.2, 1}},
}
