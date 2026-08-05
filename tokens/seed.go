// Seed-derived palettes. FromSeed turns one brand colour into ADR-007's
// complete paired light and dark ramp sets, and DefaultLight/DefaultDark
// are FromSeed of the default seed.
//
// Derivation rules, and where each comes from:
//
//   - The shared lightness scale is ADR-007's: CIELAB L* per step, measured
//     by the D0.1 spike from the Claude Design reference project's own
//     ramps. Light 100–900 = 97, 92, 85, 74, 63, 51, 39, 28, 6 — the
//     measured 900 was L* 18, but D2.4's APCA gate deepened it: APCA's soft
//     black clamp caps even pure black near Lc 92 over the L* 92 step-200
//     ground, so Lc ≥ 90 needs the 900 stop at L* 6 (min Lc 90.7 across the
//     five default ramps; L* 18 measured Lc 85–87). The dark scale is the
//     paired scale measured from the same source's dark column (ADR-007's
//     evidence table): 8, 13, 19, 30, 65, —, 82, —, 94; the 600 and 800
//     stops the table has no surface for are interpolated to 74 and 88 (the
//     dark 900 already clears the gate at Lc 93–96, so it is untouched).
//     Both scales are swept at constant OKLCh hue and chroma via
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
//     (L* 40.08). Dark bases are the same hue and chroma re-toned to L* 82
//     — the dark scale's step-700 depth, right beside MD3's dark
//     accent-base tone 80, making the dark pin byte-identical to its ramp's
//     step 700. The D0.1 spike sat them at L* 65, the step-500 depth
//     reproducing ADR-007's recorded dark fill #a690ea, but D2.4's APCA
//     gate showed an L* 65 mid-tone is a ground no text can reach Lc 60
//     over (black tops out near Lc 52, white near 57), so the pins moved up
//     two rungs — the default seed's dark primary is now #d0c4ff — and the
//     solid state walk still lands on exact rungs (hover 800, pressed 900).
//
//   - On-colours. Light bases sit at tone 40, so their on-colour is White
//     (Lc ≥ 85, WCAG ≈ 6.4:1); dark bases sit at L* 82, so their on-colour
//     is their own dark ramp's step 100 (Lc ≥ 73, WCAG ≈ 11:1). D2.4's
//     APCA gate enforces ADR-007's Lc ≥ 60 on both.
//
// FromSeedHighContrast (task E3.3) derives the increased-contrast variant
// from the same seed by the same machinery — it is a FromSeed option, not a
// third hand-written scheme. Three widenings, each computed against the
// APCA gate rather than guessed:
//
//   - The 700 text step deepens to the default scale's 900 depth in both
//     modes — light 700 L* 39 → 6, dark 700 L* 82 → 94 — so 700 text meets
//     the same Lc ≥ 90 bar the default asks only of 900 (light min Lc 90.7,
//     dark 93.0 across the five ramps; APCA's soft black clamp caps lighter
//     choices below 90, the same wall D2.4 hit). The 800 and 900 stops
//     slide outward — light 3 and 0, dark 97 and 100 — keeping the ladder
//     strictly monotonic and the 900 gate clear with margin (light Lc 92.3,
//     dark 104.4). Steps 100–600 are the default scale unchanged: the
//     grounds stay, the text pulls away.
//
//   - Divider resolves from Neutral step 500 instead of 300: the separator
//     jumps from the subtle-border rung to ADR-007's strong-border rung.
//
//   - Each pinned base's on-colour is pushed further from its base. The
//     dark pins' on-colours drop from their ramp's step 100 (L* 8, Lc ≈ 74
//     — just under the variant's Lc ≥ 75 floor) to tone 0, the scale's
//     floor (Lc ≥ 76.3). The light pins keep White: it is already the far
//     end of the axis and already clears the floor (Lc ≥ 85.7), so the
//     pins do not move — the light primary base stays the seed
//     byte-for-byte, the same contract as FromSeed.
package tokens

import (
	stdcolor "image/color"

	"github.com/vibrantgio/spectrum/color"
)

// lightTones and darkTones are the shared perceptual lightness scale:
// CIELAB L* for steps 100–900, light and paired dark, per ADR-007. Index i
// holds step (i+1)*100, matching Ramp.
var (
	lightTones = [9]int{97, 92, 85, 74, 63, 51, 39, 28, 6}
	darkTones  = [9]int{8, 13, 19, 30, 65, 74, 82, 88, 94}
)

