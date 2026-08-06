package tokens_test

import (
	"image/color"
	"math"
	"testing"

	"github.com/vibrantgio/spectrum/tokens"
)

// contrastPair is a named foreground/background pair for WCAG AA verification.
type contrastPair struct {
	name string
	bg   color.NRGBA
	fg   color.NRGBA
}

// tokenPairs returns the foreground/background pairs every scheme must
// carry at WCAG AA. The pinned roles pair with their "On" colours; the
// surface pairs are named by the ramp steps they resolve from, which is
// what the five MD3 aliases deleted in v0.2.0 resolved to — the coverage
// outlived the field names.
func tokenPairs(t tokens.ColorTokens) []contrastPair {
	n := t.Ramps.Neutral
	return []contrastPair{
		{"Background/Text", t.Background, t.Text},
		{"Surface/Neutral.Step(900)", t.Surface, n.Step(900)},
		{"Neutral.Step(300)/Neutral.Step(700)", n.Step(300), n.Step(700)},
		{"Primary/OnPrimary", t.Primary, t.OnPrimary},
		{"Secondary/OnSecondary", t.Secondary, t.OnSecondary},
		{"Tertiary/OnTertiary", t.Tertiary, t.OnTertiary},
		{"Error/OnError", t.Error, t.OnError},
	}
}

func linearize(c float64) float64 {
	if c <= 0.04045 {
		return c / 12.92
	}
	return math.Pow((c+0.055)/1.055, 2.4)
}

func relativeLuminance(c color.NRGBA) float64 {
	r := linearize(float64(c.R) / 255.0)
	g := linearize(float64(c.G) / 255.0)
	b := linearize(float64(c.B) / 255.0)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func contrastRatio(c1, c2 color.NRGBA) float64 {
	l1 := relativeLuminance(c1)
	l2 := relativeLuminance(c2)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

const wcagAA = 4.5

// TestRampStepAddressing verifies Step's 100–900 addressing over the
// backing array and that out-of-vocabulary steps panic.
func TestRampStepAddressing(t *testing.T) {
	var r tokens.Ramp
	for i := range r {
		r[i] = color.NRGBA{R: uint8(i + 1), A: 0xff}
	}
	for n := 100; n <= 900; n += 100 {
		want := color.NRGBA{R: uint8(n / 100), A: 0xff}
		if got := r.Step(n); got != want {
			t.Errorf("Step(%d) = %v, want %v", n, got, want)
		}
	}
	for _, bad := range []int{0, 50, 99, 150, 901, 1000, -100} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("Step(%d): expected panic", bad)
				}
			}()
			r.Step(bad)
		}()
	}
}

// TestSemanticLayerResolvesFromRamps verifies the semantic layer resolves
// per ADR-007's surface mapping in both default schemes. Through v0.1.x
// this also pinned the five MD3 aliases to their ramp steps; F3.3 deleted
// them, leaving Surface and Divider as the only resolved fields.
func TestSemanticLayerResolvesFromRamps(t *testing.T) {
	for _, s := range []struct {
		name string
		tok  tokens.ColorTokens
	}{
		{"DefaultLight", tokens.DefaultLight},
		{"DefaultDark", tokens.DefaultDark},
	} {
		n := s.tok.Ramps.Neutral
		checks := []struct {
			name string
			got  color.NRGBA
			want color.NRGBA
		}{
			{"Surface = Neutral.Step(200)", s.tok.Surface, n.Step(200)},
			{"Divider = Neutral.Step(300)", s.tok.Divider, n.Step(300)},
		}
		for _, c := range checks {
			if c.got != c.want {
				t.Errorf("%s: %s: got %v, want %v", s.name, c.name, c.got, c.want)
			}
		}
	}
}

