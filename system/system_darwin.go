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
// so it execs on every Read(). The accent (AppleAccentColor) changes rarely and
// no consumer yet maps it to a colour, so it is re-read at most once per
// accentInterval and otherwise served from cache — halving steady-state exec
// cost without a CGO notification bridge.
type darwinSource struct {
	accentInterval time.Duration
	now            func() time.Time // injectable clock for tests
	readAccentFn   func() int       // injectable accent reader for tests

	mu         sync.Mutex
	accent     int
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
		Dark:        readDark(),
		AccentIndex: s.readAccentThrottled(),
	}, nil
}

// readAccentThrottled execs `defaults read -g AppleAccentColor` at most once
// per accentInterval, serving the cached value in between. The first Read()
// always performs the exec so the initial Appearance is accurate.
func (s *darwinSource) readAccentThrottled() int {
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

// readAccent returns the AppleAccentColor key as an integer (-1..7).
// Zero is returned both when the key is missing and when parsing fails;
// callers that need to distinguish "explicit zero" from "absent" should
// not rely on AccentIndex alone.
func readAccent() int {
	out, err := exec.Command("defaults", "read", "-g", "AppleAccentColor").Output()
	if err != nil {
		return 0
	}
	n, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0
	}
	return n
}

func defaultSource() Source { return newDarwinSource() }
