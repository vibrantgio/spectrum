package export

import (
	"fmt"
	stdcolor "image/color"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/vibrantgio/spectrum/color"
	"github.com/vibrantgio/spectrum/theme"
)

// pageFiles are the foundation pages Write must emit, relative to the
// project root.
var pageFiles = []string{
	filepath.Join("foundations", "color.html"),
	filepath.Join("foundations", "type.html"),
	filepath.Join("foundations", "layout.html"),
}

// writeProject writes the default theme's full project into a temp dir and
// returns the snapshot, the parsed sheet and the page sources by name.
func writeProject(t *testing.T) (Snapshot, map[string]map[string]string, map[string]string) {
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
	pages := map[string]string{}
	for _, name := range append([]string{"readme.md"}, pageFiles...) {
		src, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("Write must emit %s: %v", name, err)
		}
		if len(src) == 0 {
			t.Fatalf("%s is empty", name)
		}
		pages[name] = string(src)
	}
	return snap, parseSheet(t, string(css)), pages
}

var varRefRE = regexp.MustCompile(`var\((--[a-zA-Z0-9-]+)\)`)

// TestPageVarClosure asserts every var() reference in every page resolves in
// the emitted sheet's :root block — a page can never name a token the sheet
// does not declare — and that every referenced colour variable is also
// overridden in .dark, so the toggle restyles all of them.
func TestPageVarClosure(t *testing.T) {
	_, sheet, pages := writeProject(t)
	root, dark := sheet[":root"], sheet[".dark"]
	for _, name := range pageFiles {
		refs := varRefRE.FindAllStringSubmatch(pages[name], -1)
		if len(refs) == 0 {
			t.Errorf("%s references no token variables at all", name)
		}
		for _, ref := range refs {
			v := ref[1]
			if _, ok := root[v]; !ok {
				t.Errorf("%s references %s, which styles.css :root does not declare", name, v)
			}
			if strings.HasPrefix(v, "--color-") {
				if _, ok := dark[v]; !ok {
					t.Errorf("%s references colour %s, which the .dark block does not override", name, v)
				}
			}
		}
	}
}

var (
	styleBlockRE = regexp.MustCompile(`(?s)<style>(.*?)</style>`)
	styleAttrRE  = regexp.MustCompile(`style="([^"]*)"`)
	hexLitRE     = regexp.MustCompile(`#[0-9a-fA-F]{3,8}\b`)
	pxLitRE      = regexp.MustCompile(`\d(?:\.\d+)?px\b`)
)

// TestPagesNoHardCodedTokenValues enforces the task's rule: the pages read
// only from the token sheet. In every style context — <style> blocks and
// style attributes — no literal hex colour and no literal px length may
// appear; token values reach styling exclusively through var() references.
// (Hexes and px numbers in annotation text are reader-facing data, not
// styling, and are out of scope by construction.)
func TestPagesNoHardCodedTokenValues(t *testing.T) {
	_, _, pages := writeProject(t)
	for _, name := range pageFiles {
		src := pages[name]
		var contexts []string
		for _, m := range styleBlockRE.FindAllStringSubmatch(src, -1) {
			contexts = append(contexts, m[1])
		}
		attrs := styleAttrRE.FindAllStringSubmatch(src, -1)
		if len(attrs) == 0 {
			t.Errorf("%s has no style attributes; the specimens are expected to be var()-styled inline", name)
		}
		for _, m := range attrs {
			contexts = append(contexts, m[1])
		}
		for _, css := range contexts {
			if hit := hexLitRE.FindString(css); hit != "" {
				t.Errorf("%s: hard-coded colour %q in a style context", name, hit)
			}
			if hit := pxLitRE.FindString(css); hit != "" {
				t.Errorf("%s: hard-coded length %q in a style context", name, hit)
			}
		}
	}
}

// TestPagesDarkToggle asserts each page links the shared sheet relatively
// and carries the light/dark toggle script flipping .dark on the root
// element — the mechanism the dark render confirmation rests on.
func TestPagesDarkToggle(t *testing.T) {
	_, _, pages := writeProject(t)
	for _, name := range pageFiles {
		src := pages[name]
		if !strings.Contains(src, `<link rel="stylesheet" href="../styles.css">`) {
			t.Errorf("%s does not link ../styles.css", name)
		}
		if !strings.Contains(src, `classList.toggle("dark")`) {
			t.Errorf("%s has no .dark toggle script", name)
		}
	}
}

// wantRow renders a contrast table row the way the colour page must,
// written out independently so the page and the test cannot drift together:
// APCA Lc (signed, one decimal) and the WCAG 2 ratio (two decimals), light
// then dark.
func wantRow(label string, lt, lg, dt, dg stdcolor.NRGBA) string {
	return fmt.Sprintf(`<tr><th scope="row">%s</th><td>%.1f</td><td>%.2f:1</td><td>%.1f</td><td>%.2f:1</td></tr>`,
		label,
		color.APCA(lt, lg), color.ContrastRatio(lt, lg),
		color.APCA(dt, dg), color.ContrastRatio(dt, dg))
}