// TestAllRampStepsPopulated verifies every step of every role ramp, every
// pin and every semantic colour is an opaque colour in both default schemes.
func TestAllRampStepsPopulated(t *testing.T) {
	for _, s := range []struct {
		name string
		tok  tokens.ColorTokens
	}{
		{"DefaultLight", tokens.DefaultLight},
		{"DefaultDark", tokens.DefaultDark},
	} {
		for _, r := range namedRamps(s.tok) {
			for step := 100; step <= 900; step += 100 {
				if c := r.ramp.Step(step); c.A != 0xff {
					t.Errorf("%s: %s.Step(%d) = %v, want an opaque colour", s.name, r.name, step, c)
				}
			}
		}
		for name, c := range map[string]color.NRGBA{
			"Primary": s.tok.Primary, "OnPrimary": s.tok.OnPrimary,
			"Secondary": s.tok.Secondary, "OnSecondary": s.tok.OnSecondary,
			"Tertiary": s.tok.Tertiary, "OnTertiary": s.tok.OnTertiary,
			"Error": s.tok.Error, "OnError": s.tok.OnError,
			"Background": s.tok.Background, "Text": s.tok.Text,
		} {
			if c.A != 0xff {
				t.Errorf("%s: %s = %v, want an opaque colour", s.name, name, c)
			}
		}
	}
}

type namedRamp struct {
	name string
	ramp tokens.Ramp
}

func namedRamps(t tokens.ColorTokens) []namedRamp {
	return []namedRamp{
		{"Neutral", t.Ramps.Neutral},
		{"Primary", t.Ramps.Primary},
		{"Secondary", t.Ramps.Secondary},
		{"Tertiary", t.Ramps.Tertiary},
		{"Error", t.Ramps.Error},
	}
}

// TestRampLightnessDirection verifies the paired scales run the documented
// way in every ramp: in light mode step 100 is the lightest and luminance
// strictly falls toward 900; in dark mode step 100 is the darkest ground
// and luminance strictly rises toward 900 — same step, same job.
func TestRampLightnessDirection(t *testing.T) {
	for _, s := range []struct {
		name       string
		tok        tokens.ColorTokens
		descending bool
	}{
		{"DefaultLight", tokens.DefaultLight, true},
		{"DefaultDark", tokens.DefaultDark, false},
	} {
		for _, r := range namedRamps(s.tok) {
			for step := 200; step <= 900; step += 100 {
				prev := relativeLuminance(r.ramp.Step(step - 100))
				cur := relativeLuminance(r.ramp.Step(step))
				if s.descending && cur >= prev {
					t.Errorf("%s: %s.Step(%d) luminance %.4f not below step %d's %.4f",
						s.name, r.name, step, cur, step-100, prev)
				}
				if !s.descending && cur <= prev {
					t.Errorf("%s: %s.Step(%d) luminance %.4f not above step %d's %.4f",
						s.name, r.name, step, cur, step-100, prev)
				}
			}
		}
	}
}

// TestFromSeedPinsSeedExactly verifies the light primary base is the seed
// byte-for-byte — never read off a ramp step — with only the alpha forced
// opaque, for the default seed and for arbitrary ones.
func TestFromSeedPinsSeedExactly(t *testing.T) {
	for _, seed := range []color.NRGBA{
		{0x67, 0x50, 0xA4, 0xff}, // the default seed
		{0x3b, 0x82, 0xf6, 0xff},
		{0x0f, 0x76, 0x0f, 0x80}, // non-opaque alpha must be forced to 0xff
	} {
		light, _ := tokens.FromSeed(seed)
		want := color.NRGBA{seed.R, seed.G, seed.B, 0xff}
		if light.Primary != want {
			t.Errorf("FromSeed(%v): light Primary = %v, want the seed %v byte-for-byte",
				seed, light.Primary, want)
		}
	}
	if p := tokens.DefaultLight.Primary; p != (color.NRGBA{0x67, 0x50, 0xA4, 0xff}) {
		t.Errorf("DefaultLight.Primary = %v, want the seed #6750A4 byte-for-byte", p)
	}
}

