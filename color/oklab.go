// OKLab and OKLCh, after Björn Ottosson's "A perceptual color space for
// image processing" (bottosson.github.io/posts/oklab). Unlike the CIELAB
// chain in lab.go this file is new code, not part of the reactivego/luminance
// lift, so it uses ordinary Go names in lab.go's XFromY adapter style
// (LabFromNRGBA); lab.go keeps its underscored spec-style names
// (RGB_to_XYZ_D65) so the lifted code stays recognisably the lift.
//
// The chain is sRGB ↔ linear sRGB ↔ LMS ↔ OKLab ↔ OKLCh. All matrix
// coefficients are Ottosson's published values at full precision, including
// his published inverses — the pairs are rounded independently rather than
// derived from each other, exactly as the reference implementation has them.
package color

import (
	stdcolor "image/color"
	"math"
)

// neutralChroma is the OKLab chroma below which hue is considered undefined
// and reported as 0°. Exact greys land near 4e-8 through the published
// (rounded) matrices; anything under this threshold is numeric noise.
const neutralChroma = 1e-7

// LinearFromSRGB converts one sRGB channel value in [0,1] to linear sRGB
// in [0,1] using the standard sRGB EOTF (the 0.04045/12.92 piecewise).
func LinearFromSRGB(V float64) float64 {
	if V <= 0.04045 {
		return V / 12.92
	}
	return math.Pow((V+0.055)/1.055, 2.4)
}

// SRGBFromLinear converts one linear sRGB channel value in [0,1] to sRGB
// in [0,1], the inverse of LinearFromSRGB.
func SRGBFromLinear(v float64) float64 {
	if v <= 0.0031308 {
		return v * 12.92
	}
	return 1.055*math.Pow(v, 1.0/2.4) - 0.055
}

// OKLabFromLinearRGB converts linear sRGB in [0,1] to OKLab: L in [0,1],
// a,b roughly in [-0.4,0.4] for in-gamut colours. Out-of-gamut input is
// accepted; the cube root extends over negatives.
func OKLabFromLinearRGB(red, green, blue float64) (L, a, b float64) {
	l := 0.4122214708*red + 0.5363325363*green + 0.0514459929*blue
	m := 0.2119034982*red + 0.6806995451*green + 0.1073969566*blue
	s := 0.0883024619*red + 0.2817188376*green + 0.6299787005*blue

	l = math.Cbrt(l)
	m = math.Cbrt(m)
	s = math.Cbrt(s)

	L = 0.2104542553*l + 0.7936177850*m - 0.0040720468*s
	a = 1.9779984951*l - 2.4285922050*m + 0.4505937099*s
	b = 0.0259040371*l + 0.7827717662*m - 0.8086757660*s
	return L, a, b
}

// LinearRGBFromOKLab converts OKLab to linear sRGB, the inverse of
// OKLabFromLinearRGB. Out-of-gamut results are returned as-is, outside
// [0,1]; nothing is clamped here.
func LinearRGBFromOKLab(L, a, b float64) (red, green, blue float64) {
	l := L + 0.3963377774*a + 0.2158037573*b
	m := L - 0.1055613458*a - 0.0638541728*b
	s := L - 0.0894841775*a - 1.2914855480*b

	l = l * l * l
	m = m * m * m
	s = s * s * s

	red = 4.0767416621*l - 3.3077115913*m + 0.2309699292*s
	green = -1.2684380046*l + 2.6097574011*m - 0.3413193965*s
	blue = -0.0041960863*l - 0.7034186147*m + 1.7076147010*s
	return red, green, blue
}

// OKLab returns the OKLab L in [0,1] and a,b (roughly [-0.4,0.4]) for the
// sRGB triple with channel values in range [0,255].
func OKLab(R, G, B uint8) (L, a, b float64) {
	red := LinearFromSRGB(float64(R) / 255.0)
	green := LinearFromSRGB(float64(G) / 255.0)
	blue := LinearFromSRGB(float64(B) / 255.0)
	return OKLabFromLinearRGB(red, green, blue)
}

