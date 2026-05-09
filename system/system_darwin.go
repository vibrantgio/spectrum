package system

import (
	"errors"
	"os/exec"
	"strconv"
	"strings"
)

// darwinSource reads OS appearance via `defaults read -g`. We deliberately
// use os/exec rather than Cgo + NSUserDefaults: NSUserDefaults caches keys
// in-process, which would require an explicit CFPreferencesAppSynchronize
// before every poll to see external `defaults write` updates. Spawning the
// `defaults` binary is ~50 ms per call but always reflects fresh state via
// cfprefsd, which is what the G2.2 acceptance test requires. (This is the
// asymmetry with prism/a11y, where NSWorkspace flags do not have the same
// staleness problem.)
type darwinSource struct{}

func (darwinSource) Read() (Appearance, error) {
	return Appearance{
		Dark:        readDark(),
		AccentIndex: readAccent(),
	}, nil
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

func defaultSource() Source { return darwinSource{} }