// defaultGolden returns the recorded palette FromSeed derives from
// DefaultSeed (#6750A4): every ramp step, pin and semantic colour, byte for
// byte. It is the regression anchor for the whole colour engine — any
// change to the derivation (scales, chromas, hues, pins) must show up here
// as an explicit, reviewed palette change. Landmarks that tie the recording
// to ADR-007's evidence table: the dark Neutral steps reproduce the ADR's
// measured dark column (#18171c, #222126, #2e2e33, #47464c, #9e9da4,
// #cccbd2, #eeedf4) and the ADR's recorded dark fill #a690ea survives as
// the dark Primary ramp's step 500. Two D2.4 retunings moved recorded
// values off the spike's measurements: the light 900 stops deepened from
// L* 18 to L* 6 (with Text following the neutral 900) to clear the
// Lc ≥ 90 gate over the step-200 grounds, and the dark pins rose from
// L* 65 (step-500 depth) to L* 82 (step-700 depth, byte-identical to the
// ramp's 700 stop) because no on-colour reaches Lc 60 over an L* 65
// mid-tone.
func defaultGolden() (light, dark tokens.ColorTokens) {
	hex := func(r, g, b uint8) color.NRGBA { return color.NRGBA{r, g, b, 0xff} }
	light = tokens.ColorTokens{
		Ramps: tokens.RampSet{
			Neutral: tokens.Ramp{
				hex(0xf7, 0xf6, 0xfd),
				hex(0xe8, 0xe7, 0xef),
				hex(0xd5, 0xd3, 0xdb),
				hex(0xb6, 0xb5, 0xbc),
				hex(0x99, 0x98, 0x9f),
				hex(0x7a, 0x79, 0x7f),
				hex(0x5c, 0x5b, 0x61),
				hex(0x42, 0x41, 0x47),
				hex(0x14, 0x13, 0x18),
			},
			Primary: tokens.Ramp{
				hex(0xf7, 0xf5, 0xff),
				hex(0xea, 0xe5, 0xff),
				hex(0xd8, 0xce, 0xff),
				hex(0xbd, 0xaa, 0xff),
				hex(0xa0, 0x8b, 0xe5),
				hex(0x82, 0x6c, 0xc2),
				hex(0x64, 0x4d, 0xa1),
				hex(0x4b, 0x32, 0x83),
				hex(0x1d, 0x00, 0x44),
			},
			Secondary: tokens.Ramp{
				hex(0xf7, 0xf5, 0xff),
				hex(0xea, 0xe5, 0xff),
				hex(0xd6, 0xd1, 0xf0),
				hex(0xb8, 0xb2, 0xd1),
				hex(0x9a, 0x95, 0xb3),
				hex(0x7b, 0x76, 0x93),
				hex(0x5e, 0x58, 0x74),
				hex(0x44, 0x3f, 0x59),
				hex(0x15, 0x10, 0x26),
			},
			Tertiary: tokens.Ramp{
				hex(0xff, 0xf4, 0xf8),
				hex(0xff, 0xe0, 0xeb),
				hex(0xfb, 0xc6, 0xda),
				hex(0xdb, 0xa8, 0xbc),
				hex(0xbc, 0x8b, 0x9e),
				hex(0x9b, 0x6c, 0x7f),
				hex(0x7b, 0x4f, 0x61),
				hex(0x5f, 0x35, 0x47),
				hex(0x2a, 0x05, 0x17),
			},
			Error: tokens.Ramp{
				hex(0xff, 0xf4, 0xf2),
				hex(0xff, 0xe2, 0xdd),
				hex(0xff, 0xc7, 0xbf),
				hex(0xff, 0x9b, 0x8d),
				hex(0xfa, 0x6b, 0x5b),
				hex(0xd5, 0x49, 0x3c),
				hex(0xb1, 0x24, 0x1c),
				hex(0x8b, 0x00, 0x02),
				hex(0x32, 0x00, 0x00),
			},
		},
		Primary:     hex(0x67, 0x50, 0xa4), // the seed, byte-exact
		OnPrimary:   tokens.White,
		Secondary:   hex(0x60, 0x5b, 0x76),
		OnSecondary: tokens.White,
		Tertiary:    hex(0x7e, 0x51, 0x63),
		OnTertiary:  tokens.White,
		Error:       hex(0xb4, 0x27, 0x1f),
		OnError:     tokens.White,
		Background:  hex(0xf7, 0xf6, 0xfd),
		Text:        hex(0x14, 0x13, 0x18),
	}
	dark = tokens.ColorTokens{
		Ramps: tokens.RampSet{
			Neutral: tokens.Ramp{
				hex(0x18, 0x17, 0x1c),
				hex(0x22, 0x21, 0x26),
				hex(0x2e, 0x2e, 0x33),
				hex(0x47, 0x46, 0x4c),
				hex(0x9e, 0x9d, 0xa4),
				hex(0xb6, 0xb5, 0xbc),
				hex(0xcc, 0xcb, 0xd2),
				hex(0xdd, 0xdc, 0xe3),
				hex(0xee, 0xed, 0xf4),
			},
			Primary: tokens.Ramp{
				hex(0x22, 0x00, 0x4e),
				hex(0x2c, 0x0b, 0x5d),
				hex(0x38, 0x1c, 0x6c),
				hex(0x50, 0x37, 0x89),
				hex(0xa6, 0x90, 0xea),
				hex(0xbd, 0xaa, 0xff),
				hex(0xd0, 0xc4, 0xff),
				hex(0xdf, 0xd8, 0xff),
				hex(0xef, 0xec, 0xff),
			},
			Secondary: tokens.Ramp{
				hex(0x1a, 0x14, 0x2b),
				hex(0x24, 0x1e, 0x36),
				hex(0x30, 0x2b, 0x43),
				hex(0x49, 0x43, 0x5d),
				hex(0xa0, 0x9a, 0xb8),
				hex(0xb8, 0xb2, 0xd1),
				hex(0xce, 0xc8, 0xe8),
				hex(0xdf, 0xd9, 0xf9),
				hex(0xef, 0xec, 0xff),
			},
			Tertiary: tokens.Ramp{
				hex(0x2f, 0x09, 0x1c),
				hex(0x3b, 0x14, 0x26),
				hex(0x49, 0x21, 0x33),
				hex(0x64, 0x3a, 0x4c),
				hex(0xc2, 0x90, 0xa3),
				hex(0xdb, 0xa8, 0xbc),
				hex(0xf2, 0xbe, 0xd2),
				hex(0xff, 0xd0, 0xe2),
				hex(0xff, 0xe8, 0xf0),
			},
			Error: tokens.Ramp{
				hex(0x3a, 0x00, 0x00),
				hex(0x4d, 0x00, 0x01),
				hex(0x65, 0x00, 0x01),
				hex(0x93, 0x00, 0x03),
				hex(0xff, 0x71, 0x61),
				hex(0xff, 0x9b, 0x8d),
				hex(0xff, 0xbc, 0xb1),
				hex(0xff, 0xd3, 0xcc),
				hex(0xff, 0xe9, 0xe6),
			},
		},
		Primary:     hex(0xd0, 0xc4, 0xff), // L* 82, the step-700 depth (D2.4)
		OnPrimary:   hex(0x22, 0x00, 0x4e),
		Secondary:   hex(0xce, 0xc8, 0xe8),
		OnSecondary: hex(0x1a, 0x14, 0x2b),
		Tertiary:    hex(0xf2, 0xbe, 0xd2),
		OnTertiary:  hex(0x2f, 0x09, 0x1c),
		Error:       hex(0xff, 0xbc, 0xb1),
		OnError:     hex(0x3a, 0x00, 0x00),
		Background:  hex(0x18, 0x17, 0x1c),
		Text:        hex(0xee, 0xed, 0xf4),
	}
	fill := func(t tokens.ColorTokens) tokens.ColorTokens {
		n := t.Ramps.Neutral
		t.Surface = n.Step(200)
		t.Divider = n.Step(300)
		return t
	}
	return fill(light), fill(dark)
}

