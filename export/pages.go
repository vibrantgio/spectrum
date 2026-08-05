package export

import (
	"fmt"
	"html"
	stdcolor "image/color"
	"strings"

	"github.com/vibrantgio/spectrum/color"
	"github.com/vibrantgio/spectrum/tokens"
)

// The foundation pages are static HTML that reads only from the emitted
// token sheet: every colour, size, radius, shadow and font value in a style
// position is a var(--...) reference into ../styles.css, so regenerating the
// sheet from another seed reflows every page with no page edit. The only
// literal token values in a page are annotation text — hexes, px numbers and
// contrast measurements printed for the reader — which the generator
// computes from the Snapshot at generation time. Chrome CSS (flex, grid,
// margins for the specimen scaffolding) uses tokens where a token fits and
// token-free units (rem, thin) elsewhere; it never carries a literal colour
// or px length, and the page test enforces exactly that.

// chromeCSS is the scaffolding shared by every foundation page. No literal
// hex colours and no literal px lengths: colours, sizes, spacing and radii
// all resolve through the token sheet.
const chromeCSS = `body {
  margin: 0;
  background: var(--color-bg);
  color: var(--color-text);
  font-family: var(--font-family), system-ui, sans-serif;
  font-size: var(--font-body-medium-size);
  line-height: var(--font-body-medium-line-height);
  letter-spacing: var(--font-body-medium-tracking);
}
main {
  max-width: 64rem;
  margin: 0 auto;
  padding: var(--space-6) var(--space-6) var(--space-16);
}
.masthead {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-4);
}
h1 {
  font-size: var(--font-headline-large-size);
  line-height: var(--font-headline-large-line-height);
  font-weight: var(--font-headline-large-weight);
  letter-spacing: var(--font-headline-large-tracking);
  margin: var(--space-6) 0 var(--space-2);
}
h2 {
  font-size: var(--font-title-large-size);
  line-height: var(--font-title-large-line-height);
  font-weight: var(--font-title-large-weight);
  letter-spacing: var(--font-title-large-tracking);
  margin: var(--space-10) 0 var(--space-3);
}
h3 {
  font-size: var(--font-title-medium-size);
  line-height: var(--font-title-medium-line-height);
  font-weight: var(--font-title-medium-weight);
  letter-spacing: var(--font-title-medium-tracking);
  margin: var(--space-6) 0 var(--space-2);
}
.intro {
  max-width: 48rem;
  color: var(--color-neutral-700);
}
.annot {
  font-size: var(--font-label-small-size);
  line-height: var(--font-label-small-line-height);
  font-weight: var(--font-label-small-weight);
  letter-spacing: var(--font-label-small-tracking);
  color: var(--color-neutral-700);
  margin: var(--space-1) 0 0;
}
.mode-toggle {
  font-family: var(--font-family), system-ui, sans-serif;
  font-size: var(--font-label-large-size);
  line-height: var(--font-label-large-line-height);
  font-weight: var(--font-label-large-weight);
  letter-spacing: var(--font-label-large-tracking);
  padding: var(--space-2) var(--space-4);
  color: var(--color-text);
  background: var(--color-surface);
  border: thin solid var(--color-neutral-500);
  border-radius: var(--radius-base);
  cursor: pointer;
}
.mode-toggle:hover {
  background: var(--color-neutral-300);
}
`

// page wraps a body in the shared skeleton: the sheet link (foundations/ is
// a subdirectory, so the sheet is ../styles.css), the chrome, a masthead
// with the light/dark toggle, and the toggle script that flips the .dark
// class on the root element — the sheet's .dark override block restyles
// everything var()-driven from there.
func page(title, heading, intro, style, body string) string {
	var b strings.Builder
	b.WriteString("<!doctype html>\n<html lang=\"en\">\n<head>\n")
	b.WriteString("<meta charset=\"utf-8\">\n")
	b.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n")
	fmt.Fprintf(&b, "<title>%s</title>\n", html.EscapeString(title))
	b.WriteString("<link rel=\"stylesheet\" href=\"../styles.css\">\n")
	b.WriteString("<style>\n")
	b.WriteString(chromeCSS)
	b.WriteString(style)
	b.WriteString("</style>\n</head>\n<body>\n<main>\n")
	fmt.Fprintf(&b, "<div class=\"masthead\">\n<h1>%s</h1>\n", html.EscapeString(heading))
	b.WriteString("<button type=\"button\" class=\"mode-toggle\">Light / dark</button>\n</div>\n")
	fmt.Fprintf(&b, "<p class=\"intro\">%s</p>\n", intro)
	b.WriteString(body)
	b.WriteString("</main>\n<script>\n")
	b.WriteString("document.querySelector(\".mode-toggle\").addEventListener(\"click\", function () {\n")
	b.WriteString("  document.documentElement.classList.toggle(\"dark\");\n")
	b.WriteString("});\n")
	b.WriteString("</script>\n</body>\n</html>\n")
	return b.String()
}

