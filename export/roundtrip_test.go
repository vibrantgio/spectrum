package export

import (
	"encoding/json"
	"fmt"
	stdcolor "image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/color"
	"github.com/vibrantgio/spectrum/theme"
	"github.com/vibrantgio/spectrum/tokens"
)

// parseSheet is a tolerant reader of the emitted styles.css: it returns the
// custom properties per selector block, ignoring anything that is not a
// block opener, a block closer or a --declaration.
func parseSheet(t *testing.T, src string) map[string]map[string]string {
	t.Helper()
	blocks := map[string]map[string]string{}
	var cur map[string]string
	for _, line := range strings.Split(src, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasSuffix(line, "{"):
			sel := strings.TrimSpace(strings.TrimSuffix(line, "{"))
			if _, dup := blocks[sel]; dup {
				t.Fatalf("styles.css: duplicate block %q", sel)
			}
			cur = map[string]string{}
			blocks[sel] = cur
		case line == "}":
			cur = nil
		case strings.HasPrefix(line, "--"):
			if cur == nil {
				t.Fatalf("styles.css: declaration outside a block: %q", line)
			}
			name, val, ok := strings.Cut(line, ":")
			if !ok || !strings.HasSuffix(val, ";") {
				t.Fatalf("styles.css: malformed declaration: %q", line)
			}
			name = strings.TrimSpace(name)
			if _, dup := cur[name]; dup {
				t.Fatalf("styles.css: duplicate declaration %q", name)
			}
			cur[name] = strings.TrimSpace(strings.TrimSuffix(val, ";"))
		}
	}
	return blocks
}

// wantHex formats a colour the way the sheet must: lowercase #rrggbb.
// Deliberately written out rather than shared with the implementation, so
// the test and the serialiser cannot drift together.
func wantHex(c stdcolor.NRGBA) string {
	const digits = "0123456789abcdef"
	return string([]byte{'#',
		digits[c.R>>4], digits[c.R&0xf],
		digits[c.G>>4], digits[c.G&0xf],
		digits[c.B>>4], digits[c.B&0xf]})
}

// wantPx parses a px length back to its float32 value.
func wantPx(t *testing.T, name, v string) float32 {
	t.Helper()
	num, ok := strings.CutSuffix(v, "px")
	if !ok {
		t.Fatalf("%s: value %q is not a px length", name, v)
	}
	f, err := strconv.ParseFloat(num, 32)
	if err != nil {
		t.Fatalf("%s: value %q: %v", name, v, err)
	}
	return float32(f)
}

func writeDefault(t *testing.T) (Snapshot, map[string]map[string]string, []byte) {
	t.Helper()
	snap, err := Capture(theme.Default())
	if err != nil {
		t.Fatalf("Capture(theme.Default()): %v", err)
	}
	dir := t.TempDir()
	if err := Write(dir, snap); err != nil {
		t.Fatalf("Write: %v", err)
	}
	css, err := os.ReadFile(filepath.Join(dir, "styles.css"))
	if err != nil {
		t.Fatal(err)
	}
	js, err := os.ReadFile(filepath.Join(dir, "theme.json"))
	if err != nil {
		t.Fatal(err)
	}
	return snap, parseSheet(t, string(css)), js
}

// TestRoundTripColors parses the emitted CSS back and asserts every colour
// variable in both blocks equals the Go token it came from.
func TestRoundTripColors(t *testing.T) {
	snap, sheet, _ := writeDefault(t)
	root, dark := sheet[":root"], sheet[".dark"]
	if root == nil || dark == nil {
		t.Fatalf("styles.css must carry a :root and a .dark block; got %v", len(sheet))
	}

	schemes := []struct {
		vars   map[string]string
		tokens tokens.ColorTokens
	}{{root, snap.Light}, {dark, snap.Dark}}

	for _, scheme := range schemes {
		for _, role := range rampRoles {
			ramp := role.ramp(scheme.tokens.Ramps)
			for step := 100; step <= 900; step += 100 {
				name := fmt.Sprintf("--color-%s-%d", role.name, step)
				if got, want := scheme.vars[name], wantHex(ramp.Step(step)); got != want {
					t.Errorf("%s = %q, want %q", name, got, want)
				}
			}
		}
		for _, pin := range pinRoles {
			name := "--color-" + pin.name
			if got, want := scheme.vars[name], wantHex(pin.pick(scheme.tokens)); got != want {
				t.Errorf("%s = %q, want %q", name, got, want)
			}
		}
	}

	// The dark block carries exactly the colour overrides — every variable
	// it declares must exist in :root, and nothing but colours may differ
	// per mode.
	for name := range dark {
		if _, ok := root[name]; !ok {
			t.Errorf(".dark declares %s which :root does not", name)
		}
		if !strings.HasPrefix(name, "--color-") {
			t.Errorf(".dark declares non-colour variable %s", name)
		}
	}
	if want := len(rampRoles)*9 + len(pinRoles); len(dark) != want {
		t.Errorf(".dark declares %d variables, want %d", len(dark), want)
	}
}