// hcGolden returns the recorded palette FromSeedHighContrast derives from
// DefaultSeed: the E3.3 high-contrast variant, byte for byte. Landmarks
// that tie the recording to the variant's three widenings: steps 100–600 of
// every ramp are byte-identical to defaultGolden's (the grounds stay), the
// 700 stops sit at the default scale's 900 depth (light #141318 neutral 700
// = defaultGolden's neutral 900; dark #eeedf4 mirrors it) with 800/900
// sliding to the axis ends (light 900 pure black, dark 900 pure white, Text
// following in both modes), Divider is the step-500 strong border rather
// than 300, and the dark pins keep their L* 82 bases while their on-colours
// drop to pure black (tone 0). The light pins are byte-identical to
// defaultGolden's — White already clears the raised Lc ≥ 75 floor — so the
// HC light Primary is still the seed byte-for-byte.
func hcGolden() (light, dark tokens.ColorTokens) {
	hex := func(r, g, b uint8) color.NRGBA { return color.NRGBA{r, g, b, 0xff} }
	black, white := hex(0x00, 0x00, 0x00), hex(0xff, 0xff, 0xff)
	light = tokens.ColorTokens{
		Ramps: tokens.RampSet{
			Neutral: tokens.Ramp{
				hex(0xf7, 0xf6, 0xfd),
				hex(0xe8, 0xe7, 0xef),
				hex(0xd5, 0xd3, 0xdb),
				hex(0xb6, 0xb5, 0xbc),
				hex(0x99, 0x98, 0x9f),
				hex(0x7a, 0x79, 0x7f),
				hex(0x14, 0x13, 0x18),
				hex(0x0b, 0x0a, 0x0f),
				black,
			},
			Primary: tokens.Ramp{
				hex(0xf7, 0xf5, 0xff),
				hex(0xea, 0xe5, 0xff),
				hex(0xd8, 0xce, 0xff),
				hex(0xbd, 0xaa, 0xff),
				hex(0xa0, 0x8b, 0xe5),
				hex(0x82, 0x6c, 0xc2),
				hex(0x1d, 0x00, 0x44),
				hex(0x12, 0x00, 0x2f),
				black,
			},
			Secondary: tokens.Ramp{
				hex(0xf7, 0xf5, 0xff),
				hex(0xea, 0xe5, 0xff),
				hex(0xd6, 0xd1, 0xf0),
				hex(0xb8, 0xb2, 0xd1),
				hex(0x9a, 0x95, 0xb3),
				hex(0x7b, 0x76, 0x93),
				hex(0x15, 0x10, 0x26),
				hex(0x0d, 0x07, 0x1c),
				black,
			},
			Tertiary: tokens.Ramp{
				hex(0xff, 0xf4, 0xf8),
				hex(0xff, 0xe0, 0xeb),
				hex(0xfb, 0xc6, 0xda),
				hex(0xdb, 0xa8, 0xbc),
				hex(0xbc, 0x8b, 0x9e),
				hex(0x9b, 0x6c, 0x7f),
				hex(0x2a, 0x05, 0x17),
				hex(0x1f, 0x00, 0x0f),
				black,
			},
			Error: tokens.Ramp{
				hex(0xff, 0xf4, 0xf2),
				hex(0xff, 0xe2, 0xdd),
				hex(0xff, 0xc7, 0xbf),
				hex(0xff, 0x9b, 0x8d),
				hex(0xfa, 0x6b, 0x5b),
				hex(0xd5, 0x49, 0x3c),
				hex(0x32, 0x00, 0x00),
				hex(0x22, 0x00, 0x00),
				black,
			},
		},
		Primary:     hex(0x67, 0x50, 0xa4), // still the seed, byte-exact
		OnPrimary:   tokens.White,
		Secondary:   hex(0x60, 0x5b, 0x76),
		OnSecondary: tokens.White,
		Tertiary:    hex(0x7e, 0x51, 0x63),
		OnTertiary:  tokens.White,
		Error:       hex(0xb4, 0x27, 0x1f),
		OnError:     tokens.White,
		Background:  hex(0xf7, 0xf6, 0xfd),
		Text:        black,
	}
	dark = tokens.ColorTokens{
		Ramps: tokens.RampSet{
			Neutral: tokens.Ramp{
				hex(0x18, 0x17, 0x1c),
				hex(0x22, 0x21, 0x26),
				hex(0x2e, 0x2e, 0x33),
				hex(0x47, 0x46, 0x4c),
				hex(0x9e, 0x9d, 0xa4),
				hex(0xb6, 0xb5, 0xbc),
				hex(0xee, 0xed, 0xf4),
				hex(0xf7, 0xf6, 0xfd),
				white,
			},
			Primary: tokens.Ramp{
				hex(0x22, 0x00, 0x4e),
				hex(0x2c, 0x0b, 0x5d),
				hex(0x38, 0x1c, 0x6c),
				hex(0x50, 0x37, 0x89),
				hex(0xa6, 0x90, 0xea),
				hex(0xbd, 0xaa, 0xff),
				hex(0xef, 0xec, 0xff),
				hex(0xf7, 0xf5, 0xff),
				white,
			},
			Secondary: tokens.Ramp{
				hex(0x1a, 0x14, 0x2b),
				hex(0x24, 0x1e, 0x36),
				hex(0x30, 0x2b, 0x43),
				hex(0x49, 0x43, 0x5d),
				hex(0xa0, 0x9a, 0xb8),
				hex(0xb8, 0xb2, 0xd1),
				hex(0xef, 0xec, 0xff),
				hex(0xf7, 0xf5, 0xff),
				white,
			},
			Tertiary: tokens.Ramp{
				hex(0x2f, 0x09, 0x1c),
				hex(0x3b, 0x14, 0x26),
				hex(0x49, 0x21, 0x33),
				hex(0x64, 0x3a, 0x4c),
				hex(0xc2, 0x90, 0xa3),
				hex(0xdb, 0xa8, 0xbc),
				hex(0xff, 0xe8, 0xf0),
				hex(0xff, 0xf4, 0xf8),
				white,
			},
			Error: tokens.Ramp{
				hex(0x3a, 0x00, 0x00),
				hex(0x4d, 0x00, 0x01),
				hex(0x65, 0x00, 0x01),
				hex(0x93, 0x00, 0x03),
				hex(0xff, 0x71, 0x61),
				hex(0xff, 0x9b, 0x8d),
				hex(0xff, 0xe9, 0xe6),
				hex(0xff, 0xf4, 0xf2),
				white,
			},
		},
		Primary:     hex(0xd0, 0xc4, 0xff), // L* 82, unchanged from the default
		OnPrimary:   black,
		Secondary:   hex(0xce, 0xc8, 0xe8),
		OnSecondary: black,
		Tertiary:    hex(0xf2, 0xbe, 0xd2),
		OnTertiary:  black,
		Error:       hex(0xff, 0xbc, 0xb1),
		OnError:     black,
		Background:  hex(0x18, 0x17, 0x1c),
		Text:        white,
	}
	fill := func(t tokens.ColorTokens) tokens.ColorTokens {
		n := t.Ramps.Neutral
		t.Surface = n.Step(200)
		t.Divider = n.Step(500) // the HC variant's strong-border divider
		return t
	}
	return fill(light), fill(dark)
}