// lcStr formats an APCA Lc measurement in the signed convention: positive
// dark-on-light, negative light-on-dark.
func lcStr(text, ground stdcolor.NRGBA) string {
	return fmt.Sprintf("%.1f", color.APCA(text, ground))
}

// wcagStr formats a WCAG 2 contrast ratio.
func wcagStr(a, b stdcolor.NRGBA) string {
	return fmt.Sprintf("%.2f:1", color.ContrastRatio(a, b))
}

// stepPurpose is ADR-007's job per ramp step: the step number carries the
// meaning, identically in both modes.
func stepPurpose(step int) string {
	switch step {
	case 100:
		return "tinted fill · ground"
	case 200:
		return "tinted fill · card"
	case 300:
		return "hover · subtle border"
	case 500:
		return "mid reference · strong border"
	case 700:
		return "low-contrast text · pressed"
	case 900:
		return "high-contrast text · pressed"
	default: // 400, 600, 800
		return "intermediate · state walk"
	}
}

// modeHex annotates one token's value in both modes, labelled, because text
// cannot flip with a CSS class the way a painted swatch does.
func modeHex(light, dark stdcolor.NRGBA) string {
	return fmt.Sprintf("L %s · D %s", hexRGB(light), hexRGB(dark))
}

// contrastRow renders one measured text pair as a table row: APCA Lc and the
// WCAG 2 ratio, in both modes. Per ADR-007 the Lc numbers gate the palette
// and the ratios are reported alongside.
func contrastRow(b *strings.Builder, label string, lightText, lightGround, darkText, darkGround stdcolor.NRGBA) {
	fmt.Fprintf(b, "<tr><th scope=\"row\">%s</th><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>\n",
		html.EscapeString(label),
		lcStr(lightText, lightGround), wcagStr(lightText, lightGround),
		lcStr(darkText, darkGround), wcagStr(darkText, darkGround))
}

// colorPageCSS is the colour page's specimen scaffolding.
const colorPageCSS = `.ramp {
  display: grid;
  grid-template-columns: repeat(9, 1fr);
  gap: var(--space-2);
  margin: var(--space-4) 0;
}
.chip {
  height: var(--space-16);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--radius-sm);
  border: thin solid var(--color-neutral-300);
  font-size: var(--font-label-medium-size);
  font-weight: var(--font-label-medium-weight);
}
.pins {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-4);
  margin: var(--space-4) 0;
}
.pin {
  min-width: var(--space-24);
}
.pin .chip {
  padding: 0 var(--space-4);
}
.contrast {
  border-collapse: collapse;
  margin: var(--space-4) 0;
}
.contrast th, .contrast td {
  border: thin solid var(--color-neutral-300);
  padding: var(--space-1) var(--space-3);
  font-size: var(--font-body-small-size);
  line-height: var(--font-body-small-line-height);
  text-align: right;
}
.contrast th[scope="row"] {
  text-align: left;
  font-weight: var(--font-label-medium-weight);
}
.contrast thead th {
  background: var(--color-surface);
  text-align: right;
}
`

// colorRole is one ramp role's page model: the CSS name, both mode ramps,
// and the pinned chips shown beside the ramp — for the accent roles the
// pinned base with its on-colour, for neutral the semantic layer.
type colorRole struct {
	name string
	pins []pinChip
	// pairLabel names the pinned text pair measured in the contrast table.
	pairLabel                    string
	lightPinText, lightPinGround stdcolor.NRGBA
	darkPinText, darkPinGround   stdcolor.NRGBA
}

