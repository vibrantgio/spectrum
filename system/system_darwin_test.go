package system_test

import (
	"errors"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/system"
)

// TestDarkModeFlipEmitsWithinOneSecond is the G2.2 acceptance test.
// It exercises the real macOS source against a real `defaults write`,
// asserting that the Live() observable surfaces the change within the
// 1-second budget named in PLAN.md G2.2 Measurable.
//
// The test mutates the user's actual NSGlobalDomain — there is no way to
// satisfy the Measurable verbatim ("defaults write -g AppleInterfaceStyle
// Dark") with a sandboxed domain. The user's prior setting is captured up
// front and restored via t.Cleanup, even on failure.
func TestDarkModeFlipEmitsWithinOneSecond(t *testing.T) {
	if _, err := exec.LookPath("defaults"); err != nil {
		t.Skipf("defaults binary unavailable: %v", err)
	}

	restore := captureAppleInterfaceStyle(t)
	t.Cleanup(restore)

	// Establish a known starting point: Light (key absent).
	if err := setAppleInterfaceStyleLight(); err != nil {
		t.Fatalf("setup: clear AppleInterfaceStyle: %v", err)
	}

	// 100 ms poll interval → emissions land within ~100–200 ms of cfprefsd
	// updating, with comfortable margin under the 1 s acceptance budget.
	const pollInterval = 100 * time.Millisecond
	obs := system.Live(pollInterval)

	emissions := make(chan system.Appearance, 16)
	sub := obs.Subscribe(rx.GoroutineContext(), func(a system.Appearance, _ error, done bool) {
		if !done {
			select {
			case emissions <- a:
			default:
				// Channel buffer is large; dropping here would mask a
				// runaway emitter. Failing fast is the right signal.
				panic("emissions buffer overflow — Live emitting faster than test consumes")
			}
		}
	})
	t.Cleanup(sub.Unsubscribe)

	// Drain the initial Light emission before flipping.
	select {
	case got := <-emissions:
		if got.Dark {
			t.Fatalf("initial emission unexpectedly Dark: %+v", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("never received initial Light emission from Live()")
	}

	// The acceptance deadline starts the moment we issue `defaults write`.
	flipStart := time.Now()
	if err := setAppleInterfaceStyleDark(); err != nil {
		t.Fatalf("defaults write Dark: %v", err)
	}
	deadline := flipStart.Add(time.Second)

	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			t.Fatalf("no Dark emission within 1 s of `defaults write`")
		}
		select {
		case got := <-emissions:
			if got.Dark {
				t.Logf("Dark emission %v after defaults write", time.Since(flipStart))
				return
			}
			// A residual Light emission can sneak in if the poll fired
			// before cfprefsd settled; keep waiting.
		case <-time.After(remaining):
			t.Fatalf("no Dark emission within 1 s of `defaults write`")
		}
	}
}

// captureAppleInterfaceStyle reads the current AppleInterfaceStyle key and
// returns a function that restores the original state. If the key was
// absent, the restore callback deletes whatever the test wrote.
func captureAppleInterfaceStyle(t *testing.T) func() {
	t.Helper()
	out, err := exec.Command("defaults", "read", "-g", "AppleInterfaceStyle").Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			// Key absent → restore by deleting whatever the test sets.
			return func() { _ = setAppleInterfaceStyleLight() }
		}
		t.Fatalf("capture state: %v", err)
		return func() {}
	}
	prev := strings.TrimSpace(string(out))
	return func() {
		// Best-effort restore; if it fails the user can re-toggle from
		// System Settings — but we still log so the test report is honest.
		if err := exec.Command("defaults", "write", "-g", "AppleInterfaceStyle", prev).Run(); err != nil {
			t.Logf("warning: failed to restore AppleInterfaceStyle=%s: %v", prev, err)
		}
	}
}

func setAppleInterfaceStyleLight() error {
	// `defaults delete` returns nonzero if the key is already absent.
	// That is not an error from the test's point of view — the
	// post-condition we want is "key absent".
	cmd := exec.Command("defaults", "delete", "-g", "AppleInterfaceStyle")
	if err := cmd.Run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil
		}
		return err
	}
	return nil
}

func setAppleInterfaceStyleDark() error {
	return exec.Command("defaults", "write", "-g", "AppleInterfaceStyle", "Dark").Run()
}