// TestRoundTripScales asserts the font, spacing, radius and shadow
// variables all parse back to the Go values they came from.
func TestRoundTripScales(t *testing.T) {
	snap, sheet, _ := writeDefault(t)
	root := sheet[":root"]

	if got, err := strconv.Unquote(root["--font-family"]); err != nil || got != snap.Typography.BodyLarge.Typeface {
		t.Errorf("--font-family = %q (%v), want quoted %q", root["--font-family"], err, snap.Typography.BodyLarge.Typeface)
	}
	for _, role := range typeRoles {
		style := role.pick(snap.Typography)
		base := "--font-" + role.name
		if got := wantPx(t, base+"-size", root[base+"-size"]); got != style.Size {
			t.Errorf("%s-size = %v, want %v", base, got, style.Size)
		}
		if got := wantPx(t, base+"-line-height", root[base+"-line-height"]); got != style.LineHeight {
			t.Errorf("%s-line-height = %v, want %v", base, got, style.LineHeight)
		}
		if got, err := strconv.Atoi(root[base+"-weight"]); err != nil || got != style.Weight {
			t.Errorf("%s-weight = %q, want %d", base, root[base+"-weight"], style.Weight)
		}
		if got := wantPx(t, base+"-tracking", root[base+"-tracking"]); got != style.Tracking {
			t.Errorf("%s-tracking = %v, want %v", base, got, style.Tracking)
		}
	}

	for _, key := range spaceKeys {
		name := "--space-" + key.name
		if got := wantPx(t, name, root[name]); got != key.pick(snap.Spacing) {
			t.Errorf("%s = %v, want %v", name, got, key.pick(snap.Spacing))
		}
	}
	for _, key := range radiusKeys {
		name := "--radius-" + key.name
		if got := wantPx(t, name, root[name]); got != key.pick(snap.Radius) {
			t.Errorf("%s = %v, want %v", name, got, key.pick(snap.Radius))
		}
	}

	for _, level := range elevationLevels {
		name, dp := "--shadow-"+level.name, level.pick(snap.Elevation)
		v := root[name]
		if dp == 0 {
			if v != "none" {
				t.Errorf("%s = %q, want \"none\" at depth 0", name, v)
			}
			continue
		}
		mid, ok := strings.CutPrefix(v, "0 ")
		if !ok {
			t.Errorf("%s = %q: want a y-offset shadow with no x-offset", name, v)
			continue
		}
		mid, ok = strings.CutSuffix(mid, " 0 rgba(0, 0, 0, 0.2)")
		if !ok {
			t.Errorf("%s = %q: want no spread and black at 20%%", name, v)
			continue
		}
		lengths := strings.Fields(mid)
		if len(lengths) != 2 {
			t.Errorf("%s = %q: want a y-offset and a blur", name, v)
			continue
		}
		y, blur := wantPx(t, name, lengths[0]), wantPx(t, name, lengths[1])
		if y != dp || blur != 2*dp {
			t.Errorf("%s = %q: y %v blur %v, want dp %v and 2dp", name, v, y, blur, dp)
		}
	}
}