// pinChip is one pinned base rendered as a labelled chip: painted by its
// var, captioned with both modes' hexes.
type pinChip struct {
	bgVar, fgVar string         // CSS variable names, without var()
	label        string         // caption
	lightBg      stdcolor.NRGBA // annotation values
	darkBg       stdcolor.NRGBA
}

// colorHTML renders foundations/color.html: five roles, each with its full
// nine-step ramp, its pinned base(s), ADR-007's step purposes, and the
// measured APCA Lc (WCAG 2 ratio alongside) of each text pair in both modes.
func colorHTML(s Snapshot) string {
	roles := []colorRole{
		{
			name: "neutral",
			pins: []pinChip{
				{"--color-bg", "--color-text", "bg (pinned)", s.Light.Background, s.Dark.Background},
				{"--color-surface", "--color-text", "surface (neutral-200)", s.Light.Surface, s.Dark.Surface},
				{"--color-divider", "--color-text", "divider (neutral-300)", s.Light.Divider, s.Dark.Divider},
				{"--color-text", "--color-bg", "text (pinned)", s.Light.Text, s.Dark.Text},
			},
			pairLabel:      "text on bg",
			lightPinText:   s.Light.Text,
			lightPinGround: s.Light.Background,
			darkPinText:    s.Dark.Text,
			darkPinGround:  s.Dark.Background,
		},
		{
			name: "primary",
			pins: []pinChip{
				{"--color-accent", "--color-on-accent", "accent (pinned primary)", s.Light.Primary, s.Dark.Primary},
			},
			pairLabel:      "on-accent on accent",
			lightPinText:   s.Light.OnPrimary,
			lightPinGround: s.Light.Primary,
			darkPinText:    s.Dark.OnPrimary,
			darkPinGround:  s.Dark.Primary,
		},
		{
			name: "secondary",
			pins: []pinChip{
				{"--color-secondary", "--color-on-secondary", "secondary (pinned)", s.Light.Secondary, s.Dark.Secondary},
			},
			pairLabel:      "on-secondary on secondary",
			lightPinText:   s.Light.OnSecondary,
			lightPinGround: s.Light.Secondary,
			darkPinText:    s.Dark.OnSecondary,
			darkPinGround:  s.Dark.Secondary,
		},
		{
			name: "tertiary",
			pins: []pinChip{
				{"--color-tertiary", "--color-on-tertiary", "tertiary (pinned)", s.Light.Tertiary, s.Dark.Tertiary},
			},
			pairLabel:      "on-tertiary on tertiary",
			lightPinText:   s.Light.OnTertiary,
			lightPinGround: s.Light.Tertiary,
			darkPinText:    s.Dark.OnTertiary,
			darkPinGround:  s.Dark.Tertiary,
		},
		{
			name: "error",
			pins: []pinChip{
				{"--color-error", "--color-on-error", "error (pinned)", s.Light.Error, s.Dark.Error},
			},
			pairLabel:      "on-error on error",
			lightPinText:   s.Light.OnError,
			lightPinGround: s.Light.Error,
			darkPinText:    s.Dark.OnError,
			darkPinGround:  s.Dark.Error,
		},
	}

	var b strings.Builder
	for _, role := range roles {
		lightRamp := rampNamed(s.Light.Ramps, role.name)
		darkRamp := rampNamed(s.Dark.Ramps, role.name)
		fmt.Fprintf(&b, "<section>\n<h2>%s</h2>\n", role.name)

		// The ramp: nine swatches painted by their vars, captioned with the
		// step purpose and both modes' values.
		b.WriteString("<div class=\"ramp\">\n")
		for step := 100; step <= 900; step += 100 {
			labelVar := fmt.Sprintf("--color-%s-900", role.name)
			if step >= 600 {
				labelVar = fmt.Sprintf("--color-%s-100", role.name)
			}
			fmt.Fprintf(&b, "<div class=\"step\">\n<div class=\"chip\" style=\"background: var(--color-%s-%d); color: var(%s)\">%d</div>\n",
				role.name, step, labelVar, step)
			fmt.Fprintf(&b, "<p class=\"annot\">%s<br>%s</p>\n</div>\n",
				stepPurpose(step), modeHex(lightRamp.Step(step), darkRamp.Step(step)))
		}
		b.WriteString("</div>\n")

		// The pinned base(s), painted by their vars.
		b.WriteString("<h3>Pins</h3>\n<div class=\"pins\">\n")
		for _, pin := range role.pins {
			fmt.Fprintf(&b, "<div class=\"pin\">\n<div class=\"chip\" style=\"background: var(%s); color: var(%s)\">Aa</div>\n",
				pin.bgVar, pin.fgVar)
			fmt.Fprintf(&b, "<p class=\"annot\">%s<br><code>var(%s)</code><br>%s</p>\n</div>\n",
				html.EscapeString(pin.label), pin.bgVar, modeHex(pin.lightBg, pin.darkBg))
		}
		b.WriteString("</div>\n")

		// The measured text pairs: ADR-007's gates, both modes, Lc first and
		// the WCAG ratio reported alongside.
		b.WriteString("<h3>Measured contrast</h3>\n<table class=\"contrast\">\n<thead>\n")
		b.WriteString("<tr><th scope=\"col\">text pair</th><th scope=\"col\">light Lc</th><th scope=\"col\">light WCAG</th><th scope=\"col\">dark Lc</th><th scope=\"col\">dark WCAG</th></tr>\n")
		b.WriteString("</thead>\n<tbody>\n")
		for _, pair := range [][2]int{{900, 100}, {900, 200}, {700, 100}, {700, 200}} {
			text, ground := pair[0], pair[1]
			contrastRow(&b, fmt.Sprintf("%d on %d", text, ground),
				lightRamp.Step(text), lightRamp.Step(ground),
				darkRamp.Step(text), darkRamp.Step(ground))
		}
		contrastRow(&b, role.pairLabel,
			role.lightPinText, role.lightPinGround,
			role.darkPinText, role.darkPinGround)
		b.WriteString("</tbody>\n</table>\n</section>\n")
	}

	intro := "Each role carries a nine-step functional ramp (100&ndash;900) where the step is the meaning &mdash; " +
		"100&ndash;300 tinted fills, hovers and subtle borders, 500 the mid-value reference, 700&ndash;900 text and pressed states (ADR-007) &mdash; " +
		"plus a pinned base. Dark mode is the paired ramp: the same step keeps the same job. " +
		"Swatches are painted through the token sheet, so the toggle restyles them; " +
		"annotation values are printed for both modes, labelled L and D. " +
		"APCA Lc is the gating metric (signed: negative means light-on-dark); WCAG 2 ratios are reported alongside."
	return page("Colour — Vibrant Gio foundations", "Colour", intro, colorPageCSS, b.String())
}

