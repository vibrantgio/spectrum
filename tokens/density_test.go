package tokens

import "testing"

// TestDensityPicksWithinTableBounds pins the picked control heights to the
// measured table in density.go: Compact is denser than Comfortable, and both
// stay within [28, 44] — at or above macOS's large control (the densest
// desktop reference) and below the touch-target floor 44 dp was rejected for.
func TestDensityPicksWithinTableBounds(t *testing.T) {
	if CompactControlHeight >= ComfortableControlHeight {
		t.Errorf("CompactControlHeight (%v) must be < ComfortableControlHeight (%v)",
			CompactControlHeight, ComfortableControlHeight)
	}
	for name, v := range map[string]float32{
		"ComfortableControlHeight": ComfortableControlHeight,
		"CompactControlHeight":     CompactControlHeight,
	} {
		if v < 28 || v > 44 {
			t.Errorf("%s = %v, want within [28, 44]", name, v)
		}
	}
	if ComfortableControlHeight >= MinHitTarget {
		t.Errorf("ComfortableControlHeight (%v) should sit below the hit-target floor (%v): the floor, not the control, is what 44 dp is for",
			ComfortableControlHeight, MinHitTarget)
	}
	if MinHitTarget != 44 {
		t.Errorf("MinHitTarget = %v, want 44 (WCAG 2.5.5, prism's current constant)", MinHitTarget)
	}
}
