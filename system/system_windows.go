package system

import (
	"image/color"

	"golang.org/x/sys/windows/registry"
)

// windowsSource reads the accent colour from the registry:
// HKEY_CURRENT_USER\Software\Microsoft\Windows\DWM, value AccentColor, an
// ABGR DWORD holding the arbitrary colour the user picked in Settings →
// Personalization → Colors. That colour cannot be an [Accent] enum value,
// so it travels as Appearance.AccentSeed (see the decode in
// nrgbaFromABGR). Reading uses golang.org/x/sys/windows/registry — pure
// syscalls, no cgo; the module already carried x/sys transitively via Gio,
// so this promotes an existing dependency to direct rather than adding one.
//
// Unlike the darwin source there is no throttle: a registry query is a few
// syscalls, microseconds against the poll interval, so every Read reflects
// fresh state.
//
// Dark mode is not read yet — Dark stays false. A live implementation
// would read HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion
// \Themes\Personalize\AppsUseLightTheme; that is a later milestone (see
// the package doc's support matrix).
type windowsSource struct {
	readAccentFn func() (color.NRGBA, bool) // injectable for tests
}

func newWindowsSource() *windowsSource {
	return &windowsSource{readAccentFn: readAccentColor}
}

func (s *windowsSource) Read() (Appearance, error) {
	seed, ok := s.readAccentFn()
	return Appearance{
		AccentSeed:    seed,
		AccentSeedSet: ok,
	}, nil
}

// readAccentColor reads and decodes the DWM AccentColor value. Any failure
// — key or value absent, wrong type — folds to "no accent", per the
// package's error contract: a broken source is indistinguishable from no
// accent override.
func readAccentColor() (color.NRGBA, bool) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\DWM`, registry.QUERY_VALUE)
	if err != nil {
		return color.NRGBA{}, false
	}
	defer k.Close()
	v, _, err := k.GetIntegerValue("AccentColor")
	if err != nil {
		return color.NRGBA{}, false
	}
	return nrgbaFromABGR(uint32(v)), true
}

func defaultSource() Source { return newWindowsSource() }
