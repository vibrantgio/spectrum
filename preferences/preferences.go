// Package preferences persists the user-selected theme name and
// accessibility overrides to an OS-appropriate config directory and emits
// the loaded values on app launch.
//
// The config location is [os.UserConfigDir]:
//   - darwin:  ~/Library/Application Support/<appName>/preferences.json
//   - linux:   $XDG_CONFIG_HOME/<appName>/preferences.json (or ~/.config/...)
//   - windows: %AppData%\<appName>\preferences.json
//
// Preferences are *config*, not *data*: this is why the package uses
// [os.UserConfigDir] (XDG config) rather than gioui's app.DataDir
// (XDG data). This keeps the spectrum module free of a Gio dependency.
package preferences

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/prism/a11y"
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
