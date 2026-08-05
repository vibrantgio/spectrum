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

// TestDensityHitTargetFloor discharges E1.2's independence checkbox: both
// density settings satisfy the WCAG 2.5.5 pointer-target floor, and the floor
// is identical across settings — Compact shrinks the drawn control, never the
// clickable area.
func TestDensityHitTargetFloor(t *testing.T) {
	for name, d := range map[string]Density{
		"Comfortable": Comfortable,
		"Compact":     Compact,
	} {
		if d.MinHitTarget() < 44 {
			t.Errorf("%s.MinHitTarget() = %v, want >= 44 (WCAG 2.5.5)", name, d.MinHitTarget())
		}
	}
	if Comfortable.MinHitTarget() != Compact.MinHitTarget() {
		t.Errorf("hit target must not vary with density: Comfortable %v != Compact %v",
			Comfortable.MinHitTarget(), Compact.MinHitTarget())
	}
}

// TestDensitySettingsMatchTable pins the two settings to E1.1's measured
// table: control heights are exactly the picked consts (so within the table's
// [28, 44] bounds already asserted above), Compact is strictly denser than
// Comfortable on every visual axis, and the paddings are the shadcn-derived
// pairs documented on the vars.
func TestDensitySettingsMatchTable(t *testing.T) {
	if Comfortable.ControlHeight != ComfortableControlHeight {
		t.Errorf("Comfortable.ControlHeight = %v, want %v", Comfortable.ControlHeight, ComfortableControlHeight)
	}
	if Compact.ControlHeight != CompactControlHeight {
		t.Errorf("Compact.ControlHeight = %v, want %v", Compact.ControlHeight, CompactControlHeight)
	}
	if Compact.ControlHeight >= Comfortable.ControlHeight {
		t.Errorf("Compact.ControlHeight (%v) must be < Comfortable.ControlHeight (%v)",
			Compact.ControlHeight, Comfortable.ControlHeight)
	}
	if Compact.PaddingX >= Comfortable.PaddingX || Compact.PaddingY >= Comfortable.PaddingY {
		t.Errorf("Compact padding (%v, %v) must be < Comfortable padding (%v, %v)",
			Compact.PaddingX, Compact.PaddingY, Comfortable.PaddingX, Comfortable.PaddingY)
	}
	if Comfortable.PaddingX != 16 || Comfortable.PaddingY != 8 {
		t.Errorf("Comfortable padding = (%v, %v), want (16, 8) — shadcn px-4 py-2",
			Comfortable.PaddingX, Comfortable.PaddingY)
	}
	if Compact.PaddingX != 12 || Compact.PaddingY != 6 {
		t.Errorf("Compact padding = (%v, %v), want (12, 6) — shadcn sm px-3, 2:1 vertical",
			Compact.PaddingX, Compact.PaddingY)
	}
}
