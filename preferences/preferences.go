// Package preferences persists the user's explicit appearance choice — a
// theme name and the accessibility overrides — across launches, as JSON in
// an OS-appropriate config directory.
//
// Reach for it when the application offers its own light/dark/auto control
// and that choice has to survive a restart. [Load] and [Save] take an
// application name and resolve the path themselves; [LoadFrom] and [SaveTo]
// take an explicit path, which is what a test points at a temporary
// directory. The file is:
//
//   - darwin:  ~/Library/Application Support/<appName>/preferences.json
//   - linux:   $XDG_CONFIG_HOME/<appName>/preferences.json (or ~/.config/...)
//   - windows: %AppData%\<appName>\preferences.json
//
// That is [os.UserConfigDir], not gioui's app.DataDir, because preferences
// are config rather than data — and it keeps this module free of a Gio
// dependency.
//
// Two things to know before wiring it up. [Observe] is not a live stream:
// it reads the file once at subscription time, emits, and completes, so a
// later [Save] reaches nobody and a UI that must react to its own writes
// has to publish that change itself. And nothing here turns the stored
// Theme name into a theme value — there is no name-to-theme mapping in this
// module yet, so an application persists a string and is entirely
// responsible for interpreting it, including for the A11y overrides, which
// are recorded and applied by no one.
//
// A missing file is deliberately not an error. Load returns [Default] and a
// nil error, so first launch takes the same code path as every later one;
// the cost is that an unreadable file and a fresh install are only
// distinguishable by the error, never by the value.
package preferences

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/a11y"
)

// Preferences is the persistent user-preference set: a chosen theme name
// and accessibility overrides. All fields are comparable so the value can
// be used with rx.DistinctUntilChanged.
//
// Theme is a free-form name (e.g. "light", "dark", "auto"); the mapping
// from name to a concrete theme.Theme is owned by later spectrum milestones.
// The empty string means "unset" — first-launch state.
type Preferences struct {
	Theme string         `json:"theme"`
	A11y  a11y.A11yPrefs `json:"a11y"`
}

// Default is the first-launch value: empty theme name, all a11y flags off.
var Default = Preferences{}

// Path returns the OS-appropriate preferences file path for the given app
// name. It does not create the file or any parent directories.
func Path(appName string) (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, appName, "preferences.json"), nil
}

// LoadFrom reads preferences from path. A missing file is not an error —
// it returns Default and nil so first-launch code paths are uniform with
// subsequent launches.
func LoadFrom(path string) (Preferences, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, fs.ErrNotExist) {
		return Default, nil
	}
	if err != nil {
		return Default, err
	}
	var p Preferences
	if err := json.Unmarshal(data, &p); err != nil {
		return Default, err
	}
	return p, nil
}

// SaveTo writes preferences to path, creating intermediate directories.
func SaveTo(path string, p Preferences) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Load reads preferences from the OS-appropriate config dir for appName.
func Load(appName string) (Preferences, error) {
	path, err := Path(appName)
	if err != nil {
		return Default, err
	}
	return LoadFrom(path)
}

// Save writes preferences to the OS-appropriate config dir for appName.
func Save(appName string, p Preferences) error {
	path, err := Path(appName)
	if err != nil {
		return err
	}
	return SaveTo(path, p)
}

// Observe returns an Observable that emits the persisted preferences once
// on subscription, then completes. Use this on app launch to seed the UI
// with the user's last-saved choice.
func Observe(appName string) rx.Observable[Preferences] {
	return rx.Defer(func() rx.Observable[Preferences] {
		p, err := Load(appName)
		if err != nil {
			return rx.Throw[Preferences](err)
		}
		return rx.Of(p)
	})
}

// ObserveFrom is the path-based variant of Observe, useful for tests that
// need to point at a temporary directory rather than the OS config dir.
func ObserveFrom(path string) rx.Observable[Preferences] {
	return rx.Defer(func() rx.Observable[Preferences] {
		p, err := LoadFrom(path)
		if err != nil {
			return rx.Throw[Preferences](err)
		}
		return rx.Of(p)
	})
}
