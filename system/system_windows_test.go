package system

import (
	"image/color"
	"testing"
)

// TestWindowsSourceCarriesAccentSeed verifies the source glue: a registry
// read that yields a colour lands in Appearance.AccentSeed with the set
// flag raised, and Dark stays false (dark mode is not read on Windows yet).
func TestWindowsSourceCarriesAccentSeed(t *testing.T) {
	want := color.NRGBA{R: 0x00, G: 0x78, B: 0xD7, A: 0xFF}
	src := &windowsSource{
		readAccentFn: func() (color.NRGBA, bool) { return want, true },
	}
	a, err := src.Read()
	if err != nil {
		t.Fatalf("Read(): %v", err)
	}
	if !a.AccentSeedSet || a.AccentSeed != want {
		t.Errorf("Read() AccentSeed=%+v set=%v; want %+v, true", a.AccentSeed, a.AccentSeedSet, want)
	}
	if a.Dark {
		t.Error("Dark = true; Windows dark mode is not read yet")
	}
}

// TestWindowsSourceNoAccent verifies a failed registry read folds to the
// zero Appearance, per the package's error contract.
func TestWindowsSourceNoAccent(t *testing.T) {
	src := &windowsSource{
		readAccentFn: func() (color.NRGBA, bool) { return color.NRGBA{}, false },
	}
	a, err := src.Read()
	if err != nil {
		t.Fatalf("Read(): %v", err)
	}
	if a != (Appearance{}) {
		t.Errorf("no-accent Read() = %+v; want the zero Appearance", a)
	}
}

// TestReadAccentColorLive exercises the real registry read on a Windows
// machine. Any modern desktop Windows has the DWM key; if the value is
// absent the read must still fold cleanly rather than panic.
func TestReadAccentColorLive(t *testing.T) {
	seed, ok := readAccentColor()
	t.Logf("live DWM AccentColor: %+v (set=%v)", seed, ok)
	if ok && seed.A != 0xFF {
		t.Errorf("live seed alpha = %#02x; want opaque", seed.A)
	}
}
