package system

import (
	"image/color"
	"testing"
	"time"
)

var testSeed = color.NRGBA{R: 0x35, G: 0x84, B: 0xE4, A: 0xFF}

// TestLinuxAccentThrottled mirrors the darwin throttle contract: the accent
// reader — a fork+exec of gsettings on GNOME — is invoked at most once per
// accentInterval, not once per Read. Uses an injected clock and a counting
// reader so the assertion is deterministic and performs no exec.
func TestLinuxAccentThrottled(t *testing.T) {
	var clock time.Time
	accentReads := 0

	src := &linuxSource{
		accentInterval: 10 * time.Second,
		now:            func() time.Time { return clock },
		readAccentFn: func() (color.NRGBA, bool) {
			accentReads++
			return testSeed, true
		},
	}

	const polls = 60
	for i := range polls {
		a, err := src.Read()
		if err != nil {
			t.Fatalf("Read() %d: %v", i, err)
		}
		if !a.AccentSeedSet || a.AccentSeed != testSeed {
			t.Fatalf("Read() %d: AccentSeed=%+v set=%v; want cached or fresh seed", i, a.AccentSeed, a.AccentSeedSet)
		}
		clock = clock.Add(time.Second)
	}

	// Over 60 s at a 10 s interval: reads at t=0,10,20,30,40,50 → 6 reads.
	if accentReads > 6 {
		t.Errorf("accent read %d times over %d polls; want ≤ 6 (once per 10 s)", accentReads, polls)
	}
	if accentReads >= polls {
		t.Fatalf("accent not throttled: read %d times for %d polls", accentReads, polls)
	}
}

// TestLinuxAccentReadOnFirstCall verifies the first Read() always reads the
// accent (no stale zero before the first interval elapses).
func TestLinuxAccentReadOnFirstCall(t *testing.T) {
	var clock time.Time
	accentReads := 0
	src := &linuxSource{
		accentInterval: 10 * time.Second,
		now:            func() time.Time { return clock },
		readAccentFn: func() (color.NRGBA, bool) {
			accentReads++
			return testSeed, true
		},
	}
	a, err := src.Read()
	if err != nil {
		t.Fatalf("Read(): %v", err)
	}
	if accentReads != 1 {
		t.Fatalf("first Read() performed %d accent reads; want exactly 1", accentReads)
	}
	if !a.AccentSeedSet || a.AccentSeed != testSeed {
		t.Errorf("first Read() AccentSeed=%+v set=%v; want the injected seed", a.AccentSeed, a.AccentSeedSet)
	}
}

// TestLinuxAccentRefreshesAfterInterval verifies the cache refreshes once
// the interval has elapsed — including a transition to "no accent", so a
// user clearing the accent is eventually observed.
func TestLinuxAccentRefreshesAfterInterval(t *testing.T) {
	var clock time.Time
	seed, set := testSeed, true
	src := &linuxSource{
		accentInterval: 10 * time.Second,
		now:            func() time.Time { return clock },
		readAccentFn:   func() (color.NRGBA, bool) { return seed, set },
	}

	if a, _ := src.Read(); !a.AccentSeedSet || a.AccentSeed != testSeed {
		t.Fatalf("initial accent=%+v set=%v; want the seed", a.AccentSeed, a.AccentSeedSet)
	}
	// Clear the underlying accent; before the interval elapses the cache holds.
	seed, set = color.NRGBA{}, false
	clock = clock.Add(9 * time.Second)
	if a, _ := src.Read(); !a.AccentSeedSet {
		t.Errorf("accent before interval cleared; want cached seed")
	}
	// After the interval, the cleared state is picked up.
	clock = clock.Add(2 * time.Second) // total 11 s ≥ 10 s
	if a, _ := src.Read(); a.AccentSeedSet {
		t.Errorf("accent after interval still set; want no accent")
	}
}

// TestLinuxSourceNoAccent verifies the zero path: a reader that finds no
// accent yields an Appearance whose AccentSeedSet is false and whose Dark
// is false (dark mode is not read on Linux yet).
func TestLinuxSourceNoAccent(t *testing.T) {
	src := &linuxSource{
		accentInterval: 10 * time.Second,
		now:            time.Now,
		readAccentFn:   func() (color.NRGBA, bool) { return color.NRGBA{}, false },
	}
	a, err := src.Read()
	if err != nil {
		t.Fatalf("Read(): %v", err)
	}
	if a != (Appearance{}) {
		t.Errorf("no-accent Read() = %+v; want the zero Appearance", a)
	}
}
