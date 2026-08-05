// The conversions in this file are lifted from reactivego/luminance
// (luminance.go), an MD2-era tone package whose Lighten/Darken API was a
// chroma.js port tuned to reproduce the retired material.io Color Tool.
// Only the sRGB ↔ XYZ(D65) ↔ CIELAB chain, the D65 white point and the
// CIE ϵ/κ constants were taken; Lighten, Darken, LightenRGBA, DarkenRGBA
// and Kn were left behind. The functions came out of MD2-era tone work —
// do not go looking for MD3 semantics in them; MD3 enters only through
// what this package derives with them (tone is CIELAB L*).
package color

import (
	stdcolor "image/color"
	"math"
)

const (
	// CIE Standard
	ϵ  = 216.0 / 24389.0 // ϵ ≈ 0.008856452
	κ  = 24389.0 / 27.0  // κ ≈ 903.3
	κϵ = 8.0

	// CIE Standard Illuminant D65 - 6504K
	Xr_D65 = 0.950470
	Yr_D65 = 1.0
	Zr_D65 = 1.088830
)

// RGB_to_XYZ_D65 converts R,G,B in range [0,1]
// to X,Y,Z in range [0,1]
func RGB_to_XYZ_D65(R, G, B float64) (X, Y, Z float64) {
	// 1. Inverse companding
	linearize := func(V float64) float64 {
		if V <= 0.04045 {
			return V / 12.92
		}
		return math.Pow((V+0.055)/1.055, 2.4)
	}
	red := linearize(R)
	green := linearize(G)
	blue := linearize(B)

	// 2. Linear RGB to XYZ D65
	X = (0.4124564*red + 0.3575761*green + 0.1804375*blue)
	Y = (0.2126729*red + 0.7151522*green + 0.0721750*blue)
	Z = (0.0193339*red + 0.1191920*green + 0.9503041*blue)
	return X, Y, Z
}

// XYZ_to_RGB_D65 converts X,Y,Z in range [0,1] to R,G,B in range [0,1]
func XYZ_to_RGB_D65(X, Y, Z float64) (R, G, B float64) {
	// 1. XYZ to Linear RGB D65
	red := (3.2404542*X - 1.5371385*Y - 0.4985314*Z)
	green := (-0.9692660*X + 1.8760108*Y + 0.0415560*Z)
	blue := (0.0556434*X - 0.2040259*Y + 1.0572252*Z)

	// 2. Companding
	compand := func(v float64) float64 {
		if v <= 0.0031308 {
			return v * 12.92
		}
		return 1.055*math.Pow(v, 1.0/2.4) - 0.055
	}
	R = compand(red)
	G = compand(green)
	B = compand(blue)
	return R, G, B
}

// XYZ_to_Lab_D65 converts X,Y,Z in range [0,1]
// to L in range [0,100] and a,b in range [-100,100]
func XYZ_to_Lab_D65(X, Y, Z float64) (L, a, b float64) {
	xr := X / Xr_D65
	yr := Y / Yr_D65
	zr := Z / Zr_D65
	f := func(r float64) float64 {
		if r > ϵ {
			return math.Cbrt(r)
		}
		return (κ*r + 16.0) / 116.0
	}
	fx := f(xr)
	fy := f(yr)
	fz := f(zr)
	L = 116.0*fy - 16.0
	if L < 0 {
		L = 0
	}
	a = 500.0 * (fx - fy)
	b = 200.0 * (fy - fz)
	return L, a, b
}

// Lab_to_XYZ_D65 converts L in range [0,100] and a,b in range [-100,100]
// to X,Y,Z in range [0,1]
func Lab_to_XYZ_D65(L, a, b float64) (X, Y, Z float64) {
	fy := (L + 16.0) / 116.0
	fx := fy + a/500.0
	fz := fy - b/200.0
	xr := math.Pow(fx, 3.0)
	if xr <= ϵ {
		xr = (116.0*fx - 16.0) / κ
	}
	yr := math.Pow(fy, 3.0)
	if L <= κϵ {
		yr = L / κ
	}
	zr := math.Pow(fz, 3.0)
	if zr <= ϵ {
		zr = (116.0*fz - 16.0) / κ
	}
	X = xr * Xr_D65
	Y = yr * Yr_D65
	Z = zr * Zr_D65
	return X, Y, Z
}

// Lab returns the CIELAB L* in range [0,100] and a,b value in range
// [-100,100] for the sRGB tripple with channel values in range [0,255]
func Lab(R, G, B uint8) (L float64, a float64, b float64) {
	red := float64(R) / 255.0
	green := float64(G) / 255.0
	blue := float64(B) / 255.0
	X, Y, Z := RGB_to_XYZ_D65(red, green, blue)
	L, a, b = XYZ_to_Lab_D65(X, Y, Z)
	return L, a, b
}

// RGB returns the sRGB tripple in range [0,255] for the
// CIELAB L* in range [0,100] and a,b value in range [-100,100]
// Any L* values outside the valid [0,100] range are clamped.
//
// Out-of-gamut a,b are clamped per channel, which keeps RGB total but is
// not gamut mapping — the clamp shifts hue and chroma. It is replaced by
// real gamut mapping in a later step; until then treat out-of-gamut
// results as approximate.
func RGB(L, a, b float64) (R uint8, G uint8, B uint8) {
	clamp := func(val, min, max float64) float64 {
		return math.Max(min, math.Min(val, max))
	}
	scale := func(val float64) uint8 {
		return uint8(math.Round(255.0 * clamp(val, 0, 1)))
	}
	L = clamp(L, 0, 100)
	X, Y, Z := Lab_to_XYZ_D65(L, a, b)
	red, green, blue := XYZ_to_RGB_D65(X, Y, Z)
	R = scale(red)
	G = scale(green)
	B = scale(blue)
	return R, G, B
}

// LabFromNRGBA returns the CIELAB L*, a, b for an image/color NRGBA value.
// The alpha channel is ignored: NRGBA is non-premultiplied, so the colour
// channels are the colour regardless of coverage. Not part of the lift.
func LabFromNRGBA(c stdcolor.NRGBA) (L, a, b float64) {
	return Lab(c.R, c.G, c.B)
}

// NRGBAFromLab returns the fully opaque image/color NRGBA value for the
// CIELAB L*, a, b, with the same clamping as RGB. Not part of the lift.
func NRGBAFromLab(L, a, b float64) stdcolor.NRGBA {
	R, G, B := RGB(L, a, b)
	return stdcolor.NRGBA{R: R, G: G, B: B, A: 0xff}
}