// TestFromSeedHighContrastGoldenPalette pins the high-contrast variant
// derived from the default seed — both schemes, all five ramps, every pin
// and semantic colour — to the recording in hcGolden.
func TestFromSeedHighContrastGoldenPalette(t *testing.T) {
	wantLight, wantDark := hcGolden()
	light, dark := tokens.FromSeedHighContrast(tokens.DefaultSeed)
	diffTokens(t, "FromSeedHighContrast light", light, wantLight)
	diffTokens(t, "FromSeedHighContrast dark", dark, wantDark)
}

// TestFromSeedHighContrastSharesGrounds verifies the variant widens tone
// separation without moving the grounds or the light pins: steps 100–600 of
// every ramp are byte-identical to FromSeed's, and the light pinned bases —
// the seed-exact Primary included — and their White on-colours carry over
// unchanged.
func TestFromSeedHighContrastSharesGrounds(t *testing.T) {
	seeds := []color.NRGBA{
		tokens.DefaultSeed,
		{0x3b, 0x82, 0xf6, 0xff},
	}
	for _, seed := range seeds {
		light, dark := tokens.FromSeed(seed)
		hcLight, hcDark := tokens.FromSeedHighContrast(seed)
		for i := range namedRamps(light) {
			for step := 100; step <= 600; step += 100 {
				if g, w := namedRamps(hcLight)[i].ramp.Step(step), namedRamps(light)[i].ramp.Step(step); g != w {
					t.Errorf("seed %v: HC light %s.Step(%d) = %v, want the default's %v",
						seed, namedRamps(light)[i].name, step, g, w)
				}
				if g, w := namedRamps(hcDark)[i].ramp.Step(step), namedRamps(dark)[i].ramp.Step(step); g != w {
					t.Errorf("seed %v: HC dark %s.Step(%d) = %v, want the default's %v",
						seed, namedRamps(dark)[i].name, step, g, w)
				}
			}
		}
		lightPins := [][2]color.NRGBA{
			{hcLight.Primary, light.Primary}, {hcLight.OnPrimary, light.OnPrimary},
			{hcLight.Secondary, light.Secondary}, {hcLight.OnSecondary, light.OnSecondary},
			{hcLight.Tertiary, light.Tertiary}, {hcLight.OnTertiary, light.OnTertiary},
			{hcLight.Error, light.Error}, {hcLight.OnError, light.OnError},
		}
		for i, p := range lightPins {
			if p[0] != p[1] {
				t.Errorf("seed %v: HC light pin/on #%d = %v, want the default's %v", seed, i, p[0], p[1])
			}
		}
		want := color.NRGBA{seed.R, seed.G, seed.B, 0xff}
		if hcLight.Primary != want {
			t.Errorf("seed %v: HC light Primary = %v, want the seed byte-for-byte", seed, hcLight.Primary)
		}
	}
}

