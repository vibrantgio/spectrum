package preferences_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/prism/a11y"
	"github.com/vibrantgio/spectrum/preferences"
)

func collect[T any](obs rx.Observable[T]) ([]T, error) {
	var out []T
	err := obs.Subscribe(context.Background(), func(v T, _ error, done bool) {
		if !done {
			out = append(out, v)
		}
	}).Wait()
	return out, err
}

// TestPreferencesSurviveRestart is the G2.1 acceptance test: a value saved
// in one "session" is observable in a fresh session that shares no in-memory
// state — only the file on disk. The two LoadFrom calls operate on
// independent Preferences values; nothing crosses the simulated restart
// boundary except the bytes on disk.
func TestPreferencesSurviveRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")

	// Session 1: app starts fresh — no file yet, Default returned without error.
	first, err := preferences.LoadFrom(path)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	if first != preferences.Default {
		t.Fatalf("first load: got %+v, want Default %+v", first, preferences.Default)
	}

	// User picks a theme and toggles every a11y flag, then we persist.
	saved := preferences.Preferences{
		Theme: "dark",
		A11y: a11y.A11yPrefs{
			ReduceMotion:     true,
			HighContrast:     true,
			IncreaseTextSize: true,
		},
	}
	if err := preferences.SaveTo(path, saved); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Session 2: simulated restart — *no* in-memory state from session 1
	// reaches this load. The only path is via the file system.
	second, err := preferences.LoadFrom(path)
	if err != nil {
		t.Fatalf("second load: %v", err)
	}
	if second != saved {
		t.Errorf("post-restart load: got %+v, want %+v", second, saved)
	}
}

// TestPreferencesSurviveRestartViaObserve covers the same acceptance via the
// rx.Observable seam used at app launch — that is, the value emitted by
// Observe on a fresh subscription matches what was previously saved.
func TestPreferencesSurviveRestartViaObserve(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")

	saved := preferences.Preferences{
		Theme: "auto",
		A11y:  a11y.A11yPrefs{ReduceMotion: true},
	}
	if err := preferences.SaveTo(path, saved); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := collect(preferences.ObserveFrom(path))
	if err != nil {
		t.Fatalf("observe: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission on launch, got %d", len(got))
	}
	if got[0] != saved {
		t.Errorf("launch emission: got %+v, want %+v", got[0], saved)
	}
}

func TestLoadMissingFileReturnsDefault(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.json")

	got, err := preferences.LoadFrom(path)
	if err != nil {
		t.Fatalf("missing file should not error, got %v", err)
	}
	if got != preferences.Default {
		t.Errorf("missing file: got %+v, want Default %+v", got, preferences.Default)
	}
}

func TestSaveCreatesIntermediateDirs(t *testing.T) {
	// The config dir under os.UserConfigDir typically does not exist on first
	// launch — Save must create it.
	path := filepath.Join(t.TempDir(), "fresh", "nested", "preferences.json")

	if err := preferences.SaveTo(path, preferences.Preferences{Theme: "light"}); err != nil {
		t.Fatalf("save into fresh dir: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file to exist at %s: %v", path, err)
	}
}

func TestPathUsesOSConfigDir(t *testing.T) {
	got, err := preferences.Path("vibrantgio-test")
	if err != nil {
		t.Fatalf("Path: %v", err)
	}
	want, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("UserConfigDir: %v", err)
	}
	rel, err := filepath.Rel(want, got)
	if err != nil || rel == "" || rel[:2] == ".." {
		t.Errorf("Path %q is not under UserConfigDir %q", got, want)
	}
	if filepath.Base(got) != "preferences.json" {
		t.Errorf("Path basename: got %q, want preferences.json", filepath.Base(got))
	}
}

func TestSaveOverwrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "preferences.json")

	if err := preferences.SaveTo(path, preferences.Preferences{Theme: "light"}); err != nil {
		t.Fatalf("first save: %v", err)
	}
	want := preferences.Preferences{Theme: "dark", A11y: a11y.A11yPrefs{HighContrast: true}}
	if err := preferences.SaveTo(path, want); err != nil {
		t.Fatalf("second save: %v", err)
	}
	got, err := preferences.LoadFrom(path)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got != want {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
