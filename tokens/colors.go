package tokens

import "image/color"

// ColorScale holds the eleven Tailwind shade stops for one hue family (50–950).
type ColorScale struct {
	C50, C100, C200, C300, C400, C500, C600, C700, C800, C900, C950 color.NRGBA
}

// Palette families taken verbatim from the Tailwind CSS v3 default config.
var (
	Slate = ColorScale{
		C50:  color.NRGBA{0xf8, 0xfa, 0xfc, 0xff},
		C100: color.NRGBA{0xf1, 0xf5, 0xf9, 0xff},
		C200: color.NRGBA{0xe2, 0xe8, 0xf0, 0xff},
		C300: color.NRGBA{0xcb, 0xd5, 0xe1, 0xff},
		C400: color.NRGBA{0x94, 0xa3, 0xb8, 0xff},
		C500: color.NRGBA{0x64, 0x74, 0x8b, 0xff},
		C600: color.NRGBA{0x47, 0x55, 0x69, 0xff},
		C700: color.NRGBA{0x33, 0x41, 0x55, 0xff},
		C800: color.NRGBA{0x1e, 0x29, 0x3b, 0xff},
		C900: color.NRGBA{0x0f, 0x17, 0x2a, 0xff},
		C950: color.NRGBA{0x02, 0x06, 0x17, 0xff},
	}
	Blue = ColorScale{
		C50:  color.NRGBA{0xef, 0xf6, 0xff, 0xff},
		C100: color.NRGBA{0xdb, 0xea, 0xfe, 0xff},
		C200: color.NRGBA{0xbf, 0xdb, 0xfe, 0xff},
		C300: color.NRGBA{0x93, 0xc5, 0xfd, 0xff},
		C400: color.NRGBA{0x60, 0xa5, 0xfa, 0xff},
		C500: color.NRGBA{0x3b, 0x82, 0xf6, 0xff},
		C600: color.NRGBA{0x25, 0x63, 0xeb, 0xff},
		C700: color.NRGBA{0x1d, 0x4e, 0xd8, 0xff},
		C800: color.NRGBA{0x1e, 0x40, 0xaf, 0xff},
		C900: color.NRGBA{0x1e, 0x3a, 0x8a, 0xff},
		C950: color.NRGBA{0x17, 0x25, 0x54, 0xff},
	}
	Red = ColorScale{
		C50:  color.NRGBA{0xfe, 0xf2, 0xf2, 0xff},
		C100: color.NRGBA{0xfe, 0xe2, 0xe2, 0xff},
		C200: color.NRGBA{0xfe, 0xca, 0xca, 0xff},
		C300: color.NRGBA{0xfc, 0xa5, 0xa5, 0xff},
		C400: color.NRGBA{0xf8, 0x71, 0x71, 0xff},
		C500: color.NRGBA{0xef, 0x44, 0x44, 0xff},
		C600: color.NRGBA{0xdc, 0x26, 0x26, 0xff},
		C700: color.NRGBA{0xb9, 0x1c, 0x1c, 0xff},
		C800: color.NRGBA{0x99, 0x1b, 0x1b, 0xff},
		C900: color.NRGBA{0x7f, 0x1d, 0x1d, 0xff},
		C950: color.NRGBA{0x45, 0x0a, 0x0a, 0xff},
	}

	White = color.NRGBA{0xff, 0xff, 0xff, 0xff}
)

// ColorTokens holds the semantic foreground/background token pairs consumed by
// every Prism component. Each "On" field is the recommended text/icon colour
// rendered on top of its companion field.
type ColorTokens struct {
	Background       color.NRGBA
	OnBackground     color.NRGBA
	Surface          color.NRGBA
	OnSurface        color.NRGBA
	SurfaceVariant   color.NRGBA
	OnSurfaceVariant color.NRGBA
	Primary          color.NRGBA
	OnPrimary        color.NRGBA
	Secondary        color.NRGBA
	OnSecondary      color.NRGBA
	Error            color.NRGBA
	OnError          color.NRGBA
	Outline          color.NRGBA // border/divider; no "On" counterpart
}

// DefaultLight is the canonical light-mode colour token set.
var DefaultLight = ColorTokens{
	Background:       White,
	OnBackground:     Slate.C900,
	Surface:          Slate.C50,
	OnSurface:        Slate.C900,
	SurfaceVariant:   Slate.C100,
	OnSurfaceVariant: Slate.C700,
	Primary:          Blue.C700,
	OnPrimary:        White,
	Secondary:        Slate.C600,
	OnSecondary:      White,
	Error:            Red.C700,
	OnError:          White,
	Outline:          Slate.C300,
}

// DefaultDark is the canonical dark-mode colour token set.
var DefaultDark = ColorTokens{
	Background:       Slate.C950,
	OnBackground:     Slate.C50,
	Surface:          Slate.C900,
	OnSurface:        Slate.C100,
	SurfaceVariant:   Slate.C800,
	OnSurfaceVariant: Slate.C300,
	Primary:          Blue.C400,
	OnPrimary:        Slate.C900,
	Secondary:        Slate.C400,
	OnSecondary:      Slate.C900,
	Error:            Red.C400,
	OnError:          Slate.C900,
	Outline:          Slate.C700,
}
