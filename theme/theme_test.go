package theme_test

// rx.TestScheduler does not exist in github.com/reactivego/rx; tests rely on
// Subscribe's default serial trampoline scheduler (context without a scheduler
// attached), which provides the deterministic, synchronous execution needed to
// verify emission shape.

import (
	"context"
	"testing"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/theme"
	"github.com/vibrantgio/spectrum/tokens"
)

// collect subscribes to obs synchronously and returns all emitted values.
func collect[T any](obs rx.Observable[T]) ([]T, error) {
	var out []T
	err := obs.Subscribe(context.Background(), func(v T, err error, done bool) {
		if !done {
			out = append(out, v)
		}
	}).Wait()
	return out, err
}

func TestDefaultThemeFieldsNonNil(t *testing.T) {
	th := theme.Default()
	if th.Color == nil {
		t.Error("Color is nil")
	}
	if th.Type == nil {
		t.Error("Type is nil")
	}
	if th.Motion == nil {
		t.Error("Motion is nil")
	}
	if th.Spacing == nil {
		t.Error("Spacing is nil")
	}
	if th.Radius == nil {
		t.Error("Radius is nil")
	}
	if th.Elevation == nil {
		t.Error("Elevation is nil")
	}
}

func TestDefaultColorEmission(t *testing.T) {
	th := theme.Default()
	got, err := collect(th.Color)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	if got[0] != tokens.DefaultLight {
		t.Error("emitted ColorTokens does not match tokens.DefaultLight")
	}
}

func TestDefaultTypeEmission(t *testing.T) {
	th := theme.Default()
	got, err := collect(th.Type)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	if got[0] != tokens.DefaultTypeScale {
		t.Error("emitted TypeScale does not match tokens.DefaultTypeScale")
	}
}

func TestDefaultMotionEmission(t *testing.T) {
	th := theme.Default()
	got, err := collect(th.Motion)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	if got[0] != tokens.Motion {
		t.Error("emitted MotionScale does not match tokens.Motion")
	}
}

func TestDefaultSpacingEmission(t *testing.T) {
	th := theme.Default()
	got, err := collect(th.Spacing)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	if got[0] != tokens.Spacing {
		t.Error("emitted SpacingScale does not match tokens.Spacing")
	}
}

func TestDefaultRadiusEmission(t *testing.T) {
	th := theme.Default()
	got, err := collect(th.Radius)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	if got[0] != tokens.Radius {
		t.Error("emitted RadiusScale does not match tokens.Radius")
	}
}

func TestDefaultElevationEmission(t *testing.T) {
	th := theme.Default()
	got, err := collect(th.Elevation)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	if got[0] != tokens.Elevation {
		t.Error("emitted ElevationScale does not match tokens.Elevation")
	}
}
