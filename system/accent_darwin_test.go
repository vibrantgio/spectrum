package system

import "testing"

// TestAccentFromIndex pins the raw AppleAccentColor → Accent table for the
// full documented range −1..7 (7 is unassigned by macOS and must fold to
// the default), plus out-of-range values.
func TestAccentFromIndex(t *testing.T) {
	cases := []struct {
		raw  int
		want Accent
	}{
		{-1, AccentGraphite},
		{0, AccentRed},
		{1, AccentOrange},
		{2, AccentYellow},
		{3, AccentGreen},
		{4, AccentBlue},
		{5, AccentPurple},
		{6, AccentPink},
		{7, AccentDefault},  // unassigned by macOS; fold, don't guess
		{-2, AccentDefault}, // out of range low
		{42, AccentDefault}, // out of range high
	}
	for _, tc := range cases {
		if got := accentFromIndex(tc.raw); got != tc.want {
			t.Errorf("accentFromIndex(%d) = %d, want %d", tc.raw, got, tc.want)
		}
	}
}

// TestAccentSeedTotal verifies Seed() is defined for every named accent
// except AccentDefault, and undefined outside the enum, so the palette
// fallback path is exactly the "no override" set.
func TestAccentSeedTotal(t *testing.T) {
	for a := AccentRed; a <= AccentGraphite; a++ {
		if _, ok := a.Seed(); !ok {
			t.Errorf("Accent(%d).Seed() not defined", a)
		}
	}
	for _, a := range []Accent{AccentDefault, Accent(-1), AccentGraphite + 1} {
		if _, ok := a.Seed(); ok {
			t.Errorf("Accent(%d).Seed() should be undefined", a)
		}
	}
}