// rampNamed resolves a ramp by its CSS role name via the shared rampRoles
// table, so the pages and the sheet cannot disagree about which ramp a name
// means.
func rampNamed(set tokens.RampSet, name string) tokens.Ramp {
	for _, role := range rampRoles {
		if role.name == name {
			return role.ramp(set)
		}
	}
	panic("export: rampNamed: unknown role " + name)
}

// typePageCSS is the type page's specimen scaffolding.
const typePageCSS = `.type-role {
  margin: var(--space-8) 0;
  border-bottom: thin solid var(--color-divider);
  padding-bottom: var(--space-4);
}
.specimen {
  margin: 0;
}
`

// typeHTML renders foundations/type.html: all fifteen type roles at their
// real size, weight, line height and tracking, each styled entirely through
// its --font-<role>-* vars and annotated with the numbers.
func typeHTML(s Snapshot) string {
	var b strings.Builder
	for _, role := range typeRoles {
		style := role.pick(s.Typography)
		fmt.Fprintf(&b, "<section class=\"type-role\">\n")
		fmt.Fprintf(&b, "<p class=\"specimen\" style=\"font-size: var(--font-%[1]s-size); line-height: var(--font-%[1]s-line-height); font-weight: var(--font-%[1]s-weight); letter-spacing: var(--font-%[1]s-tracking)\">The five boxing wizards jump quickly</p>\n", role.name)
		fmt.Fprintf(&b, "<p class=\"annot\">%s &middot; %s / %s &middot; weight %d &middot; tracking %s</p>\n",
			role.name, px(style.Size), px(style.LineHeight), style.Weight, px(style.Tracking))
		b.WriteString("</section>\n")
	}
	intro := fmt.Sprintf("Every type role at its real size, weight, line height and tracking, styled through the "+
		"<code>--font-&lt;role&gt;-*</code> tokens. The face is the family the tokens name &mdash; %s &mdash; via "+
		"<code>var(--font-family)</code>; the browser must have it installed, otherwise the system fallback face renders at the same metrics.",
		html.EscapeString(s.Typography.BodyLarge.Typeface))
	return page("Type — Vibrant Gio foundations", "Type", intro, typePageCSS, b.String())
}

