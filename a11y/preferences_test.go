package a11y_test

import (
	"context"
	"testing"
	"time"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/a11y"
)

// fakeSource returns successive values from vals on each Read call,
// repeating the last value once the slice is exhausted.
type fakeSource struct {
	vals []a11y.A11yPrefs
	n    int
}

func (f *fakeSource) Read() (a11y.A11yPrefs, error) {
	v := f.vals[f.n]
	if f.n < len(f.vals)-1 {
		f.n++
	}
	return v, nil
}

func collect[T any](obs rx.Observable[T]) ([]T, error) {
	var out []T
	err := obs.Subscribe(context.Background(), func(v T, err error, done bool) {
		if !done {
			out = append(out, v)
		}
	}).Wait()
	return out, err
}

func TestFromSourceEmitsInitialValue(t *testing.T) {
	want := a11y.A11yPrefs{ReduceMotion: true, HighContrast: false, IncreaseTextSize: false}
	src := &fakeSource{vals: []a11y.A11yPrefs{want}}

	got, err := collect(a11y.FromSource(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	if got[0] != want {
		t.Errorf("got %+v, want %+v", got[0], want)
	}
}

func TestFromSourceAllPrefsTrue(t *testing.T) {
	want := a11y.A11yPrefs{ReduceMotion: true, HighContrast: true, IncreaseTextSize: true}
	src := &fakeSource{vals: []a11y.A11yPrefs{want}}

	got, err := collect(a11y.FromSource(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	if got[0] != want {
		t.Errorf("got %+v, want %+v", got[0], want)
	}
}

func TestFromSourceEmitsOnChange(t *testing.T) {
	a := a11y.A11yPrefs{ReduceMotion: false}
	b := a11y.A11yPrefs{ReduceMotion: true}
	src := &fakeSource{vals: []a11y.A11yPrefs{a, b}}

	got, err := collect(a11y.FromSource(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 emissions, got %d", len(got))
	}
	if got[0] != a {
		t.Errorf("first emission: got %+v, want %+v", got[0], a)
	}
	if got[1] != b {
		t.Errorf("second emission: got %+v, want %+v", got[1], b)
	}
}

func TestFromSourceDeduplicates(t *testing.T) {
	// Source returns [a, a, b]: the repeated a must be suppressed.
	a := a11y.A11yPrefs{ReduceMotion: false}
	b := a11y.A11yPrefs{HighContrast: true}
	src := &fakeSource{vals: []a11y.A11yPrefs{a, a, b}}

	// Take(2) should collect [a, b] — the duplicate a is filtered.
	got, err := collect(a11y.FromSource(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 emissions (a then b, no duplicate), got %d", len(got))
	}
	if got[0] != a {
		t.Errorf("first emission: got %+v, want %+v", got[0], a)
	}
	if got[1] != b {
		t.Errorf("second emission: got %+v, want %+v", got[1], b)
	}
}

func TestFromSourceHighContrastToggle(t *testing.T) {
	off := a11y.A11yPrefs{HighContrast: false}
	on := a11y.A11yPrefs{HighContrast: true}
	src := &fakeSource{vals: []a11y.A11yPrefs{off, on}}

	got, err := collect(a11y.FromSource(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 emissions, got %d", len(got))
	}
	if got[0].HighContrast {
		t.Error("first emission: HighContrast should be false")
	}
	if !got[1].HighContrast {
		t.Error("second emission: HighContrast should be true")
	}
}
