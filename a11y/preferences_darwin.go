package a11y

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit

#import <AppKit/AppKit.h>

// a11yRead returns OS accessibility flags as a bit field:
//   bit 0 — reduce motion   (NSWorkspace.accessibilityDisplayShouldReduceMotion)
//   bit 1 — increase contrast (NSWorkspace.accessibilityDisplayShouldIncreaseContrast)
static int a11yRead(void) {
    int flags = 0;
    NSWorkspace *ws = [NSWorkspace sharedWorkspace];
    if ([ws accessibilityDisplayShouldReduceMotion])     flags |= 1;
    if ([ws accessibilityDisplayShouldIncreaseContrast]) flags |= 2;
    return flags;
}
*/
import "C"

type darwinSource struct{}

func (darwinSource) Read() (A11yPrefs, error) {
	flags := int(C.a11yRead())
	return A11yPrefs{
		ReduceMotion: flags&1 != 0,
		HighContrast: flags&2 != 0,
		// NSWorkspace exposes no "increase text size" API on macOS; always false.
		IncreaseTextSize: false,
	}, nil
}

func defaultSource() Source { return darwinSource{} }
