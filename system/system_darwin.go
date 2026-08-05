package system

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"
)

// darwinSource reads OS appearance via `defaults read -g`. We deliberately
// use os/exec rather than Cgo + NSUserDefaults: NSUserDefaults caches keys
// in-process, which would require an explicit CFPreferencesAppSynchronize
// before every poll to see external `defaults write` updates. Spawning the
// `defaults` binary always reflects fresh state via cfprefsd, which is what
// the G2.2 acceptance test requires. (This is the asymmetry with prism/a11y,
// where NSWorkspace flags do not have the same staleness problem.)
//
// Cost split (GX.11): each `defaults` call is a fork+exec — measured ~5.5 ms
// each, so the original two-exec Read() was ~11 ms, i.e. ~1.1% CPU at a 1 s
// poll. Dark mode (AppleInterfaceStyle) is the signal a UI must track promptly,
// so it execs on every Read(). The accent (AppleAccentColor) changes rarely, so
// it is re-read at most once per accentInterval and otherwise served from
// cache — halving steady-state exec cost without a CGO notification bridge.
// A worst-case accent change therefore reaches the theme within
// accentInterval plus one poll, not within one poll.
type darwinSource struct {
	accentInterval time.Duration
	now            func() time.Time // injectable clock for tests
	readAccentFn   func() Accent    // injectable accent reader for tests

	mu         sync.Mutex
	accent     Accent
	accentRead bool      // whether accent has ever been read
	accentAt   time.Time // when accent was last read
}

func newDarwinSource() *darwinSource {
	return &darwinSource{
		accentInterval: 10 * time.Second,
		now:            time.Now,
		readAccentFn:   readAccent,
	}
}

func (s *darwinSource) Read() (Appearance, error) {
	return Appearance{
		Dark:   readDark(),
		Accent: s.readAccentThrottled(),
	}, nil
}

// readAccentThrottled execs `defaults read -g AppleAccentColor` at most once
// per accentInterval, serving the cached value in between. The first Read()
// always performs the exec so the initial Appearance is accurate.
func (s *darwinSource) readAccentThrottled() Accent {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	if !s.accentRead || now.Sub(s.accentAt) >= s.accentInterval {
		s.accent = s.readAccentFn()
		s.accentRead = true
		s.accentAt = now
	}
	return s.accent
}

// readDark returns true iff `defaults read -g AppleInterfaceStyle`
// succeeds with a value of "Dark". A missing key (the cfprefsd "does not
// exist" path) means light mode and is not an error.
func readDark() bool {
	out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	if err != nil {
		// Missing key surfaces as ExitError with stderr "does not exist".
		// Any other failure (binary missing, ENOEXEC, etc.) also collapses
		// to "light" rather than producing an error stream — see the
		// FromSource contract.
		var ee *exec.ExitError
		_ = errors.As(err, &ee)
		return false
	}
	return strings.TrimSpace(string(out)) == "Dark"
}

// readAccent reads the AppleAccentColor key and maps it onto the Accent
// enum. A missing key means the user never chose an accent — macOS's
// multicolour default — and folds to AccentDefault, as does a parse
// failure. (An earlier revision returned the raw integer with zero for
// "absent", which conflated "absent" with red; the enum's zero value now
// carries the "no accent" meaning unambiguously.)
func readAccent() Accent {
	out, err := exec.Command("defaults", "read", "-g", "AppleAccentColor").Output()
	if err != nil {
		return AccentDefault
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return AccentDefault
	}
	return accentFromIndex(n)
}

// accentFromIndex maps the raw AppleAccentColor integer onto the Accent
// enum, per the mapping macOS has used since accent colours appeared in
// 10.14 (and unchanged through the multicolour default Big Sur added,
// which is the absent key, handled in readAccent):
//
//	-1 graphite · 0 red · 1 orange · 2 yellow · 3 green · 4 blue ·
//	5 purple · 6 pink
//
// Any value outside that range — nothing ships one today — folds to
// AccentDefault rather than guessing.
func accentFromIndex(n int) Accent {
	switch n {
	case -1:
		return AccentGraphite
	case 0:
		return AccentRed
	case 1:
		return AccentOrange
	case 2:
		return AccentYellow
	case 3:
		return AccentGreen
	case 4:
		return AccentBlue
	case 5:
		return AccentPurple
	case 6:
		return AccentPink
	}
	return AccentDefault
}

func defaultSource() Source { return newDarwinSource() }