// TestThemeJSONReproduces asserts theme.json's reproducibility claim: its
// seed alone regenerates the exported palette through FromSeed, and every
// recorded parameter matches the tokens and the sheet.
func TestThemeJSONReproduces(t *testing.T) {
	snap, sheet, js := writeDefault(t)
	var p Parameters
	if err := json.Unmarshal(js, &p); err != nil {
		t.Fatalf("theme.json: %v", err)
	}

	var r, g, b uint8
	if _, err := fmt.Sscanf(p.Seed, "#%02x%02x%02x", &r, &g, &b); err != nil {
		t.Fatalf("theme.json seed %q: %v", p.Seed, err)
	}
	seed := stdcolor.NRGBA{R: r, G: g, B: b, A: 0xff}
	if seed != tokens.DefaultSeed {
		t.Errorf("seed = %q, want the default seed %s", p.Seed, wantHex(tokens.DefaultSeed))
	}
	light, dark := tokens.FromSeed(seed)
	if light != snap.Light || dark != snap.Dark {
		t.Errorf("FromSeed(theme.json seed) does not reproduce the exported palette")
	}

	_, chroma, hue := color.OKLChFromNRGBA(seed)
	if diff := p.Hue - hue; diff < -0.005 || diff > 0.005 {
		t.Errorf("hue = %v, want %v within 0.005", p.Hue, hue)
	}
	if diff := p.Sat - chroma; diff < -0.00005 || diff > 0.00005 {
		t.Errorf("sat = %v, want %v within 0.00005", p.Sat, chroma)
	}

	root, darkVars := sheet[":root"], sheet[".dark"]
	for _, mode := range []struct {
		pins Pins
		vars map[string]string
	}{{p.Pins.Light, root}, {p.Pins.Dark, darkVars}} {
		checks := []struct{ name, got string }{
			{"--color-bg", mode.pins.Bg},
			{"--color-text", mode.pins.Text},
			{"--color-accent", mode.pins.Accent},
			{"--color-secondary", mode.pins.Secondary},
			{"--color-tertiary", mode.pins.Tertiary},
			{"--color-error", mode.pins.Error},
		}
		for _, c := range checks {
			if c.got != mode.vars[c.name] {
				t.Errorf("theme.json pin %s = %q, sheet says %q", c.name, c.got, mode.vars[c.name])
			}
		}
	}

	if p.Fonts.Heading != "Roboto" || p.Fonts.Body != "Roboto" {
		t.Errorf("fonts = %+v, want Roboto/Roboto", p.Fonts)
	}
	if p.Radius != float64(snap.Radius.Base) {
		t.Errorf("radius = %v, want the base radius %v", p.Radius, snap.Radius.Base)
	}
	if want := [9]int{97, 92, 85, 74, 63, 51, 39, 28, 6}; p.Scale.Light != want {
		t.Errorf("scale.light = %v, want ADR-007's shared scale %v", p.Scale.Light, want)
	}
	if want := [9]int{8, 13, 19, 30, 65, 74, 82, 88, 94}; p.Scale.Dark != want {
		t.Errorf("scale.dark = %v, want the paired dark scale %v", p.Scale.Dark, want)
	}
}

// TestCaptureRejectsIrreproducible asserts Capture refuses inputs
// theme.json could not honestly reproduce.
func TestCaptureRejectsIrreproducible(t *testing.T) {
	if _, err := Capture(theme.Theme{}); err == nil {
		t.Error("Capture of a zero Theme (nil observables) must error")
	}

	th := theme.Default()
	th.Color = rx.Of(tokens.DefaultDark) // a dark scheme: its Primary pin is not the seed
	if _, err := Capture(th); err == nil {
		t.Error("Capture of a dark colour emission must error: FromSeed(pin) cannot reproduce it")
	}
}

// TestCaptureCustomSeed asserts a re-branded light scheme captures with its
// own seed recovered.
func TestCaptureCustomSeed(t *testing.T) {
	seed := stdcolor.NRGBA{R: 0x00, G: 0x68, B: 0x74, A: 0xff}
	light, dark := tokens.FromSeed(seed)
	th := theme.Default()
	th.Color = rx.Of(light)
	snap, err := Capture(th)
	if err != nil {
		t.Fatalf("Capture: %v", err)
	}
	if snap.Seed != seed {
		t.Errorf("Seed = %v, want %v", snap.Seed, seed)
	}
	if snap.Dark != dark {
		t.Errorf("Dark scheme is not FromSeed(seed)'s pair")
	}
}
