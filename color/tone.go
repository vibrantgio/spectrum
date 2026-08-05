// Tone: the palette-facing tonal entry point. A thin, ergonomic wrapper
// over gamut.go's solver for the common case — integer MD3 tone stops at a
// fixed hue and chroma. New code, not part of the reactivego/luminance
// lift, so like oklab.go it uses ordinary Go names.
package color

import stdcolor "image/color"

// Tone returns the fully opaque image/color NRGBA value at the given MD3
// tone for a tonal palette with OKLCh hue in degrees and OKLCh chroma.
// Tone is Material Design 3's tone axis — CIELAB L*, 0 exactly black to
// 100 exactly white — and MD3's thirteen standard stops are 0, 10, 20, 30,
// 40, 50, 60, 70, 80, 90, 95, 99 and 100, though any tone in [0,100] is
// valid; tones outside that range are clamped to it. Chroma is reduced
// toward 0 at constant tone and hue when the request does not fit sRGB,
// per NRGBAFromToneChromaHue, which this wraps.
func Tone(hue, chroma float64, tone int) stdcolor.NRGBA {
	return NRGBAFromToneChromaHue(float64(tone), chroma, hue)
}
