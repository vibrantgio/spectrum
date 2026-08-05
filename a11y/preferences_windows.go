package a11y

import (
	"syscall"
	"unsafe"
)

const (
	spiGetClientAreaAnimation = 0x1042
	spiGetHighContrast        = 0x0042
	hcfHighContrastOn         = 0x00000001
)

// highContrast mirrors the Windows HIGHCONTRASTW layout.
type highContrast struct {
	cbSize            uint32
	dwFlags           uint32
	lpszDefaultScheme uintptr // LPWSTR — same size as pointer, not dereferenced
}

var (
	modUser32                = syscall.NewLazyDLL("user32.dll")
	procSystemParametersInfo = modUser32.NewProc("SystemParametersInfoW")
)

type windowsSource struct{}

func (windowsSource) Read() (A11yPrefs, error) {
	var animEnabled uint32
	procSystemParametersInfo.Call(
		uintptr(spiGetClientAreaAnimation),
		0,
		uintptr(unsafe.Pointer(&animEnabled)),
		0,
	)

	var hc highContrast
	hc.cbSize = uint32(unsafe.Sizeof(hc))
	procSystemParametersInfo.Call(
		uintptr(spiGetHighContrast),
		uintptr(unsafe.Sizeof(hc)),
		uintptr(unsafe.Pointer(&hc)),
		0,
	)

	return A11yPrefs{
		ReduceMotion: animEnabled == 0,
		HighContrast: hc.dwFlags&hcfHighContrastOn != 0,
		// No Windows API directly exposes an "increase text size" preference.
		IncreaseTextSize: false,
	}, nil
}

func defaultSource() Source { return windowsSource{} }