// diffTokens reports every field and ramp step where got differs from want.
func diffTokens(t *testing.T, scheme string, got, want tokens.ColorTokens) {
	t.Helper()
	if got == want {
		return
	}
	gotRamps, wantRamps := namedRamps(got), namedRamps(want)
	for i := range gotRamps {
		for step := 100; step <= 900; step += 100 {
			if g, w := gotRamps[i].ramp.Step(step), wantRamps[i].ramp.Step(step); g != w {
				t.Errorf("%s: %s.Step(%d) = %v, want %v", scheme, gotRamps[i].name, step, g, w)
			}
		}
	}
	fields := func(c tokens.ColorTokens) map[string]color.NRGBA {
		return map[string]color.NRGBA{
			"Primary": c.Primary, "OnPrimary": c.OnPrimary,
			"Secondary": c.Secondary, "OnSecondary": c.OnSecondary,
			"Tertiary": c.Tertiary, "OnTertiary": c.OnTertiary,
			"Error": c.Error, "OnError": c.OnError,
			"Background": c.Background, "Text": c.Text,
			"Surface": c.Surface, "Divider": c.Divider,
		}
	}
	g, w := fields(got), fields(want)
	for name := range w {
		if g[name] != w[name] {
			t.Errorf("%s: %s = %v, want %v", scheme, name, g[name], w[name])
		}
	}
}