// TestColorPageAnnotatesContrast asserts the colour page carries the
// measured APCA Lc and WCAG ratio, in both modes, for every gated text pair:
// the four ramp pairs per role (900/700 on 100/200) and each role's pinned
// pair.
func TestColorPageAnnotatesContrast(t *testing.T) {
	snap, _, pages := writeProject(t)
	src := pages[filepath.Join("foundations", "color.html")]

	for _, role := range rampRoles {
		light, dark := role.ramp(snap.Light.Ramps), role.ramp(snap.Dark.Ramps)
		for _, pair := range [][2]int{{900, 100}, {900, 200}, {700, 100}, {700, 200}} {
			text, ground := pair[0], pair[1]
			row := wantRow(fmt.Sprintf("%d on %d", text, ground),
				light.Step(text), light.Step(ground), dark.Step(text), dark.Step(ground))
			if !strings.Contains(src, row) {
				t.Errorf("color.html lacks the measured row for %s %d on %d:\n%s", role.name, text, ground, row)
			}
		}
	}

	pinPairs := []struct {
		label  string
		lt, lg stdcolor.NRGBA
		dt, dg stdcolor.NRGBA
	}{
		{"text on bg", snap.Light.Text, snap.Light.Background, snap.Dark.Text, snap.Dark.Background},
		{"on-accent on accent", snap.Light.OnPrimary, snap.Light.Primary, snap.Dark.OnPrimary, snap.Dark.Primary},
		{"on-secondary on secondary", snap.Light.OnSecondary, snap.Light.Secondary, snap.Dark.OnSecondary, snap.Dark.Secondary},
		{"on-tertiary on tertiary", snap.Light.OnTertiary, snap.Light.Tertiary, snap.Dark.OnTertiary, snap.Dark.Tertiary},
		{"on-error on error", snap.Light.OnError, snap.Light.Error, snap.Dark.OnError, snap.Dark.Error},
	}
	for _, p := range pinPairs {
		row := wantRow(p.label, p.lt, p.lg, p.dt, p.dg)
		if !strings.Contains(src, row) {
			t.Errorf("color.html lacks the measured pin row %q:\n%s", p.label, row)
		}
	}
}

// TestColorPageAnnotatesBothModeValues spot-checks that swatch annotations
// carry both modes' hexes, labelled, since text cannot flip with the class.
func TestColorPageAnnotatesBothModeValues(t *testing.T) {
	snap, _, pages := writeProject(t)
	src := pages[filepath.Join("foundations", "color.html")]
	for _, role := range rampRoles {
		light, dark := role.ramp(snap.Light.Ramps), role.ramp(snap.Dark.Ramps)
		for step := 100; step <= 900; step += 100 {
			want := fmt.Sprintf("L %s · D %s", wantHex(light.Step(step)), wantHex(dark.Step(step)))
			if !strings.Contains(src, want) {
				t.Errorf("color.html lacks the dual-mode annotation for %s-%d: %q", role.name, step, want)
			}
		}
	}
}

// TestReadmeNamesFamilies asserts readme.md names every token family the
// sheet emits, deriving the expected mentions from the same tables the
// emitter renders from.
func TestReadmeNamesFamilies(t *testing.T) {
	snap, sheet, pages := writeProject(t)
	readme := pages["readme.md"]

	var want []string
	for _, role := range rampRoles {
		want = append(want, "--color-"+role.name+"-100", "--color-"+role.name+"-900")
	}
	for _, pin := range pinRoles {
		want = append(want, "--color-"+pin.name)
	}
	want = append(want, "--font-family", "-size", "-line-height", "-weight", "-tracking")
	for _, role := range typeRoles {
		want = append(want, role.name)
	}
	for _, key := range spaceKeys {
		want = append(want, "--space-"+key.name)
	}
	for _, key := range radiusKeys {
		want = append(want, "--radius-"+key.name)
	}
	for _, level := range elevationLevels {
		want = append(want, "--shadow-"+level.name)
	}
	want = append(want,
		wantHex(snap.Seed),
		"styles.css", "theme.json",
		"foundations/color.html", "foundations/type.html", "foundations/layout.html",
	)
	for _, w := range want {
		if !strings.Contains(readme, w) {
			t.Errorf("readme.md does not mention %q", w)
		}
	}

	// Paranoia in the other direction: every variable the sheet actually
	// emits must be documented, so a new family cannot ship unnamed. Each
	// variable maps to the string the readme must contain for it.
	rampStepRE := regexp.MustCompile(`^(--color-[a-z]+)-[1-9]00$`)
	for name := range sheet[":root"] {
		mention := name // pins, --space-*, --radius-*, --shadow-*: listed in full
		if m := rampStepRE.FindStringSubmatch(name); m != nil && !isPinName(name) {
			mention = m[1] + "-100" // the ramp family's endpoint mention
		} else if role, metric, ok := fontMetric(name); ok {
			if !strings.Contains(readme, metric) {
				t.Errorf("readme.md does not mention the %q metric suffix for %s", metric, name)
			}
			mention = role
		}
		if !strings.Contains(readme, mention) {
			t.Errorf("sheet variable %s is undocumented: readme.md lacks %q", name, mention)
		}
	}
}

// isPinName reports whether a sheet variable is a pinned/semantic colour
// rather than a ramp step.
func isPinName(name string) bool {
	for _, pin := range pinRoles {
		if name == "--color-"+pin.name {
			return true
		}
	}
	return false
}

// fontMetric splits a --font-<role>-<metric> variable into the role name
// and its metric suffix; --font-family is not a per-role metric.
func fontMetric(name string) (role, metric string, ok bool) {
	base, found := strings.CutPrefix(name, "--font-")
	if !found || name == "--font-family" {
		return "", "", false
	}
	for _, suf := range []string{"-size", "-line-height", "-weight", "-tracking"} {
		if r, has := strings.CutSuffix(base, suf); has {
			return r, suf, true
		}
	}
	return "", "", false
}
