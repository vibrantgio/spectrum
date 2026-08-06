package system_test

// FX.5's poller-count proof, in the shape A3.5 used to measure the defect:
// a counting Source whose completed reads are the observable fact, driven
// at one and at three subscribers. Before the fix every subscription ran
// its own poll loop (A3.5 measured 4 reads for one subscriber against 222
// for three); after it, the read counts at both subscriber counts match
// exactly, because there is exactly one loop.
//
// The sources are gated: Read blocks until the test feeds a value, so a
// "read" happens exactly when the test allows one and the counts are
// deterministic rather than wall-clock samples. Feeding is orchestrated —
// the next value is released only after every subscriber has reported the
// previous one — so conflation can never skip an emission a subscriber is
// still owed.

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/a11y"
	"github.com/vibrantgio/spectrum/system"
)

// gatedSource is a counting Source: Read blocks until a value is fed on
// vals, and reads counts COMPLETED reads — the number of values the poll
// loop actually consumed.
type gatedSource struct {
	vals  chan system.Appearance
	reads atomic.Int32
}

func newGatedSource() *gatedSource {
	return &gatedSource{vals: make(chan system.Appearance)}
}

func (g *gatedSource) Read() (system.Appearance, error) {
	v := <-g.vals
	g.reads.Add(1)
	return v, nil
}

// gatedA11ySource is the a11y twin of gatedSource.
type gatedA11ySource struct {
	vals  chan a11y.A11yPrefs
	reads atomic.Int32
}

func newGatedA11ySource() *gatedA11ySource {
	return &gatedA11ySource{vals: make(chan a11y.A11yPrefs)}
}

func (g *gatedA11ySource) Read() (a11y.A11yPrefs, error) {
	v := <-g.vals
	g.reads.Add(1)
	return v, nil
}

// subscribeCounting subscribes n observers to obs on the concurrent
// scheduler and returns a channel that receives one signal per emission
// per subscriber, plus the subscriptions.
func subscribeCounting[T any](t *testing.T, obs rx.Observable[T], n int) (<-chan struct{}, []rx.Subscription) {
	t.Helper()
	emitted := make(chan struct{}, 64)
	subs := make([]rx.Subscription, 0, n)
	for i := 0; i < n; i++ {
		sub := obs.Subscribe(rx.GoroutineContext(), func(_ T, _ error, done bool) {
			if !done {
				emitted <- struct{}{}
			}
		})
		subs = append(subs, sub)
	}
	return emitted, subs
}

// awaitEmissions fails the test unless n emission signals arrive in time.
func awaitEmissions(t *testing.T, emitted <-chan struct{}, n int, what string) {
	t.Helper()
	for i := 0; i < n; i++ {
		select {
		case <-emitted:
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: got %d of %d expected emissions before timeout "+
				"(a cold, per-subscriber poll loop starves subscribers of a gated source)", what, i, n)
		}
	}
}

// feed hands one value to the single poll loop, failing if no loop is
// there to take it.
func feed[T any](t *testing.T, ch chan<- T, v T, what string) {
	t.Helper()
	select {
	case ch <- v:
	case <-time.After(5 * time.Second):
		t.Fatalf("%s: no poll loop consumed the value within the timeout", what)
	}
}

// appearanceReadsAt drives one FromSource stream with n subscribers through
// a light→dark flip and returns how many source reads that took.
func appearanceReadsAt(t *testing.T, n int) int32 {
	t.Helper()
	src := newGatedSource()
	obs := system.FromSource(src, time.Millisecond)

	emitted, subs := subscribeCounting(t, obs, n)
	defer func() {
		for _, s := range subs {
			s.Unsubscribe()
		}
	}()

	feed(t, src.vals, system.Appearance{Dark: false}, "light value")
	awaitEmissions(t, emitted, n, "light emissions")
	feed(t, src.vals, system.Appearance{Dark: true}, "dark value")
	awaitEmissions(t, emitted, n, "dark emissions")

	return src.reads.Load()
}

// TestFromSourceSharesOnePollLoop is the FX.5 acceptance: the number of
// source reads it takes to deliver a light→dark flip to every subscriber
// is the same at one and at three subscribers — one shared loop, not one
// loop per subscription.
func TestFromSourceSharesOnePollLoop(t *testing.T) {
	one := appearanceReadsAt(t, 1)
	three := appearanceReadsAt(t, 3)
	if one != three {
		t.Fatalf("source reads: %d at one subscriber, %d at three — subscribers are not sharing one poll loop", one, three)
	}
	if one != 2 {
		t.Errorf("source reads at one subscriber = %d, want exactly 2 (one per fed value)", one)
	}
}

