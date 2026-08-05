package tokens

import "testing"

// TestElevationDpPreserved pins the dp shadow depths: the secondary cue
// E2.1 keeps byte-for-byte, and what pulse/depth and spectrum/export read.
func TestElevationDpPreserved(t *testing.T) {
	want := map[ElevationLevel]float32{
		Level0: 0, Level1: 1, Level2: 3, Level3: 6, Level4: 8, Level5: 12,
	}
	for level, dp := range want {
		if got := Elevation.Dp(level); got != dp {
			t.Errorf("Elevation.Dp(%d) = %v, want %v", level, got, dp)
		}
	}
	// The legacy field names still carry the same dp values.
	fields := []float32{
		Elevation.Level0, Elevation.Level1, Elevation.Level2,
		Elevation.Level3, Elevation.Level4, Elevation.Level5,
	}
	for i, got := range fields {
		if want := Elevation.Dp(ElevationLevel(i)); got != want {
			t.Errorf("Elevation.Level%d field = %v, want %v", i, got, want)
		}
	}
}

// TestElevationSurfaceSteps asserts each level's fill sits on the neutral
// ramp — levels 1–3 on steps 200/300/400 — with level 0 the Background pin
// (sentinel step 0).
func TestElevationSurfaceSteps(t *testing.T) {
	want := map[ElevationLevel]int{
		Level0: 0, // Background pin, not a ramp step
		Level1: 200,
		Level2: 300,
		Level3: 400,
	}
	for level, step := range want {
		if got := Elevation.SurfaceStep(level); got != step {
			t.Errorf("Elevation.SurfaceStep(%d) = %d, want %d", level, got, step)
		}
	}
}

// TestElevationClamp asserts levels 4 and 5 clamp to level 3's step,
// exactly D2.3's step-walk clamp.
func TestElevationClamp(t *testing.T) {
	top := Elevation.SurfaceStep(Level3)
	for _, level := range []ElevationLevel{Level4, Level5} {
		if got := Elevation.SurfaceStep(level); got != top {
			t.Errorf("Elevation.SurfaceStep(%d) = %d, want clamp to level 3's %d", level, got, top)
		}
	}
	for _, mode := range []struct {
		name string
		tok  ColorTokens
	}{{"light", DefaultLight}, {"dark", DefaultDark}} {
		top := mode.tok.SurfaceAt(Level3)
		for _, level := range []ElevationLevel{Level4, Level5} {
			if got := mode.tok.SurfaceAt(level); got != top {
				t.Errorf("%s: SurfaceAt(%d) = %v, want level 3's %v", mode.name, level, got, top)
			}
		}
	}
}

// TestSurfaceAt asserts the resolver in both modes: level 0 is the
// Background pin, levels 1–3 the neutral ramp steps 200/300/400.
func TestSurfaceAt(t *testing.T) {
	for _, mode := range []struct {
		name string
		tok  ColorTokens
	}{{"light", DefaultLight}, {"dark", DefaultDark}} {
		if got := mode.tok.SurfaceAt(Level0); got != mode.tok.Background {
			t.Errorf("%s: SurfaceAt(Level0) = %v, want Background pin %v", mode.name, got, mode.tok.Background)
		}
		steps := map[ElevationLevel]int{Level1: 200, Level2: 300, Level3: 400}
		for level, step := range steps {
			if got, want := mode.tok.SurfaceAt(level), mode.tok.Ramps.Neutral.Step(step); got != want {
				t.Errorf("%s: SurfaceAt(%d) = %v, want Neutral.Step(%d) %v", mode.name, level, got, step, want)
			}
		}
	}
}

// TestSurfaceStateComposition asserts D2.3's state walks compose on top of
// the elevation mapping: a state on an elevated surface uses the level's
// step as its ground, so hover on a level-1 surface is step 300 and
// pressed is step 400 — in both modes, courtesy of the paired scales.
func TestSurfaceStateComposition(t *testing.T) {
	for _, mode := range []struct {
		name string
		tok  ColorTokens
	}{{"light", DefaultLight}, {"dark", DefaultDark}} {
		ground := Elevation.SurfaceStep(Level1) // 200
		if got, want := mode.tok.StateColor(RoleNeutral, ground, StateHover), mode.tok.Ramps.Neutral.Step(300); got != want {
			t.Errorf("%s: hover on level-1 surface = %v, want Neutral.Step(300) %v", mode.name, got, want)
		}
		if got, want := mode.tok.StateColor(RoleNeutral, ground, StatePressed), mode.tok.Ramps.Neutral.Step(400); got != want {
			t.Errorf("%s: pressed on level-1 surface = %v, want Neutral.Step(400) %v", mode.name, got, want)
		}
	}
}

// TestElevationLevelPanics asserts out-of-vocabulary levels panic,
// matching Ramp.Step.
func TestElevationLevelPanics(t *testing.T) {
	for _, level := range []ElevationLevel{-1, 6} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Elevation.SurfaceStep(%d) did not panic", level)
				}
			}()
			Elevation.SurfaceStep(level)
		}()
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Elevation.Dp(%d) did not panic", level)
				}
			}()
			Elevation.Dp(level)
		}()
	}
}