// layoutPageCSS is the layout page's specimen scaffolding.
const layoutPageCSS = `.space-row {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  margin: var(--space-2) 0;
}
.space-row .annot {
  flex: 0 0 10rem;
  margin: 0;
}
.space-bar {
  height: var(--space-4);
  background: var(--color-accent);
  border-radius: var(--radius-sm);
}
.specimen-grid {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-8);
  margin: var(--space-4) 0;
}
.radius-box {
  width: var(--space-24);
  height: var(--space-24);
  background: var(--color-surface);
  border: thin solid var(--color-neutral-500);
  display: flex;
  align-items: center;
  justify-content: center;
}
.shadow-card {
  width: var(--space-24);
  height: var(--space-20);
  background: var(--color-surface);
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
}
`

// layoutHTML renders foundations/layout.html: the spacing scale as sized
// bars, the radius scale on sample boxes, and today's shadow elevation on
// cards — each specimen sized, rounded or shadowed by its var.
func layoutHTML(s Snapshot) string {
	var b strings.Builder

	b.WriteString("<section>\n<h2>Spacing</h2>\n")
	for _, key := range spaceKeys {
		fmt.Fprintf(&b, "<div class=\"space-row\">\n<p class=\"annot\"><code>--space-%s</code> &middot; %s</p>\n", key.name, px(key.pick(s.Spacing)))
		fmt.Fprintf(&b, "<div class=\"space-bar\" style=\"width: var(--space-%s)\"></div>\n</div>\n", key.name)
	}
	b.WriteString("</section>\n")

	b.WriteString("<section>\n<h2>Radius</h2>\n<div class=\"specimen-grid\">\n")
	for _, key := range radiusKeys {
		fmt.Fprintf(&b, "<div>\n<div class=\"radius-box\" style=\"border-radius: var(--radius-%s)\">Aa</div>\n", key.name)
		fmt.Fprintf(&b, "<p class=\"annot\"><code>--radius-%s</code> &middot; %s</p>\n</div>\n", key.name, px(key.pick(s.Radius)))
	}
	b.WriteString("</div>\n</section>\n")

	b.WriteString("<section>\n<h2>Elevation</h2>\n")
	b.WriteString("<p class=\"intro\">Shadow depths as they stand today: level N casts a y-offset of its dp depth with twice that blur. E2.1 remaps elevation to surface roles and E5.1 re-renders this section.</p>\n")
	b.WriteString("<div class=\"specimen-grid\">\n")
	for _, level := range elevationLevels {
		fmt.Fprintf(&b, "<div>\n<div class=\"shadow-card\" style=\"box-shadow: var(--shadow-%s)\">%s</div>\n", level.name, level.name)
		fmt.Fprintf(&b, "<p class=\"annot\"><code>--shadow-%s</code> &middot; depth %sdp</p>\n</div>\n", level.name, fnum(level.pick(s.Elevation)))
	}
	b.WriteString("</div>\n</section>\n")

	intro := "The spacing scale as sized bars, the radius scale on sample boxes and the elevation steps on cards &mdash; " +
		"every bar width, corner radius and shadow resolves through its token, so the sheet is the single source of these shapes."
	return page("Layout — Vibrant Gio foundations", "Layout", intro, layoutPageCSS, b.String())
}
