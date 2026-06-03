package system

import (
	"os/exec"
	"testing"
	"time"
)

// BenchmarkDarwinDefaultsExec documents the cost that motivates GX.11's cadence
// split: each `defaults read -g` is a fork+exec. Run with:
//
//	go test -bench=DarwinDefaultsExec -benchtime=200x ./spectrum/system/...
//
// On the development machine this reports ~5.5 ms/op for a single key, so the
// original two-exec Read() was ~11 ms — i.e. ~1.1% wall-clock CPU at a 1 s poll,
// ~0.22% at 5 s. (FEEDBACK-G5.1's unmeasured "10% at 1 s" was ~9× high.) The
// throttled accent path cuts steady-state execs from two per tick to ~one.
func BenchmarkDarwinDefaultsExec(b *testing.B) {
	if _, err := exec.LookPath("defaults"); err != nil {
		b.Skipf("defaults binary unavailable: %v", err)
	}
	b.ReportAllocs()
	for b.Loop() {
		_ = readDark() // one fork+exec of `defaults read -g AppleInterfaceStyle`
	}
}

// TestDarwinAccentThrottledBelowDark verifies the GX.11 cadence split: the
// accent (AppleAccentColor) reader is invoked far less often than Read() —
// at most once per accentInterval — while every Read() still reflects dark
// mode promptly (dark is read unconditionally, exercised by the real
// acceptance test). Uses an injected clock and a counting accent reader so the
// assertion is deterministic and performs no fork+exec.
func TestDarwinAccentThrottledBelowDark(t *testing.T) {
	var clock time.Time
	accentReads := 0

	src := &darwinSource{
		accentInterval: 10 * time.Second,
		now:            func() time.Time { return clock },
		readAccentFn: func() int {
			accentReads++
			return 7
		},
	}

	// Simulate a 1 s poll over 60 s: 60 Read() calls. With a 10 s accent
	// interval, accent should be read on the first call and then once per
	// 10 s boundary — i.e. far fewer than 60 times.
	const polls = 60
	for i := range polls {
		a, err := src.Read()
		if err != nil {
			t.Fatalf("Read() %d: %v", i, err)
		}
		if a.AccentIndex != 7 {
			t.Fatalf("Read() %d: AccentIndex=%d; want 7 (cached or fresh)", i, a.AccentIndex)
		}
		clock = clock.Add(time.Second)
	}

	// Over 60 s at a 10 s interval: reads at t=0,10,20,30,40,50 → 6 reads.
	// Assert it is at most that, and dramatically below the poll count.
	if accentReads > 6 {
		t.Errorf("accent read %d times over %d polls; want ≤ 6 (once per 10 s)", accentReads, polls)
	}
	if accentReads >= polls {
		t.Fatalf("accent not throttled: read %d times for %d polls", accentReads, polls)
	}
	t.Logf("accent read %d times over %d one-second polls (throttle working)", accentReads, polls)
}

// TestDarwinAccentReadOnFirstCall verifies the first Read() always reads the
// accent (no stale zero before the first interval elapses).
func TestDarwinAccentReadOnFirstCall(t *testing.T) {
	var clock time.Time
	accentReads := 0
	src := &darwinSource{
		accentInterval: 10 * time.Second,
		now:            func() time.Time { return clock },
		readAccentFn: func() int {
			accentReads++
			return 3
		},
	}
	a, err := src.Read()
	if err != nil {
		t.Fatalf("Read(): %v", err)
	}
	if accentReads != 1 {
		t.Fatalf("first Read() performed %d accent reads; want exactly 1", accentReads)
	}
	if a.AccentIndex != 3 {
		t.Errorf("first Read() AccentIndex=%d; want 3", a.AccentIndex)
	}
}

// TestDarwinAccentRefreshesAfterInterval verifies the cache refreshes once the
// interval has elapsed, so a genuine accent change is eventually observed.
func TestDarwinAccentRefreshesAfterInterval(t *testing.T) {
	var clock time.Time
	value := 1
	src := &darwinSource{
		accentInterval: 10 * time.Second,
		now:            func() time.Time { return clock },
		readAccentFn:   func() int { return value },
	}

	if a, _ := src.Read(); a.AccentIndex != 1 {
		t.Fatalf("initial accent=%d; want 1", a.AccentIndex)
	}
	// Change the underlying value; before the interval elapses the cache holds.
	value = 5
	clock = clock.Add(9 * time.Second)
	if a, _ := src.Read(); a.AccentIndex != 1 {
		t.Errorf("accent before interval=%d; want cached 1", a.AccentIndex)
	}
	// After the interval, the new value is picked up.
	clock = clock.Add(2 * time.Second) // total 11 s ≥ 10 s
	if a, _ := src.Read(); a.AccentIndex != 5 {
		t.Errorf("accent after interval=%d; want refreshed 5", a.AccentIndex)
	}
}