// hcLightTones and hcDarkTones are the high-contrast variant's scales:
// steps 100–600 are the default scale unchanged, 700 deepens to the default
// 900 depth so 700 text meets the Lc ≥ 90 bar, and 800/900 slide to the
// axis ends to keep the ladder strictly monotonic. See the file header.
var (
	hcLightTones = [9]int{97, 92, 85, 74, 63, 51, 6, 3, 0}
	hcDarkTones  = [9]int{8, 13, 19, 30, 65, 74, 94, 97, 100}
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
	darkPinTone      = 82    // the dark scale's step-700 L*; D2.4 raised it from the
	// spike's 65 — no on-colour reaches Lc 60 over an L* 65 mid-tone
	darkOnTone   = 8 // dark pins' on-colour depth: the dark scale's step-100 L*
	hcDarkOnTone = 0 // high contrast pushes the dark on-colours to the axis floor
)

// derivation is the knob set that separates FromSeed from its high-contrast
// variant: the two lightness scales, the ramp step Divider resolves from,
// and the CIELAB L* the dark pins' on-colours are realized at. Everything
// else — hues, chromas, pin depths, the seed-exact light primary — is
// shared, which is what makes the variant a FromSeed option rather than a
// third hand-written scheme.
type derivation struct {
	lightTones, darkTones [9]int
	dividerStep           int // ramp step Divider resolves from
	darkOnTone            int // L* of the dark pins' on-colours
}

var (
	defaultDerivation = derivation{lightTones, darkTones, 300, darkOnTone}
	hcDerivation      = derivation{hcLightTones, hcDarkTones, 500, hcDarkOnTone}
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
	return fromSeed(seed, defaultDerivation)
}

// FromSeedHighContrast derives the increased-contrast variant of FromSeed's
// pair from the same seed: same roles, hues, chromas and pin depths, with
// the tone separation widened where it counts — the 700 text step deepened
// to the default 900 depth (Lc ≥ 90 where the default asks 60), Divider
// resolved from Neutral step 500 instead of 300, and the dark pins'
// on-colours pushed to the tonal axis floor (Lc ≥ 75 over their bases; the
// light pins keep White, which already clears that floor, so the light
// primary base is still the seed byte-for-byte). The full rules are in the
// file header; the derived default-seed variant is recorded in this
// package's high-contrast golden test.
//
// It is the palette spectrum/system swaps in while the OS reports increased
// contrast (see system.HighContrastVariant).
func FromSeedHighContrast(seed stdcolor.NRGBA) (light, dark ColorTokens) {
	return fromSeed(seed, hcDerivation)
}

// fromSeed is the shared derivation engine behind FromSeed and
// FromSeedHighContrast; d selects which variant's knobs apply.
func fromSeed(seed stdcolor.NRGBA, d derivation) (light, dark ColorTokens) {
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
		lr[i] = rampOf(d.lightTones, role.hue, role.chroma)
		dr[i] = rampOf(d.darkTones, role.hue, role.chroma)
	}
	lightPin := func(i int) stdcolor.NRGBA {
		return color.Tone(roles[i].hue, roles[i].chroma, lightPinTone)
	}
	darkPin := func(i int) stdcolor.NRGBA {
		return color.Tone(roles[i].hue, roles[i].chroma, darkPinTone)
	}
	darkOn := func(i int) stdcolor.NRGBA {
		return color.Tone(roles[i].hue, roles[i].chroma, d.darkOnTone)
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
	}, d.dividerStep)
	dark = resolveAliases(ColorTokens{
		Ramps:       RampSet{Neutral: dr[0], Primary: dr[1], Secondary: dr[2], Tertiary: dr[3], Error: dr[4]},
		Primary:     darkPin(1),
		OnPrimary:   darkOn(1),
		Secondary:   darkPin(2),
		OnSecondary: darkOn(2),
		Tertiary:    darkPin(3),
		OnTertiary:  darkOn(3),
		Error:       darkPin(4),
		OnError:     darkOn(4),
		Background:  dr[0].Step(100),
		Text:        dr[0].Step(900),
	}, d.dividerStep)
	return light, dark
}

// DefaultSeed is the brand seed DefaultLight and DefaultDark derive from:
// #6750A4, the seed every ADR-002/ADR-007 measurement was made against.
var DefaultSeed = stdcolor.NRGBA{R: 0x67, G: 0x50, B: 0xA4, A: 0xff}

// DefaultLight and DefaultDark are the canonical colour token sets:
// FromSeed(DefaultSeed), light and paired dark. The exact derived palette
// is recorded byte-for-byte in this package's golden test.
var DefaultLight, DefaultDark = FromSeed(DefaultSeed)
