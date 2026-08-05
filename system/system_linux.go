package system

import (
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

// linuxSource reads the desktop accent colour where a desktop exposes one,
// detected via XDG_CURRENT_DESKTOP (see desktopFromXDG):
//
//   - GNOME 47+: `gsettings get org.gnome.desktop.interface accent-color`,
//     a named enum mapped to libadwaita's published colours
//     (gnomeAccentSeed). Older GNOME has no such key; the gsettings error
//     folds to "no accent".
//   - KDE Plasma: the AccentColor=r,g,b key in [General] of
//     $XDG_CONFIG_HOME/kdeglobals (kdeGlobalsAccent). The key exists only
//     when the user picked an explicit accent; a plain colour scheme folds
//     to "no accent".
//   - Anything else: no accent.
//
// "No accent" leaves Appearance.AccentSeedSet false, and LiveTheme then
// falls back to the default seed's palette. The colour travels as
// Appearance.AccentSeed — Linux accents are arbitrary colours (KDE
// literally so), not the macOS [Accent] enum.
//
// The accent read is throttled exactly like the darwin source's: the GNOME
// path is a fork+exec of `gsettings`, too costly per poll tick, so the
// value is re-read at most once per accentInterval and served from cache
// in between. A worst-case accent change reaches the theme within
// accentInterval plus one poll.
//
// Dark mode is not read yet — Dark stays false. A live implementation
// would ask the org.freedesktop.appearance portal (or gsettings
// color-scheme); that is a later milestone (see the package doc's support
// matrix).
type linuxSource struct {
	accentInterval time.Duration
	now            func() time.Time           // injectable clock for tests
	readAccentFn   func() (color.NRGBA, bool) // injectable accent reader for tests

	mu         sync.Mutex
	seed       color.NRGBA
	seedSet    bool
	accentRead bool      // whether accent has ever been read
	accentAt   time.Time // when accent was last read
}

func newLinuxSource() *linuxSource {
	return &linuxSource{
		accentInterval: 10 * time.Second,
		now:            time.Now,
		readAccentFn:   readLinuxAccent,
	}
}

func (s *linuxSource) Read() (Appearance, error) {
	seed, ok := s.readAccentThrottled()
	return Appearance{
		AccentSeed:    seed,
		AccentSeedSet: ok,
	}, nil
}

// readAccentThrottled invokes the accent reader at most once per
// accentInterval, serving the cached value in between. The first Read()
// always performs the read so the initial Appearance is accurate.
func (s *linuxSource) readAccentThrottled() (color.NRGBA, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if !s.accentRead || now.Sub(s.accentAt) >= s.accentInterval {
		s.seed, s.seedSet = s.readAccentFn()
		s.accentRead = true
		s.accentAt = now
	}
	return s.seed, s.seedSet
}

// readLinuxAccent dispatches on the detected desktop family. Every failure
// along the way — unknown desktop, missing binary, missing key, malformed
// file — folds to "no accent" rather than an error, per the package's
// error contract.
func readLinuxAccent() (color.NRGBA, bool) {
	switch desktopFromXDG(os.Getenv("XDG_CURRENT_DESKTOP")) {
	case "gnome":
		return readGNOMEAccent()
	case "kde":
		return readKDEAccent()
	}
	return color.NRGBA{}, false
}

func readGNOMEAccent() (color.NRGBA, bool) {
	out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "accent-color").Output()
	if err != nil {
		return color.NRGBA{}, false
	}
	return gnomeAccentSeed(string(out))
}

func readKDEAccent() (color.NRGBA, bool) {
	content, err := os.ReadFile(filepath.Join(configHome(), "kdeglobals"))
	if err != nil {
		return color.NRGBA{}, false
	}
	return kdeGlobalsAccent(string(content))
}

// configHome resolves $XDG_CONFIG_HOME with the basedir-spec default of
// ~/.config, which is where KDE keeps kdeglobals.
func configHome() string {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config")
}

func defaultSource() Source { return newLinuxSource() }