// RGBFromOKLab returns the sRGB triple in range [0,255] for the OKLab L,a,b.
// L outside [0,1] is clamped.
//
// Out-of-gamut a,b are gamut mapped: chroma is reduced toward 0 at
// constant OKLab L and constant hue until the colour fits sRGB — here the
// caller's lightness axis is OKLab L, so that is what is held, where RGB
// holds CIELAB L*. In-gamut input is returned untouched.
func RGBFromOKLab(L, a, b float64) (R, G, B uint8) {
	L = math.Max(0, math.Min(L, 1))
	red, green, blue := LinearRGBFromOKLab(L, a, b)
	if !inSRGBGamut(red, green, blue) {
		red, green, blue = mapOKLabChroma(L, a, b)
	}
	return quantizeLinear(red, green, blue)
}

// OKLChFromOKLab converts OKLab to its polar form OKLCh: C = √(a²+b²) and
// h = atan2(b,a) in degrees normalized to [0,360). When C is below
// neutralChroma the hue is undefined and reported as 0.
func OKLChFromOKLab(l, a, b float64) (L, C, h float64) {
	L = l
	C = math.Hypot(a, b)
	if C < neutralChroma {
		return L, C, 0
	}
	h = math.Atan2(b, a) * 180.0 / math.Pi
	if h < 0 {
		h += 360.0
	}
	return L, C, h
}

// OKLabFromOKLCh converts the polar OKLCh form back to OKLab. h is in
// degrees; any value is accepted, not only [0,360).
func OKLabFromOKLCh(l, C, h float64) (L, a, b float64) {
	L = l
	rad := h * math.Pi / 180.0
	a = C * math.Cos(rad)
	b = C * math.Sin(rad)
	return L, a, b
}

// OKLCh returns the OKLCh L in [0,1], chroma C and hue h in degrees
// [0,360) for the sRGB triple with channel values in range [0,255].
func OKLCh(R, G, B uint8) (L, C, h float64) {
	return OKLChFromOKLab(OKLab(R, G, B))
}

// RGBFromOKLCh returns the sRGB triple in range [0,255] for the OKLCh
// L,C,h, with the same gamut mapping as RGBFromOKLab.
func RGBFromOKLCh(L, C, h float64) (R, G, B uint8) {
	return RGBFromOKLab(OKLabFromOKLCh(L, C, h))
}

// OKLabFromNRGBA returns the OKLab L,a,b for an image/color NRGBA value.
// The alpha channel is ignored: NRGBA is non-premultiplied, so the colour
// channels are the colour regardless of coverage.
func OKLabFromNRGBA(c stdcolor.NRGBA) (L, a, b float64) {
	return OKLab(c.R, c.G, c.B)
}

// NRGBAFromOKLab returns the fully opaque image/color NRGBA value for the
// OKLab L,a,b, with the same gamut mapping as RGBFromOKLab.
func NRGBAFromOKLab(L, a, b float64) stdcolor.NRGBA {
	R, G, B := RGBFromOKLab(L, a, b)
	return stdcolor.NRGBA{R: R, G: G, B: B, A: 0xff}
}

// OKLChFromNRGBA returns the OKLCh L,C,h for an image/color NRGBA value,
// ignoring alpha like OKLabFromNRGBA.
func OKLChFromNRGBA(c stdcolor.NRGBA) (L, C, h float64) {
	return OKLCh(c.R, c.G, c.B)
}

// NRGBAFromOKLCh returns the fully opaque image/color NRGBA value for the
// OKLCh L,C,h, with the same gamut mapping as RGBFromOKLCh.
func NRGBAFromOKLCh(L, C, h float64) stdcolor.NRGBA {
	R, G, B := RGBFromOKLCh(L, C, h)
	return stdcolor.NRGBA{R: R, G: G, B: B, A: 0xff}
}
