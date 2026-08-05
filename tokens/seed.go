// Seed-derived palettes. FromSeed turns one brand colour into ADR-007's
// complete paired light and dark ramp sets, and DefaultLight/DefaultDark
// are FromSeed of the default seed.
//
// Derivation rules, and where each comes from:
//
//   - The shared lightness scale is ADR-007's: CIELAB L* per step, measured
//     by the D0.1 spike from the Claude Design reference project's own
//     ramps. Light 100–900 = 97, 92, 85, 74, 63, 51, 39, 28, 18. The dark
//     scale is the paired scale measured from the same source's dark column
//     (ADR-007's evidence table): 8, 13, 19, 30, 65, —, 82, —, 94; the 600
//     and 800 stops the table has no surface for are interpolated to 74 and
//     88. Both scales are swept at constant OKLCh hue and chroma via
//     color.Tone, which gamut-maps by chroma reduction (ADR-002).
//
//   - Accent hues and chromas follow MD3's material-color-utilities
//     conventions, which ADR-007 does not supersede, converted into the
//     OKLCh chroma axis. The conversion anchor: the canonical seed #6750A4
//     has HCT chroma 48 and measures OKLCh chroma 0.1305, so one HCT chroma
//     unit ≈ 0.00272 OKLCh chroma. Neutral = seed hue at chroma 0.010
//     (MD3's neutral 4; ADR-007's measured reference columns sit at
//     0.009–0.011). Secondary = seed hue at chroma 0.044 (MD3's 16).
//     Tertiary = seed hue +60° at chroma 0.065 (MD3's 24). Error is a fixed
//     red: hue 28.7° at chroma 0.178, the OKLCh coordinates of MD3's
//     canonical error base #B3261E (its "hue 25, chroma 84") measured with
//     this module's own converters. Primary uses the seed's measured hue
//     and chroma unchanged.
//
//   - Pins. The light primary base is the seed byte-for-byte (ADR-007:
//     "the seed sits deep, so bases are pins" — reading it off the ramp
//     would lighten it); only its alpha is forced opaque. The other light
//     bases are their role's hue and chroma at tone 40, the depth MD3 pins
//     accent bases at and the depth the default seed itself sits at
//     (L* 40.08). Dark bases are the same hue and chroma re-toned to L* 65
//     — the dark scale's step-500 depth — which reproduces ADR-007's
//     recorded dark fill #a690ea for the default seed byte-for-byte.
//
//   - On-colours. Light bases sit at tone 40, so their on-colour is White;
//     dark bases sit at L* 65, so their on-colour is their own dark ramp's
//     step 100. Both meet ADR-007's Lc ≥ 60 intent with room to spare
//     (WCAG ≈ 6.4:1 either way); D2.4 adds the APCA gate that enforces it.
package tokens

import (
	stdcolor "image/color"

	"github.com/vibrantgio/spectrum/color"
)

// lightTones and darkTones are the shared perceptual lightness scale:
// CIELAB L* for steps 100–900, light and paired dark, per ADR-007. Index i
// holds step (i+1)*100, matching Ramp.
var (
	lightTones = [9]int{97, 92, 85, 74, 63, 51, 39, 28, 18}
	darkTones  = [9]int{8, 13, 19, 30, 65, 74, 82, 88, 94}
)

// Accent-derivation constants; see the file header for provenance.
const (
	neutralChroma    = 0.010 // MD3 neutral chroma 4 in OKLCh units
	secondaryChroma  = 0.044 // MD3 secondary chroma 16
	tertiaryChroma   = 0.065 // MD3 tertiary chroma 24
	tertiaryHueShift = 60    // MD3: tertiary is the seed hue rotated +60°
	errorHue         = 28.7  // OKLCh hue of MD3's error base #B3261E
	errorChroma      = 0.178 // OKLCh chroma of #B3261E
	lightPinTone     = 40    // MD3's accent-base tone; the default seed's own depth
	darkPinTone      = 65    // the dark scale's step-500 L*; yields #a690ea for #6750A4
)

// rampOf sweeps one role's hue and chroma across a lightness scale.
func rampOf(tones [9]int, hue, chroma float64) Ramp {
	var r Ramp
	for i, tone := range tones {
		r[i] = color.Tone(hue, chroma, tone)
	}
	return r
}

// FromSeed derives the complete paired light and dark colour token sets
// from one brand seed: for every role a nine-step ramp on the shared
// lightness scale in both modes — the same step keeps the same job — plus
// the pinned bases, on-colours and semantic layer, per the rules in the
// file header. The light primary base is the seed itself, byte-for-byte
// (alpha forced opaque); every other value is generated.
//
// DefaultLight and DefaultDark are FromSeed(DefaultSeed). Applications
// re-brand by calling FromSeed with their own colour and handing the pair
// to a theme.
func FromSeed(seed stdcolor.NRGBA) (light, dark ColorTokens) {
	seed.A = 0xff
	_, seedChroma, seedHue := color.OKLChFromNRGBA(seed)

	roles := []struct {
		hue, chroma float64
	}{
		{seedHue, neutralChroma},
		{seedHue, seedChroma},
		{seedHue, secondaryChroma},
		{seedHue + tertiaryHueShift, tertiaryChroma},
		{errorHue, errorChroma},
	}
	var lr, dr [5]Ramp
	for i, role := range roles {
		lr[i] = rampOf(lightTones, role.hue, role.chroma)
		dr[i] = rampOf(darkTones, role.hue, role.chroma)
	}
	lightPin := func(i int) stdcolor.NRGBA {
		return color.Tone(roles[i].hue, roles[i].chroma, lightPinTone)
	}
	darkPin := func(i int) stdcolor.NRGBA {
		return color.Tone(roles[i].hue, roles[i].chroma, darkPinTone)
	}

	light = resolveAliases(ColorTokens{
		Ramps:       RampSet{Neutral: lr[0], Primary: lr[1], Secondary: lr[2], Tertiary: lr[3], Error: lr[4]},
		Primary:     seed, // pinned to the seed exactly, never read off the ramp
		OnPrimary:   White,
		Secondary:   lightPin(2),
		OnSecondary: White,
		Tertiary:    lightPin(3),
		OnTertiary:  White,
		Error:       lightPin(4),
		OnError:     White,
		Background:  lr[0].Step(100),
		Text:        lr[0].Step(900),
	})
	dark = resolveAliases(ColorTokens{
		Ramps:       RampSet{Neutral: dr[0], Primary: dr[1], Secondary: dr[2], Tertiary: dr[3], Error: dr[4]},
		Primary:     darkPin(1),
		OnPrimary:   dr[1].Step(100),
		Secondary:   darkPin(2),
		OnSecondary: dr[2].Step(100),
		Tertiary:    darkPin(3),
		OnTertiary:  dr[3].Step(100),
		Error:       darkPin(4),
		OnError:     dr[4].Step(100),
		Background:  dr[0].Step(100),
		Text:        dr[0].Step(900),
	})
	return light, dark
}

// DefaultSeed is the brand seed DefaultLight and DefaultDark derive from:
// #6750A4, the seed every ADR-002/ADR-007 measurement was made against.
var DefaultSeed = stdcolor.NRGBA{R: 0x67, G: 0x50, B: 0xA4, A: 0xff}

// DefaultLight and DefaultDark are the canonical colour token sets:
// FromSeed(DefaultSeed), light and paired dark. The exact derived palette
// is recorded byte-for-byte in this package's golden test.
var DefaultLight, DefaultDark = FromSeed(DefaultSeed)
