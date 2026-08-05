package system

import (
	"image/color"
	"testing"
)

// TestNRGBAFromABGR pins the DWM AccentColor DWORD decode: the byte layout
// is 0xAABBGGRR, and the alpha byte is discarded in favour of opaque.
func TestNRGBAFromABGR(t *testing.T) {
	cases := []struct {
		name string
		raw  uint32
		want color.NRGBA
	}{
		// Windows' default blue accent #0078D7 as DWM stores it.
		{"default blue", 0xFFD77800, color.NRGBA{R: 0x00, G: 0x78, B: 0xD7, A: 0xFF}},
		// Distinct channel bytes prove the ordering (R lowest, B highest).
		{"channel order", 0x00332211, color.NRGBA{R: 0x11, G: 0x22, B: 0x33, A: 0xFF}},
		// The alpha byte, whatever DWM wrote, never leaks through.
		{"alpha discarded", 0xC4D77800, color.NRGBA{R: 0x00, G: 0x78, B: 0xD7, A: 0xFF}},
		{"zero", 0x00000000, color.NRGBA{A: 0xFF}},
		{"all ones", 0xFFFFFFFF, color.NRGBA{R: 0xFF, G: 0xFF, B: 0xFF, A: 0xFF}},
	}
	for _, tc := range cases {
		if got := nrgbaFromABGR(tc.raw); got != tc.want {
			t.Errorf("%s: nrgbaFromABGR(%#08x) = %+v, want %+v", tc.name, tc.raw, got, tc.want)
		}
	}
}

// TestGnomeAccentSeedNames pins the GNOME name → colour table against
// libadwaita's published accent_bg_color values, independently of the
// package's own table, so a silent edit fails here. Inputs carry the
// GVariant quoting and trailing newline gsettings actually emits.
func TestGnomeAccentSeedNames(t *testing.T) {
	cases := []struct {
		out  string
		want color.NRGBA
	}{
		{"'blue'\n", color.NRGBA{R: 0x35, G: 0x84, B: 0xE4, A: 0xFF}},
		{"'teal'\n", color.NRGBA{R: 0x21, G: 0x90, B: 0xA4, A: 0xFF}},
		{"'green'\n", color.NRGBA{R: 0x3A, G: 0x94, B: 0x4A, A: 0xFF}},
		{"'yellow'\n", color.NRGBA{R: 0xC8, G: 0x88, B: 0x00, A: 0xFF}},
		{"'orange'\n", color.NRGBA{R: 0xED, G: 0x5B, B: 0x00, A: 0xFF}},
		{"'red'\n", color.NRGBA{R: 0xE6, G: 0x2D, B: 0x42, A: 0xFF}},
		{"'pink'\n", color.NRGBA{R: 0xD5, G: 0x61, B: 0x99, A: 0xFF}},
		{"'purple'\n", color.NRGBA{R: 0x91, G: 0x41, B: 0xAC, A: 0xFF}},
		{"'slate'\n", color.NRGBA{R: 0x6F, G: 0x83, B: 0x96, A: 0xFF}},
	}
	for _, tc := range cases {
		got, ok := gnomeAccentSeed(tc.out)
		if !ok {
			t.Errorf("gnomeAccentSeed(%q): not ok", tc.out)
			continue
		}
		if got != tc.want {
			t.Errorf("gnomeAccentSeed(%q) = %+v, want %+v", tc.out, got, tc.want)
		}
	}
	// Unquoted input (defensive: some gsettings frontends strip quotes).
	if got, ok := gnomeAccentSeed("blue"); !ok || got != (color.NRGBA{R: 0x35, G: 0x84, B: 0xE4, A: 0xFF}) {
		t.Errorf("gnomeAccentSeed(\"blue\") = %+v, %v; want the blue seed, true", got, ok)
	}
}

// TestGnomeAccentSeedRejects verifies unknown names, empty output and a
// pre-47 GNOME's error text all fold to "no accent".
func TestGnomeAccentSeedRejects(t *testing.T) {
	for _, out := range []string{
		"",
		"\n",
		"'magenta'\n",
		"No such key “accent-color”\n",
	} {
		if _, ok := gnomeAccentSeed(out); ok {
			t.Errorf("gnomeAccentSeed(%q): ok = true, want false", out)
		}
	}
}

// TestKDEGlobalsAccent parses realistic kdeglobals fixtures.
func TestKDEGlobalsAccent(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    color.NRGBA
		wantOK  bool
	}{
		{
			name: "accent in General",
			content: "[ColorEffects:Disabled]\nColorAmount=0.55\n\n" +
				"[General]\nAccentColor=61,174,233\nColorScheme=BreezeLight\n\n" +
				"[KDE]\nLookAndFeelPackage=org.kde.breeze.desktop\n",
			want:   color.NRGBA{R: 61, G: 174, B: 233, A: 0xFF},
			wantOK: true,
		},
		{
			name:    "spaces around values",
			content: "[General]\nAccentColor = 255, 0, 128\n",
			want:    color.NRGBA{R: 255, G: 0, B: 128, A: 0xFF},
			wantOK:  true,
		},
		{
			name:    "no accent key",
			content: "[General]\nColorScheme=BreezeDark\n",
			wantOK:  false,
		},
		{
			name:    "accent outside General ignored",
			content: "[Colors:Button]\nAccentColor=1,2,3\n\n[General]\nColorScheme=Breeze\n",
			wantOK:  false,
		},
		{
			name:    "empty file",
			content: "",
			wantOK:  false,
		},
		{
			name:    "malformed: two components",
			content: "[General]\nAccentColor=61,174\n",
			wantOK:  false,
		},
		{
			name:    "malformed: not a number",
			content: "[General]\nAccentColor=red,green,blue\n",
			wantOK:  false,
		},
		{
			name:    "malformed: out of range",
			content: "[General]\nAccentColor=300,0,0\n",
			wantOK:  false,
		},
		{
			name:    "malformed: negative",
			content: "[General]\nAccentColor=-1,0,0\n",
			wantOK:  false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := kdeGlobalsAccent(tc.content)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if ok && got != tc.want {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}

// TestDesktopFromXDG pins the XDG_CURRENT_DESKTOP classification.
func TestDesktopFromXDG(t *testing.T) {
	cases := []struct {
		env  string
		want string
	}{
		{"GNOME", "gnome"},
		{"ubuntu:GNOME", "gnome"},
		{"GNOME-Classic:GNOME", "gnome"},
		{"gnome", "gnome"},
		{"KDE", "kde"},
		{"kde", "kde"},
		{"XFCE", ""},
		{"X-Cinnamon", ""},
		{"", ""},
	}
	for _, tc := range cases {
		if got := desktopFromXDG(tc.env); got != tc.want {
			t.Errorf("desktopFromXDG(%q) = %q, want %q", tc.env, got, tc.want)
		}
	}
}
