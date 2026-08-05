package system_test

import (
	"context"
	"image/color"
	"testing"
	"time"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/a11y"
	"github.com/vibrantgio/spectrum/system"
	"github.com/vibrantgio/spectrum/tokens"
)

// fakeSource returns successive values from vals on each Read call,
// repeating the last value once the slice is exhausted. Mirrors the
// pattern used in spectrum/a11y/preferences_test.go.
type fakeSource struct {
	vals []system.Appearance
	n    int
}

func (f *fakeSource) Read() (system.Appearance, error) {
	v := f.vals[f.n]
	if f.n < len(f.vals)-1 {
		f.n++
	}
	return v, nil
}

func collect[T any](obs rx.Observable[T]) ([]T, error) {
	var out []T
	err := obs.Subscribe(context.Background(), func(v T, _ error, done bool) {
		if !done {
			out = append(out, v)
		}
	}).Wait()
	return out, err
}

func TestFromSourceEmitsInitialValue(t *testing.T) {
	want := system.Appearance{Dark: true, Accent: system.AccentBlue}
	src := &fakeSource{vals: []system.Appearance{want}}

	got, err := collect(system.FromSource(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 emission, got %d", len(got))
	}
	if got[0] != want {
		t.Errorf("got %+v, want %+v", got[0], want)
	}
}

func TestFromSourceEmitsOnDarkChange(t *testing.T) {
	light := system.Appearance{Dark: false}
	dark := system.Appearance{Dark: true}
	src := &fakeSource{vals: []system.Appearance{light, dark}}

	got, err := collect(system.FromSource(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 emissions, got %d", len(got))
	}
	if got[0] != light {
		t.Errorf("first: got %+v, want %+v", got[0], light)
	}
	if got[1] != dark {
		t.Errorf("second: got %+v, want %+v", got[1], dark)
	}
}

func TestFromSourceDeduplicates(t *testing.T) {
	a := system.Appearance{Dark: false}
	b := system.Appearance{Dark: true}
	src := &fakeSource{vals: []system.Appearance{a, a, b}}

	got, err := collect(system.FromSource(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 emissions (a then b), got %d", len(got))
	}
	if got[0] != a || got[1] != b {
		t.Errorf("got %+v then %+v, want %+v then %+v", got[0], got[1], a, b)
	}
}

func TestFromSourceEmitsOnAccentChange(t *testing.T) {
	a := system.Appearance{Accent: system.AccentBlue}
	b := system.Appearance{Accent: system.AccentRed}
	src := &fakeSource{vals: []system.Appearance{a, b}}

	got, err := collect(system.FromSource(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 emissions, got %d", len(got))
	}
	if got[0].Accent != system.AccentBlue || got[1].Accent != system.AccentRed {
		t.Errorf("accent transitions wrong: %+v then %+v", got[0], got[1])
	}
}

func TestFromSourceThemeBridgesDarkToDarkColors(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{Dark: true}}}

	themes, err := collect(system.FromSourceTheme(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 1 {
		t.Fatalf("expected 1 theme, got %d", len(themes))
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != tokens.DefaultDark {
		t.Errorf("dark appearance must yield DefaultDark; got %+v", colors)
	}
}

func TestFromSourceThemeBridgesLightToLightColors(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{Dark: false}}}

	themes, err := collect(system.FromSourceTheme(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != tokens.DefaultLight {
		t.Errorf("light appearance must yield DefaultLight; got %+v", colors)
	}
}

func TestFromSourceThemeReemitsOnChange(t *testing.T) {
	light := system.Appearance{Dark: false}
	dark := system.Appearance{Dark: true}
	src := &fakeSource{vals: []system.Appearance{light, dark}}

	themes, err := collect(system.FromSourceTheme(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 2 {
		t.Fatalf("expected 2 themes, got %d", len(themes))
	}
	for i, want := range []tokens.ColorTokens{tokens.DefaultLight, tokens.DefaultDark} {
		colors, err := collect(themes[i].Color)
		if err != nil {
			t.Fatalf("theme[%d] color observe: %v", i, err)
		}
		if len(colors) != 1 || colors[0] != want {
			t.Errorf("theme[%d] colors: got %+v, want %+v", i, colors, want)
		}
	}
}

// customSeed is a brand colour distinct from tokens.DefaultSeed, so any
// leak of the default palette into an injected stream is detectable.
var customSeed = color.NRGBA{R: 0x00, G: 0x6E, B: 0x2E, A: 0xff}

func TestFromSourceThemeWithSeedEmitsSeededLight(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{Dark: false}}}
	wantLight, _ := tokens.FromSeed(customSeed)

	themes, err := collect(system.FromSourceTheme(src, time.Hour, system.WithSeed(customSeed)).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 1 {
		t.Fatalf("expected 1 theme, got %d", len(themes))
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != wantLight {
		t.Fatalf("seeded light palette mismatch")
	}
	// ADR-007: the light primary base is the seed pinned byte-exact.
	if colors[0].Primary != customSeed {
		t.Errorf("light Primary must pin the seed byte-exact: got %+v, want %+v", colors[0].Primary, customSeed)
	}
}

func TestFromSourceThemeSeedSurvivesLightToDark(t *testing.T) {
	light := system.Appearance{Dark: false}
	dark := system.Appearance{Dark: true}
	src := &fakeSource{vals: []system.Appearance{light, dark}}
	wantLight, wantDark := tokens.FromSeed(customSeed)

	themes, err := collect(system.FromSourceTheme(src, time.Millisecond, system.WithSeed(customSeed)).Take(2))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 2 {
		t.Fatalf("expected 2 themes, got %d", len(themes))
	}
	for i, want := range []tokens.ColorTokens{wantLight, wantDark} {
		colors, err := collect(themes[i].Color)
		if err != nil {
			t.Fatalf("theme[%d] color observe: %v", i, err)
		}
		if len(colors) != 1 || colors[0] != want {
			t.Errorf("theme[%d]: custom seed did not survive the transition", i)
		}
	}
	// The dark emission is the seed's dark re-tone, not the default dark.
	darkColors, err := collect(themes[1].Color)
	if err != nil {
		t.Fatalf("dark color observe: %v", err)
	}
	if darkColors[0].Primary != wantDark.Primary {
		t.Errorf("dark Primary: got %+v, want the seed's dark pin %+v", darkColors[0].Primary, wantDark.Primary)
	}
	if darkColors[0] == tokens.DefaultDark {
		t.Error("dark emission fell back to DefaultDark; the custom seed was lost")
	}
}

// accentCases pins the accent → seed table independently of the
// implementation: literal Apple HIG system-colour sRGB values, so a silent
// edit to the package's own table fails here. AccentDefault expects the
// default palette (no accent override).
var accentCases = []struct {
	name   string
	accent system.Accent
	seed   color.NRGBA
}{
	{"default", system.AccentDefault, tokens.DefaultSeed},
	{"red", system.AccentRed, color.NRGBA{R: 0xFF, G: 0x3B, B: 0x30, A: 0xFF}},
	{"orange", system.AccentOrange, color.NRGBA{R: 0xFF, G: 0x95, B: 0x00, A: 0xFF}},
	{"yellow", system.AccentYellow, color.NRGBA{R: 0xFF, G: 0xCC, B: 0x00, A: 0xFF}},
	{"green", system.AccentGreen, color.NRGBA{R: 0x28, G: 0xCD, B: 0x41, A: 0xFF}},
	{"blue", system.AccentBlue, color.NRGBA{R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF}},
	{"purple", system.AccentPurple, color.NRGBA{R: 0xAF, G: 0x52, B: 0xDE, A: 0xFF}},
	{"pink", system.AccentPink, color.NRGBA{R: 0xFF, G: 0x2D, B: 0x55, A: 0xFF}},
	{"graphite", system.AccentGraphite, color.NRGBA{R: 0x8E, G: 0x8E, B: 0x93, A: 0xFF}},
}

func TestFromSourceThemeFollowsEachAccent(t *testing.T) {
	for _, tc := range accentCases {
		t.Run(tc.name, func(t *testing.T) {
			src := &fakeSource{vals: []system.Appearance{{Dark: false, Accent: tc.accent}}}
			wantLight, _ := tokens.FromSeed(tc.seed)
			if tc.accent == system.AccentDefault {
				wantLight = tokens.DefaultLight
			}

			themes, err := collect(system.FromSourceTheme(src, time.Hour).Take(1))
			if err != nil {
				t.Fatalf("theme observe: %v", err)
			}
			if len(themes) != 1 {
				t.Fatalf("expected 1 theme, got %d", len(themes))
			}
			colors, err := collect(themes[0].Color)
			if err != nil {
				t.Fatalf("color observe: %v", err)
			}
			if len(colors) != 1 || colors[0] != wantLight {
				t.Fatalf("accent %s: light palette is not FromSeed of its seed", tc.name)
			}
			// ADR-007: the light primary base pins the seed byte-exact, so
			// an accented button matches the OS accent colour exactly.
			if colors[0].Primary != tc.seed {
				t.Errorf("accent %s: light Primary = %+v, want the seed %+v", tc.name, colors[0].Primary, tc.seed)
			}
		})
	}
}

func TestFromSourceThemeReemitsOnAccentChange(t *testing.T) {
	blue := system.Appearance{Dark: false, Accent: system.AccentBlue}
	red := system.Appearance{Dark: false, Accent: system.AccentRed}
	src := &fakeSource{vals: []system.Appearance{blue, red}}

	themes, err := collect(system.FromSourceTheme(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 2 {
		t.Fatalf("expected 2 themes (accent change, same mode), got %d", len(themes))
	}
	wantPins := []color.NRGBA{
		{R: 0x00, G: 0x7A, B: 0xFF, A: 0xFF}, // systemBlue
		{R: 0xFF, G: 0x3B, B: 0x30, A: 0xFF}, // systemRed
	}
	for i, want := range wantPins {
		colors, err := collect(themes[i].Color)
		if err != nil {
			t.Fatalf("theme[%d] color observe: %v", i, err)
		}
		if len(colors) != 1 || colors[0].Primary != want {
			t.Errorf("theme[%d] Primary = %+v, want %+v", i, colors[0].Primary, want)
		}
	}
}

func TestFromSourceThemeWithSeedBeatsAccent(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{Dark: false, Accent: system.AccentRed}}}
	wantLight, _ := tokens.FromSeed(customSeed)

	themes, err := collect(system.FromSourceTheme(src, time.Hour, system.WithSeed(customSeed)).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != wantLight {
		t.Fatalf("WithSeed must beat the OS accent; got a different palette")
	}
	if colors[0].Primary != customSeed {
		t.Errorf("Primary = %+v, want the app's own seed %+v (not the accent)", colors[0].Primary, customSeed)
	}
}

func TestFromSourceThemeWithPaletteBeatsAccent(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{Dark: false, Accent: system.AccentGreen}}}
	customLight, customDark := tokens.FromSeed(customSeed)

	themes, err := collect(system.FromSourceTheme(src, time.Hour, system.WithPalette(customLight, customDark)).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != customLight {
		t.Errorf("WithPalette must beat the OS accent; got a different palette")
	}
}

func TestFromSourceThemeAccentSurvivesDarkMode(t *testing.T) {
	purple := color.NRGBA{R: 0xAF, G: 0x52, B: 0xDE, A: 0xFF}
	src := &fakeSource{vals: []system.Appearance{{Dark: true, Accent: system.AccentPurple}}}
	_, wantDark := tokens.FromSeed(purple)

	themes, err := collect(system.FromSourceTheme(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != wantDark {
		t.Fatalf("dark accent palette is not the accent seed's dark set")
	}
	// The dark Primary is the accent's dark re-tone (ADR-007's dark pin),
	// not the default dark and not the raw seed.
	if colors[0].Primary != wantDark.Primary {
		t.Errorf("dark Primary = %+v, want the accent's dark pin %+v", colors[0].Primary, wantDark.Primary)
	}
	if colors[0] == tokens.DefaultDark {
		t.Error("dark emission fell back to DefaultDark; the accent was lost")
	}
}

// rawAccent is an arbitrary colour of the kind the Windows registry or a
// KDE kdeglobals delivers — deliberately none of the enum accent seeds.
var rawAccent = color.NRGBA{R: 0x00, G: 0x78, B: 0xD7, A: 0xFF} // Windows default blue

func TestFromSourceThemeFollowsAccentSeed(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{Dark: false, AccentSeed: rawAccent, AccentSeedSet: true}}}
	wantLight, _ := tokens.FromSeed(rawAccent)

	themes, err := collect(system.FromSourceTheme(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 1 {
		t.Fatalf("expected 1 theme, got %d", len(themes))
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != wantLight {
		t.Fatalf("light palette is not FromSeed of the raw accent seed")
	}
	// ADR-007: the light primary base pins the seed byte-exact, so an
	// accented button matches the OS accent colour exactly.
	if colors[0].Primary != rawAccent {
		t.Errorf("light Primary = %+v, want the raw seed %+v", colors[0].Primary, rawAccent)
	}
}

func TestFromSourceThemeAccentSeedSurvivesDarkMode(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{Dark: true, AccentSeed: rawAccent, AccentSeedSet: true}}}
	_, wantDark := tokens.FromSeed(rawAccent)

	themes, err := collect(system.FromSourceTheme(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != wantDark {
		t.Fatalf("dark palette is not the raw seed's dark set")
	}
	if colors[0] == tokens.DefaultDark {
		t.Error("dark emission fell back to DefaultDark; the raw accent was lost")
	}
}

func TestFromSourceThemeAccentSeedBeatsEnumAccent(t *testing.T) {
	// A source that (hypothetically) sets both shapes: the raw seed wins.
	src := &fakeSource{vals: []system.Appearance{{
		Accent:        system.AccentRed,
		AccentSeed:    rawAccent,
		AccentSeedSet: true,
	}}}
	wantLight, _ := tokens.FromSeed(rawAccent)

	themes, err := collect(system.FromSourceTheme(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != wantLight {
		t.Fatalf("AccentSeed must beat the enum accent")
	}
	if colors[0].Primary != rawAccent {
		t.Errorf("Primary = %+v, want the raw seed %+v (not systemRed)", colors[0].Primary, rawAccent)
	}
}

func TestFromSourceThemeUnsetAccentSeedIgnored(t *testing.T) {
	// AccentSeed without AccentSeedSet carries no meaning: the default
	// palette holds. Guards against a source leaving a stale colour behind.
	src := &fakeSource{vals: []system.Appearance{{AccentSeed: rawAccent, AccentSeedSet: false}}}

	themes, err := collect(system.FromSourceTheme(src, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != tokens.DefaultLight {
		t.Errorf("unset AccentSeed must leave the default palette; got %+v", colors)
	}
}

func TestFromSourceThemeWithSeedBeatsAccentSeed(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{AccentSeed: rawAccent, AccentSeedSet: true}}}
	wantLight, _ := tokens.FromSeed(customSeed)

	themes, err := collect(system.FromSourceTheme(src, time.Hour, system.WithSeed(customSeed)).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != wantLight {
		t.Fatalf("WithSeed must beat the OS AccentSeed")
	}
	if colors[0].Primary != customSeed {
		t.Errorf("Primary = %+v, want the app's own seed %+v (not the OS colour)", colors[0].Primary, customSeed)
	}
}

func TestFromSourceThemeWithPaletteBeatsAccentSeed(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{AccentSeed: rawAccent, AccentSeedSet: true}}}
	customLight, customDark := tokens.FromSeed(customSeed)

	themes, err := collect(system.FromSourceTheme(src, time.Hour, system.WithPalette(customLight, customDark)).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != customLight {
		t.Errorf("WithPalette must beat the OS AccentSeed; got a different palette")
	}
}

func TestFromSourceEmitsOnAccentSeedChange(t *testing.T) {
	a := system.Appearance{AccentSeed: rawAccent, AccentSeedSet: true}
	b := system.Appearance{AccentSeed: color.NRGBA{R: 0xE6, G: 0x2D, B: 0x42, A: 0xFF}, AccentSeedSet: true}
	src := &fakeSource{vals: []system.Appearance{a, a, b}}

	got, err := collect(system.FromSource(src, time.Millisecond).Take(2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 || got[0] != a || got[1] != b {
		t.Errorf("seed transitions wrong: got %+v", got)
	}
}

func TestFromSourceThemeWithPaletteSurvivesLightToDark(t *testing.T) {
	src := &fakeSource{vals: []system.Appearance{{Dark: false}, {Dark: true}}}
	customLight, customDark := tokens.FromSeed(customSeed)

	themes, err := collect(system.FromSourceTheme(src, time.Millisecond, system.WithPalette(customLight, customDark)).Take(2))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 2 {
		t.Fatalf("expected 2 themes, got %d", len(themes))
	}
	for i, want := range []tokens.ColorTokens{customLight, customDark} {
		colors, err := collect(themes[i].Color)
		if err != nil {
			t.Fatalf("theme[%d] color observe: %v", i, err)
		}
		if len(colors) != 1 || colors[0] != want {
			t.Errorf("theme[%d]: injected palette did not survive the transition", i)
		}
	}
}

// --- E3.2: accessibility preferences composed into the theme ---

// fakeA11ySource returns successive values from vals on each Read call,
// repeating the last value once the slice is exhausted — the a11y twin of
// fakeSource above.
type fakeA11ySource struct {
	vals []a11y.A11yPrefs
	n    int
}

func (f *fakeA11ySource) Read() (a11y.A11yPrefs, error) {
	v := f.vals[f.n]
	if f.n < len(f.vals)-1 {
		f.n++
	}
	return v, nil
}

// TestFromSourceThemeReduceMotionSnaps is E3.2's snap test: under an OS
// reduce-motion preference the emitted motion scale has every duration at
// zero, so an animated component that derives its frame count from the
// scale — pulse/motion's FramesAt(d, fps) = round(d·fps) = 0 frames for
// every stop — is at its target on the first frame it draws.
func TestFromSourceThemeReduceMotionSnaps(t *testing.T) {
	appearance := &fakeSource{vals: []system.Appearance{{}}}
	prefs := &fakeA11ySource{vals: []a11y.A11yPrefs{{ReduceMotion: true}}}

	themes, err := collect(system.FromSourceTheme(appearance, time.Hour, system.WithA11ySource(prefs)).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 1 {
		t.Fatalf("expected 1 theme, got %d", len(themes))
	}
	motions, err := collect(themes[0].Motion)
	if err != nil {
		t.Fatalf("motion observe: %v", err)
	}
	if len(motions) != 1 {
		t.Fatalf("expected 1 motion emission, got %d", len(motions))
	}
	m := motions[0]
	for _, d := range []struct {
		name string
		v    time.Duration
	}{
		{"DurXFast", m.DurXFast}, {"DurFast", m.DurFast}, {"DurNormal", m.DurNormal},
		{"DurSlow", m.DurSlow}, {"DurXSlow", m.DurXSlow},
	} {
		if d.v != 0 {
			t.Errorf("ReduceMotion: %s = %v, want 0 (animations must reach their target in one frame)", d.name, d.v)
		}
	}
	if m != tokens.Motion.Reduced() {
		t.Errorf("ReduceMotion: motion scale is not tokens.Motion.Reduced():\ngot %+v", m)
	}
}

func TestFromSourceThemeReduceMotionOffKeepsMotion(t *testing.T) {
	appearance := &fakeSource{vals: []system.Appearance{{}}}
	prefs := &fakeA11ySource{vals: []a11y.A11yPrefs{{}}}

	themes, err := collect(system.FromSourceTheme(appearance, time.Hour, system.WithA11ySource(prefs)).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	motions, err := collect(themes[0].Motion)
	if err != nil {
		t.Fatalf("motion observe: %v", err)
	}
	if len(motions) != 1 || motions[0] != tokens.Motion {
		t.Errorf("all-off prefs: motion scale must be tokens.Motion unchanged; got %+v", motions)
	}
}

func TestFromSourceThemeDefaultA11yIsHermetic(t *testing.T) {
	// Without WithA11ySource, FromSourceTheme must not read the machine's
	// real accessibility preferences: the default is a constant all-off
	// stream, so this asserts the full default emission regardless of host.
	appearance := &fakeSource{vals: []system.Appearance{{}}}

	themes, err := collect(system.FromSourceTheme(appearance, time.Hour).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	motions, err := collect(themes[0].Motion)
	if err != nil {
		t.Fatalf("motion observe: %v", err)
	}
	if len(motions) != 1 || motions[0] != tokens.Motion {
		t.Errorf("default a11y stream must be all-off: got motion %+v", motions)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != tokens.DefaultLight {
		t.Errorf("default a11y stream must be all-off: got colors %+v", colors)
	}
}

func TestFromSourceThemeReduceMotionToggleReemits(t *testing.T) {
	// A preference change alone — the appearance never changes — re-emits
	// the theme, normal motion first, reduced second.
	appearance := &fakeSource{vals: []system.Appearance{{}}}
	prefs := &fakeA11ySource{vals: []a11y.A11yPrefs{{}, {ReduceMotion: true}}}

	themes, err := collect(system.FromSourceTheme(appearance, time.Millisecond, system.WithA11ySource(prefs)).Take(2))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if len(themes) != 2 {
		t.Fatalf("expected 2 themes (one per preference value), got %d", len(themes))
	}
	for i, want := range []tokens.MotionScale{tokens.Motion, tokens.Motion.Reduced()} {
		motions, err := collect(themes[i].Motion)
		if err != nil {
			t.Fatalf("theme[%d] motion observe: %v", i, err)
		}
		if len(motions) != 1 || motions[0] != want {
			t.Errorf("theme[%d] motion: got %+v, want %+v", i, motions, want)
		}
	}
}

func TestFromSourceThemeReduceMotionComposesOnSeededPalette(t *testing.T) {
	// A11y composes ON TOP of palette precedence: reduce motion affects the
	// Motion emission regardless of the palette choice, and WithSeed keeps
	// deciding the colors.
	appearance := &fakeSource{vals: []system.Appearance{{}}}
	prefs := &fakeA11ySource{vals: []a11y.A11yPrefs{{ReduceMotion: true}}}
	wantLight, _ := tokens.FromSeed(customSeed)

	themes, err := collect(system.FromSourceTheme(appearance, time.Hour,
		system.WithSeed(customSeed), system.WithA11ySource(prefs)).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	motions, err := collect(themes[0].Motion)
	if err != nil {
		t.Fatalf("motion observe: %v", err)
	}
	if len(motions) != 1 || motions[0] != tokens.Motion.Reduced() {
		t.Errorf("reduce motion must apply with a branded palette; got %+v", motions)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != wantLight {
		t.Errorf("WithSeed palette must survive a11y composition; got %+v", colors)
	}
}

func TestFromSourceThemeHighContrastDefaultIsIdentity(t *testing.T) {
	// Until E3.3 lands the derivation, HighContrastVariant defaults to the
	// identity: high contrast on, palette unchanged — the wiring is live
	// but selects the default palette.
	appearance := &fakeSource{vals: []system.Appearance{{}}}
	prefs := &fakeA11ySource{vals: []a11y.A11yPrefs{{HighContrast: true}}}

	themes, err := collect(system.FromSourceTheme(appearance, time.Hour, system.WithA11ySource(prefs)).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != tokens.DefaultLight {
		t.Errorf("identity HighContrastVariant must leave the default palette; got %+v", colors)
	}
}

func TestFromSourceThemeHighContrastSelectsVariantOfChosenPalette(t *testing.T) {
	// The E3.3 hook: HighContrastVariant receives the pair that palette
	// precedence resolved — here WithSeed's pair, not the defaults — and
	// its result is what Color emits, on the dark side under Dark.
	appearance := &fakeSource{vals: []system.Appearance{{Dark: true}}}
	prefs := &fakeA11ySource{vals: []a11y.A11yPrefs{{HighContrast: true}}}
	seededLight, seededDark := tokens.FromSeed(customSeed)
	hcLight, hcDark := tokens.FromSeed(rawAccent) // stand-in "hc variant" pair

	var gotLight, gotDark tokens.ColorTokens
	orig := system.HighContrastVariant
	system.HighContrastVariant = func(light, dark tokens.ColorTokens) (tokens.ColorTokens, tokens.ColorTokens) {
		gotLight, gotDark = light, dark
		return hcLight, hcDark
	}
	defer func() { system.HighContrastVariant = orig }()

	themes, err := collect(system.FromSourceTheme(appearance, time.Hour,
		system.WithSeed(customSeed), system.WithA11ySource(prefs)).Take(1))
	if err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if gotLight != seededLight || gotDark != seededDark {
		t.Errorf("hook must receive the resolved (seeded) pair, not the defaults")
	}
	colors, err := collect(themes[0].Color)
	if err != nil {
		t.Fatalf("color observe: %v", err)
	}
	if len(colors) != 1 || colors[0] != hcDark {
		t.Errorf("dark + high contrast must emit the hook's dark variant; got %+v", colors)
	}
}

func TestFromSourceThemeHighContrastOffSkipsHook(t *testing.T) {
	appearance := &fakeSource{vals: []system.Appearance{{}}}
	prefs := &fakeA11ySource{vals: []a11y.A11yPrefs{{}}}

	called := false
	orig := system.HighContrastVariant
	system.HighContrastVariant = func(light, dark tokens.ColorTokens) (tokens.ColorTokens, tokens.ColorTokens) {
		called = true
		return light, dark
	}
	defer func() { system.HighContrastVariant = orig }()

	if _, err := collect(system.FromSourceTheme(appearance, time.Hour, system.WithA11ySource(prefs)).Take(1)); err != nil {
		t.Fatalf("theme observe: %v", err)
	}
	if called {
		t.Error("HighContrastVariant must not run while the preference is off")
	}
}