// TestFromSeedGoldenPalette pins the entire palette derived from the
// default seed — both schemes, all five ramps, every pin and semantic
// colour — to the recording in defaultGolden, and verifies DefaultLight
// and DefaultDark are exactly that derivation.
func TestFromSeedGoldenPalette(t *testing.T) {
	wantLight, wantDark := defaultGolden()
	light, dark := tokens.FromSeed(tokens.DefaultSeed)
	diffTokens(t, "FromSeed light", light, wantLight)
	diffTokens(t, "FromSeed dark", dark, wantDark)
	diffTokens(t, "DefaultLight", tokens.DefaultLight, wantLight)
	diffTokens(t, "DefaultDark", tokens.DefaultDark, wantDark)
}

func TestWCAGAAContrast(t *testing.T) {
	schemes := []struct {
		name   string
		tokens tokens.ColorTokens
	}{
		{"DefaultLight", tokens.DefaultLight},
		{"DefaultDark", tokens.DefaultDark},
	}
	for _, s := range schemes {
		for _, p := range tokenPairs(s.tokens) {
			cr := contrastRatio(p.bg, p.fg)
			if cr < wcagAA {
				t.Errorf("%s %s: contrast ratio %.2f:1 < %.1f:1 (WCAG AA)",
					s.name, p.name, cr, wcagAA)
			}
		}
	}
}
