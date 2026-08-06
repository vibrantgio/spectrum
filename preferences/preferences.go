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
// Two things to know before wiring it up. [Observe] is a live stream since
// FX.5: it emits the persisted value on subscription and then re-emits on
// every [Save]/[SaveTo] to the same path from this process — it never
// completes, so unsubscribe (or Take) when done, and note that writes made
// by other processes or by hand are NOT observed; there is no file watcher,
// only the in-process Save notification. And nothing here turns the stored
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
	"sync"

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

// SaveTo writes preferences to path, creating intermediate directories,
// and notifies every live [Observe]/[ObserveFrom] subscription on the same
// path (in this process) of the new value. Nothing is notified when the
// write fails.
func SaveTo(path string, p Preferences) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	notify(path, p)
	return nil
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

// Observe returns an Observable that emits the persisted preferences on
// subscription and then re-emits on every [Save] to the same app name from
// this process (FX.5: it used to complete after the one read; now Save and
// Observe agree). Consecutive duplicate values are collapsed. The stream
// never completes — unsubscribe (or Take) when done — and it does not
// watch the file: a write from another process or editor is not observed.
//
// Use it to seed the UI with the user's last-saved choice at launch and to
// keep every consumer of the choice current when the settings screen saves
// a new one.
func Observe(appName string) rx.Observable[Preferences] {
	return rx.Defer(func() rx.Observable[Preferences] {
		path, err := Path(appName)
		if err != nil {
			return rx.Throw[Preferences](err)
		}
		return observePath(path)
	})
}

// ObserveFrom is the path-based variant of Observe, useful for tests that
// need to point at a temporary directory rather than the OS config dir. It
// carries the same contract: initial value, then every SaveTo to the same
// path, never completing.
func ObserveFrom(path string) rx.Observable[Preferences] {
	return rx.Defer(func() rx.Observable[Preferences] {
		return observePath(path)
	})
}

// streams is the in-process registry behind the emit-on-write contract:
// one multicast per (cleaned) preferences path, shared by every Observe
// subscription and fed by every successful Save.
var streams struct {
	sync.Mutex
	byPath map[string]stream
}

// stream is one path's live multicast: send feeds it (Save), obs replays
// the latest value to a new subscriber and then follows (Observe).
type stream struct {
	send rx.Observer[Preferences]
	obs  rx.Observable[Preferences]
}

// streamFor returns path's multicast, creating it — seeded with the value
// currently on disk — on first use. The registry key is the cleaned path,
// so Save and Observe spellings of the same file meet the same stream.
func streamFor(path string) (stream, error) {
	key := filepath.Clean(path)
	streams.Lock()
	defer streams.Unlock()
	if s, ok := streams.byPath[key]; ok {
		return s, nil
	}
	p, err := LoadFrom(path)
	if err != nil {
		return stream{}, err
	}
	// Replay depth 1 (the current value), with buffer headroom so a burst
	// of saves never blocks the saver on a slow subscriber.
	send, obs := rx.Subject[Preferences](0, 1, 128)
	send(p, nil, false)
	s := stream{send: send, obs: obs}
	if streams.byPath == nil {
		streams.byPath = make(map[string]stream)
	}
	streams.byPath[key] = s
	return s, nil
}

// notify feeds a successfully saved value to path's live stream, if one
// exists. With no observers there is nothing to do: the next streamFor
// reads the file fresh.
func notify(path string, p Preferences) {
	key := filepath.Clean(path)
	streams.Lock()
	s, ok := streams.byPath[key]
	streams.Unlock()
	if ok {
		s.send(p, nil, false)
	}
}

// observePath is the shared Observe/ObserveFrom body: the path's multicast
// with consecutive duplicates collapsed.
func observePath(path string) rx.Observable[Preferences] {
	s, err := streamFor(path)
	if err != nil {
		return rx.Throw[Preferences](err)
	}
	return s.obs.DistinctUntilChanged(rx.Equal[Preferences]())
}
