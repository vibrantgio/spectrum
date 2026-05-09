package system_test

import (
	"testing"
	"time"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/prism/tokens"
	"github.com/vibrantgio/spectrum/system"
)

// fakeSource returns successive values from vals on each Read call,
// repeating the last value once the slice is exhausted. Mirrors the
// pattern used in prism/a11y/preferences_test.go.
type fakeSource struct {
	vals []system.Appearance
	n    int
}

func (f *fakeSource) Read() (system.Appearance, error) {
	v := f.vals[f.n]
	if f.n < len(f.vals)-1 {
		f.n++
	}
	return v, nil
}

func collect[T any](obs rx.Observable[T]) ([]T, error) {
	var out []T
	sched := rx.NewScheduler()
	err := obs.Subscribe(func(v T, _ error, done bool) {
		if !done {
			out = append(out, v)
		}
	}, sched).Wait()
	return out, err
}

func TestFromSourceEmitsInitialValue(t *testing.T) {
	want := system.Appearance{Dark: true, AccentIndex: 4}
	src := &fakeSource{vals: []system.Appearance{want}}

	got, err := collect(system.FromSource(src, time.Hour).Take(1))
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

func TestFromSourceEmitsOnDarkChange(t *testing.T) {
	light := system.Appearance{Dark: false}
	dark := system.Appearance{Dark: true}
	src := &fakeSource{vals: []system.Appearance{light, dark}}

	got, err := collect(system.FromSource(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 emissions, got %d", len(got))
	}
	if got[0] != light {
		t.Errorf("first: got %+v, want %+v", got[0], light)
	}
	if got[1] != dark {
		t.Errorf("second: got %+v, want %+v", got[1], dark)
	}
}

func TestFromSourceDeduplicates(t *testing.T) {
	a := system.Appearance{Dark: false}
	b := system.Appearance{Dark: true}
	src := &fakeSource{vals: []system.Appearance{a, a, b}}

	got, err := collect(system.FromSource(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 emissions (a then b), got %d", len(got))
	}
	if got[0] != a || got[1] != b {
		t.Errorf("got %+v then %+v, want %+v then %+v", got[0], got[1], a, b)
	}
}

func TestFromSourceEmitsOnAccentChange(t *testing.T) {
	a := system.Appearance{AccentIndex: 4}
	b := system.Appearance{AccentIndex: 0}
	src := &fakeSource{vals: []system.Appearance{a, b}}

	got, err := collect(system.FromSource(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 emissions, got %d", len(got))
	}
	if got[0].AccentIndex != 4 || got[1].AccentIndex != 0 {
		t.Errorf("accent transitions wrong: %+v then %+v", got[0], got[1])
	}
}

func TestFromSourceThemeBridgesDarkToDarkColors(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{Dark: true}}}

	themes, err := collect(system.FromSourceTheme(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 1 {
		t.Fatalf("expected 1 theme, got %d", len(themes))
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != tokens.DefaultDark {
		t.Errorf("dark appearance must yield DefaultDark; got %+v", colors)
	}
}

func TestFromSourceThemeBridgesLightToLightColors(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{Dark: false}}}

	themes, err := collect(system.FromSourceTheme(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != tokens.DefaultLight {
		t.Errorf("light appearance must yield DefaultLight; got %+v", colors)
	}
}

func TestFromSourceThemeReemitsOnChange(t *testing.T) {
	light := system.Appearance{Dark: false}
	dark := system.Appearance{Dark: true}
	src := &fakeSource{vals: []system.Appearance{light, dark}}

	themes, err := collect(system.FromSourceTheme(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 2 {
		t.Fatalf("expected 2 themes, got %d", len(themes))
	}
	for i, want := range []tokens.ColorTokens{tokens.DefaultLight, tokens.DefaultDark} {
		colors, err := collect(themes[i].Color)
		if err != nil {
			t.Fatalf("theme[%d] color observe: %v", i, err)
		}
		if len(colors) != 1 || colors[0] != want {
			t.Errorf("theme[%d] colors: got %+v, want %+v", i, colors, want)
		}
	}
}