func a11yReadsAt(t *testing.T, n int) int32 {
	t.Helper()
	src := newGatedA11ySource()
	obs := a11y.FromSource(src, time.Millisecond)

	emitted, subs := subscribeCounting(t, obs, n)
	defer func() {
		for _, s := range subs {
			s.Unsubscribe()
		}
	}()

	feed(t, src.vals, a11y.A11yPrefs{}, "all-off prefs")
	awaitEmissions(t, emitted, n, "all-off emissions")
	feed(t, src.vals, a11y.A11yPrefs{ReduceMotion: true}, "reduce-motion prefs")
	awaitEmissions(t, emitted, n, "reduce-motion emissions")

	return src.reads.Load()
}

// TestA11yFromSourceSharesOnePollLoop gives spectrum/a11y's same-shaped
// stream the same proof: reads at one and three subscribers match.
func TestA11yFromSourceSharesOnePollLoop(t *testing.T) {
	one := a11yReadsAt(t, 1)
	three := a11yReadsAt(t, 3)
	if one != three {
		t.Fatalf("a11y source reads: %d at one subscriber, %d at three — subscribers are not sharing one poll loop", one, three)
	}
	if one != 2 {
		t.Errorf("a11y source reads at one subscriber = %d, want exactly 2 (one per fed value)", one)
	}
}

// themeReadsAt drives one FromSourceTheme stream — which composes the
// appearance and a11y streams via CombineLatest2 — with n subscribers
// through a light→dark flip, and returns the reads each source took.
func themeReadsAt(t *testing.T, n int) (appearance, prefs int32) {
	t.Helper()
	appSrc := newGatedSource()
	a11ySrc := newGatedA11ySource()
	obs := system.FromSourceTheme(appSrc, time.Millisecond, system.WithA11ySource(a11ySrc))

	emitted, subs := subscribeCounting(t, obs, n)
	defer func() {
		for _, s := range subs {
			s.Unsubscribe()
		}
	}()

	// CombineLatest2 emits once both halves have a value.
	feed(t, appSrc.vals, system.Appearance{Dark: false}, "light value")
	feed(t, a11ySrc.vals, a11y.A11yPrefs{}, "all-off prefs")
	awaitEmissions(t, emitted, n, "light theme emissions")
	feed(t, appSrc.vals, system.Appearance{Dark: true}, "dark value")
	awaitEmissions(t, emitted, n, "dark theme emissions")

	return appSrc.reads.Load(), a11ySrc.reads.Load()
}

// TestFromSourceThemeSharesOnePollLoopPerSource pins the composed stream's
// cost: n theme subscribers produce ONE appearance poller and ONE a11y
// poller, so the read counts on both sources match at one and at three
// subscribers.
func TestFromSourceThemeSharesOnePollLoopPerSource(t *testing.T) {
	app1, prefs1 := themeReadsAt(t, 1)
	app3, prefs3 := themeReadsAt(t, 3)
	if app1 != app3 {
		t.Fatalf("appearance reads: %d at one subscriber, %d at three — theme subscribers are not sharing one appearance poll loop", app1, app3)
	}
	if prefs1 != prefs3 {
		t.Fatalf("a11y reads: %d at one subscriber, %d at three — theme subscribers are not sharing one a11y poll loop", prefs1, prefs3)
	}
	if app1 != 2 {
		t.Errorf("appearance reads at one subscriber = %d, want exactly 2 (one per fed value)", app1)
	}
	if prefs1 != 1 {
		t.Errorf("a11y reads at one subscriber = %d, want exactly 1 (one fed value)", prefs1)
	}
}

// TestFromSourceLateSubscriberReplaysLatest pins the replay half of the
// FX.5 contract: a subscriber that attaches after the first read — the
// second layer of every workbench window — immediately observes the
// current appearance instead of waiting out a poll interval.
func TestFromSourceLateSubscriberReplaysLatest(t *testing.T) {
	src := newGatedSource()
	obs := system.FromSource(src, time.Hour) // one read, then nothing for an hour

	emitted, subs := subscribeCounting(t, obs, 1)
	defer func() {
		for _, s := range subs {
			s.Unsubscribe()
		}
	}()

	want := system.Appearance{Dark: true}
	feed(t, src.vals, want, "dark value")
	awaitEmissions(t, emitted, 1, "first subscriber's emission")

	late := make(chan system.Appearance, 1)
	lateSub := obs.Subscribe(rx.GoroutineContext(), func(a system.Appearance, _ error, done bool) {
		if !done {
			select {
			case late <- a:
			default:
			}
		}
	})
	defer lateSub.Unsubscribe()

	select {
	case got := <-late:
		if got != want {
			t.Fatalf("late subscriber replayed %+v, want %+v", got, want)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("late subscriber received nothing: the shared stream must replay the latest value immediately")
	}
	if got := src.reads.Load(); got != 1 {
		t.Errorf("late subscriber cost %d extra reads, want 0 (reads = %d, want 1)", got-1, got)
	}
}
