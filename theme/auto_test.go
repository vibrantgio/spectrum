package theme_test

import (
	"testing"
	"time"

	"github.com/vibrantgio/spectrum/theme"
	"github.com/vibrantgio/spectrum/tokens"
)

func TestAutoLightDarkNonNil(t *testing.T) {
	if theme.AutoLightDark() == nil {
		t.Fatal("AutoLightDark returned nil")
	}
}

func TestAutoLightDarkEmitsOneTheme(t *testing.T) {
	got, err := collect(theme.AutoLightDark().Take(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	th := got[0]
	if th.Color == nil {
		t.Error("emitted Theme.Color is nil")
	}
	if th.Motion == nil {
		t.Error("emitted Theme.Motion is nil")
	}
	if th.Spacing == nil {
		t.Error("emitted Theme.Spacing is nil")
	}
	if th.Radius == nil {
		t.Error("emitted Theme.Radius is nil")
	}
	if th.Elevation == nil {
		t.Error("emitted Theme.Elevation is nil")
	}
}

func TestAutoLightDarkColorTokens(t *testing.T) {
	themes, err := collect(theme.AutoLightDark().Take(1))
	if err != nil || len(themes) != 1 {
		t.Fatalf("setup failed: err=%v len=%d", err, len(themes))
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observable error: %v", err)
	}
	if len(colors) != 1 {
		t.Fatalf("expected 1 color emission, got %d", len(colors))
	}
	h := time.Now().Hour()
	var want tokens.ColorTokens
	if h <= 6 || h >= 18 {
		want = tokens.DefaultDark
	} else {
		want = tokens.DefaultLight
	}
	if colors[0] != want {
		t.Errorf("wrong colour scheme for hour %d: got background %v, want background %v",
			h, colors[0].Background, want.Background)
	}
}
