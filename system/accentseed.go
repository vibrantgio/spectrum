package system

import (
	"image/color"
	"strconv"
	"strings"
)

// This file holds the pure halves of the Windows and Linux accent sources —
// decoding and parsing with no syscalls, no exec and no file I/O — so they
// compile and test on every platform. The platform shims
// (system_windows.go, system_linux.go) do only the I/O and feed these.

// nrgbaFromABGR decodes the Windows DWM AccentColor registry DWORD, whose
// byte layout is 0xAABBGGRR (little-endian ABGR). The alpha byte is a DWM
// composition detail, not part of the user's chosen colour, so the result
// is forced opaque — a seed colour is always fully opaque.
func nrgbaFromABGR(v uint32) color.NRGBA {
	return color.NRGBA{
		R: uint8(v),
		G: uint8(v >> 8),
		B: uint8(v >> 16),
		A: 0xFF,
	}
}

// gnomeAccentSeeds maps the GNOME 47+ accent-color names (the
// org.gnome.desktop.interface accent-color enum) to libadwaita's published
// accent background colours (AdwAccentColor accent_bg_color, libadwaita
// 1.6). These are the colours GNOME itself paints accented widgets with,
// so a seeded palette matches the desktop exactly.
var gnomeAccentSeeds = map[string]color.NRGBA{
	"blue":   {R: 0x35, G: 0x84, B: 0xE4, A: 0xFF},
	"teal":   {R: 0x21, G: 0x90, B: 0xA4, A: 0xFF},
	"green":  {R: 0x3A, G: 0x94, B: 0x4A, A: 0xFF},
	"yellow": {R: 0xC8, G: 0x88, B: 0x00, A: 0xFF},
	"orange": {R: 0xED, G: 0x5B, B: 0x00, A: 0xFF},
	"red":    {R: 0xE6, G: 0x2D, B: 0x42, A: 0xFF},
	"pink":   {R: 0xD5, G: 0x61, B: 0x99, A: 0xFF},
	"purple": {R: 0x91, G: 0x41, B: 0xAC, A: 0xFF},
	"slate":  {R: 0x6F, G: 0x83, B: 0x96, A: 0xFF},
}

// gnomeAccentSeed parses the output of
//
//	gsettings get org.gnome.desktop.interface accent-color
//
// — a GVariant string like `'blue'` plus trailing newline — and maps the
// name onto its published colour. ok is false for an unknown name, empty
// output, or the error text a pre-47 GNOME's "No such key" produces (which
// never matches a name).
func gnomeAccentSeed(out string) (seed color.NRGBA, ok bool) {
	name := strings.Trim(strings.TrimSpace(out), "'")
	seed, ok = gnomeAccentSeeds[name]
	return seed, ok
}

// kdeGlobalsAccent extracts the accent colour from the contents of KDE's
// kdeglobals file: the AccentColor=r,g,b key in the [General] section,
// written when the user picks an explicit accent in Plasma's settings.
// ok is false when the key is absent — a colour scheme with no accent
// override, KDE's default state — or malformed; the same key in any other
// section is ignored.
func kdeGlobalsAccent(content string) (seed color.NRGBA, ok bool) {
	inGeneral := false
	for line := range strings.Lines(content) {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[") {
			inGeneral = line == "[General]"
			continue
		}
		if !inGeneral {
			continue
		}
		key, val, found := strings.Cut(line, "=")
		if !found || strings.TrimSpace(key) != "AccentColor" {
			continue
		}
		parts := strings.Split(val, ",")
		if len(parts) != 3 {
			return color.NRGBA{}, false
		}
		var rgb [3]uint8
		for i, p := range parts {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || n < 0 || n > 255 {
				return color.NRGBA{}, false
			}
			rgb[i] = uint8(n)
		}
		return color.NRGBA{R: rgb[0], G: rgb[1], B: rgb[2], A: 0xFF}, true
	}
	return color.NRGBA{}, false
}

// desktopFromXDG classifies the XDG_CURRENT_DESKTOP value — a
// colon-separated list like "ubuntu:GNOME" or "KDE" — into the desktop
// families the Linux accent source knows how to read. It returns "gnome",
// "kde", or "" for everything else (Xfce, Cinnamon, an empty variable, …).
// Matching is case-insensitive per segment, and substring-based so
// variants like "GNOME-Classic" classify as GNOME.
func desktopFromXDG(xdgCurrentDesktop string) string {
	for _, part := range strings.Split(xdgCurrentDesktop, ":") {
		upper := strings.ToUpper(part)
		switch {
		case strings.Contains(upper, "GNOME"):
			return "gnome"
		case strings.Contains(upper, "KDE"):
			return "kde"
		}
	}
	return ""
}
