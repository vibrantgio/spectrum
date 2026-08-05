package tokens_test

import (
	"testing"
	"time"

	"github.com/vibrantgio/spectrum/tokens"
)

func TestSpacingMonotonic(t *testing.T) {
	steps := []struct {
		name string
		v    float32
	}{
		{"S0", tokens.Spacing.S0},
		{"S1", tokens.Spacing.S1},
		{"S2", tokens.Spacing.S2},
		{"S3", tokens.Spacing.S3},
		{"S4", tokens.Spacing.S4},
		{"S5", tokens.Spacing.S5},
		{"S6", tokens.Spacing.S6},
		{"S8", tokens.Spacing.S8},
		{"S10", tokens.Spacing.S10},
		{"S12", tokens.Spacing.S12},
		{"S16", tokens.Spacing.S16},
		{"S20", tokens.Spacing.S20},
		{"S24", tokens.Spacing.S24},
	}
	for i := 1; i < len(steps); i++ {
		if steps[i].v <= steps[i-1].v {
			t.Errorf("Spacing not monotonic: %s (%.0f) <= %s (%.0f)",
				steps[i].name, steps[i].v, steps[i-1].name, steps[i-1].v)
		}
	}
}

func TestRadiusMonotonic(t *testing.T) {
	steps := []struct {
		name string
		v    float32
	}{
		{"None", tokens.Radius.None},
		{"Sm", tokens.Radius.Sm},
		{"Base", tokens.Radius.Base},
		{"Md", tokens.Radius.Md},
		{"Lg", tokens.Radius.Lg},
		{"Xl", tokens.Radius.Xl},
		{"Xl2", tokens.Radius.Xl2},
		{"Xl3", tokens.Radius.Xl3},
		{"Full", tokens.Radius.Full},
	}
	for i := 1; i < len(steps); i++ {
		if steps[i].v <= steps[i-1].v {
			t.Errorf("Radius not monotonic: %s (%.0f) <= %s (%.0f)",
				steps[i].name, steps[i].v, steps[i-1].name, steps[i-1].v)
		}
	}
}

func TestElevationMonotonic(t *testing.T) {
	steps := []struct {
		name string
		v    float32
	}{
		{"Level0", tokens.Elevation.Level0},
		{"Level1", tokens.Elevation.Level1},
		{"Level2", tokens.Elevation.Level2},
		{"Level3", tokens.Elevation.Level3},
		{"Level4", tokens.Elevation.Level4},
		{"Level5", tokens.Elevation.Level5},
	}
	for i := 1; i < len(steps); i++ {
		if steps[i].v <= steps[i-1].v {
			t.Errorf("Elevation not monotonic: %s (%.0f) <= %s (%.0f)",
				steps[i].name, steps[i].v, steps[i-1].name, steps[i-1].v)
		}
	}
}

func TestMotionReducedZeroesEveryDuration(t *testing.T) {
	r := tokens.Motion.Reduced()
	durations := []struct {
		name string
		v    time.Duration
	}{
		{"DurXFast", r.DurXFast},
		{"DurFast", r.DurFast},
		{"DurNormal", r.DurNormal},
		{"DurSlow", r.DurSlow},
		{"DurXSlow", r.DurXSlow},
	}
	for _, d := range durations {
		if d.v != 0 {
			t.Errorf("Reduced().%s = %v, want 0 — reduce-motion animations must complete immediately", d.name, d.v)
		}
	}
}

func TestMotionReducedKeepsEverythingElse(t *testing.T) {
	// Springs and easings carry through unchanged: no finite spring
	// completes in one frame, and a snap-stiff spring would be unstable
	// in pulse's explicit integrator, so the zero durations alone are
	// the reduce-motion signal (see the Reduced doc).
	r := tokens.Motion.Reduced()
	want := tokens.Motion
	want.DurXFast, want.DurFast, want.DurNormal, want.DurSlow, want.DurXSlow = 0, 0, 0, 0, 0
	if r != want {
		t.Errorf("Reduced() changed more than the durations:\ngot  %+v\nwant %+v", r, want)
	}
}

func TestMotionDurationsMonotonic(t *testing.T) {
	steps := []struct {
		name string
		v    time.Duration
	}{
		{"DurXFast", tokens.Motion.DurXFast},
		{"DurFast", tokens.Motion.DurFast},
		{"DurNormal", tokens.Motion.DurNormal},
		{"DurSlow", tokens.Motion.DurSlow},
		{"DurXSlow", tokens.Motion.DurXSlow},
	}
	for i := 1; i < len(steps); i++ {
		if steps[i].v <= steps[i-1].v {
			t.Errorf("Motion durations not monotonic: %s (%v) <= %s (%v)",
				steps[i].name, steps[i].v, steps[i-1].name, steps[i-1].v)
		}
	}
}
