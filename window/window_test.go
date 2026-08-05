package window_test

import (
	"context"
	"testing"

	"gioui.org/layout"

	"github.com/reactivego/rx"

	"github.com/vibrantgio/spectrum/theme"
	"github.com/vibrantgio/spectrum/tokens"
	"github.com/vibrantgio/spectrum/window"
)

// collect subscribes to obs synchronously and returns all emitted values.
// Mirrors the helper used elsewhere in the spectrum/prism test suites.
func collect[T any](obs rx.Observable[T]) ([]T, error) {
	var out []T
	err := obs.Subscribe(context.Background(), func(v T, _ error, done bool) {
		if !done {
			out = append(out, v)
		}
	}).Wait()
	return out, err
}

// lightTheme and darkTheme are the two distinct themes the isolation
// test launches its windows with. They differ in their Color stream's
// emitted ColorTokens (DefaultLight vs DefaultDark), which is the
// observable property the test asserts each window keeps to itself.
func lightTheme() theme.Theme {
	return theme.Theme{
		Color:     rx.Of(tokens.DefaultLight),
		Type:      rx.Of(tokens.DefaultTypeScale),
		Motion:    rx.Of(tokens.Motion),
		Spacing:   rx.Of(tokens.Spacing),
		Radius:    rx.Of(tokens.Radius),
		Elevation: rx.Of(tokens.Elevation),
	}
}

func darkTheme() theme.Theme {
	return theme.Theme{
		Color:     rx.Of(tokens.DefaultDark),
		Type:      rx.Of(tokens.DefaultTypeScale),
		Motion:    rx.Of(tokens.Motion),
		Spacing:   rx.Of(tokens.Spacing),
		Radius:    rx.Of(tokens.Radius),
		Elevation: rx.Of(tokens.Elevation),
	}
}

// TestPerWindowThemeIsolation discharges G2.4 Measurable: two windows
// constructed with different themes maintain isolated theme streams. The
// build callback handed to each [window.Window] sees only that window's
// theme; the colour tokens consumed by window A's layers are exactly
// the ones threaded into window A and never the ones threaded into B.
//
// The test does not invoke [Window.Render] — that would call
// [mvu.Window.Render] which blocks on Gio's OS event loop and requires
// app.Main() plus a real display. The contract under test is the
// theme→layers binding the wrapper enforces, not the Gio loop. Two
// [window.Window] instances are "launched" in the contract sense: each
// is constructed with its own theme observable and the build callback
// captures what it sees.
func TestPerWindowThemeIsolation(t *testing.T) {
	// The mvu.Window pointer is unused in this test. Passing nil makes
	// the dependency on the event loop explicit — the build callback
	// never reads it, and the wrapper's Render method (which would) is
	// not exercised here.
	winA := window.New(nil, rx.Of(lightTheme()))
	winB := window.New(nil, rx.Of(darkTheme()))

	// Each build callback captures the colour tokens its window's theme
	// emits. The capture happens through the layer pipeline a real app
	// would build (theme → Color → layout.Widget), so the assertion is
	// that the wrapper threaded the right theme into the right window.
	var capturedA, capturedB []tokens.ColorTokens
	captureColors := func(dst *[]tokens.ColorTokens) func(rx.Observable[theme.Theme]) []rx.Observable[layout.Widget] {
		return func(themeObs rx.Observable[theme.Theme]) []rx.Observable[layout.Widget] {
			return []rx.Observable[layout.Widget]{
				rx.SwitchMap(themeObs, func(th theme.Theme) rx.Observable[layout.Widget] {
					return rx.Map(th.Color, func(c tokens.ColorTokens) layout.Widget {
						*dst = append(*dst, c)
						return func(layout.Context) layout.Dimensions {
							return layout.Dimensions{}
						}
					})
				}),
			}
		}
	}

	// build(w.Theme) is what (*Window).Render does internally. Calling
	// it directly here lets us drive the layer observables to completion
	// without entering mvu's event loop.
	layersA := captureColors(&capturedA)(winA.Theme)
	layersB := captureColors(&capturedB)(winB.Theme)

	if _, err := collect(layersA[0]); err != nil {
		t.Fatalf("window A layer subscribe: %v", err)
	}
	if _, err := collect(layersB[0]); err != nil {
		t.Fatalf("window B layer subscribe: %v", err)
	}

	if len(capturedA) != 1 || capturedA[0] != tokens.DefaultLight {
		t.Errorf("window A captured colours: got %+v, want [DefaultLight]", capturedA)
	}
	if len(capturedB) != 1 || capturedB[0] != tokens.DefaultDark {
		t.Errorf("window B captured colours: got %+v, want [DefaultDark]", capturedB)
	}

	// Cross-window isolation: A must never have seen B's colour, and
	// vice versa. This is the property "different themes... isolation"
	// in the milestone Measurable.
	for _, c := range capturedA {
		if c == tokens.DefaultDark {
			t.Errorf("window A leaked window B's DefaultDark colour: %+v", capturedA)
		}
	}
	for _, c := range capturedB {
		if c == tokens.DefaultLight {
			t.Errorf("window B leaked window A's DefaultLight colour: %+v", capturedB)
		}
	}
}

// TestThemeFieldIsTheArgument is a regression guard for the wrapper
// contract: New must store the supplied observable on the Theme field
// without substitution. If a future refactor wraps or transforms the
// argument, the per-window override claim breaks because two windows
// constructed with the same upstream observable would no longer share
// it (or, worse, two windows with different observables would share a
// transformed cache).
func TestThemeFieldIsTheArgument(t *testing.T) {
	want := rx.Of(lightTheme())
	w := window.New(nil, want)
	// rx.Observable is a function type; comparing function identities
	// across the boundary would require reflect, but a value-level
	// equality check is enough here: emitting from `want` and from
	// `w.Theme` must yield the same single emission.
	gotThemes, err := collect(w.Theme)
	if err != nil {
		t.Fatalf("collect Theme: %v", err)
	}
	wantThemes, err := collect(want)
	if err != nil {
		t.Fatalf("collect want: %v", err)
	}
	if len(gotThemes) != len(wantThemes) || len(gotThemes) != 1 {
		t.Fatalf("emission count mismatch: w.Theme=%d, want=%d", len(gotThemes), len(wantThemes))
	}
	gotColors, err := collect(gotThemes[0].Color)
	if err != nil {
		t.Fatalf("collect Color: %v", err)
	}
	if len(gotColors) != 1 || gotColors[0] != tokens.DefaultLight {
		t.Errorf("Theme passed through New differs from input: got %+v", gotColors)
	}
}
