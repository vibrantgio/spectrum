// Command vg-tokens writes the Claude Design project layout — theme.json,
// styles.css, readme.md and the foundation pages under foundations/ — for a
// spectrum theme into a target directory.
//
// Usage:
//
//	vg-tokens [-out dir] [-seed #rrggbb]
//
// With no flags it serialises theme.Default() into ./design. -seed rebrands
// the colour scheme from another seed; everything else stays the default.
package main

import (
	"flag"
	"fmt"
	stdcolor "image/color"
	"os"
	"strings"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/export"
	"github.com/vibrantgio/spectrum/theme"
	"github.com/vibrantgio/spectrum/tokens"
)

func main() {
	out := flag.String("out", "design", "target directory for the generated project")
	seedHex := flag.String("seed", "", "brand seed as #rrggbb (default: the built-in default seed)")
	flag.Parse()

	th := theme.Default()
	if *seedHex != "" {
		seed, err := parseSeed(*seedHex)
		if err != nil {
			fatal(err)
		}
		light, _ := tokens.FromSeed(seed)
		th.Color = rx.Of(light)
	}

	snap, err := export.Capture(th)
	if err != nil {
		fatal(err)
	}
	if err := export.Write(*out, snap); err != nil {
		fatal(err)
	}
	fmt.Printf("wrote %s: theme.json, styles.css, readme.md, foundations/{color,type,layout}.html\n", *out)
}

// parseSeed accepts #rrggbb or rrggbb.
func parseSeed(s string) (stdcolor.NRGBA, error) {
	hex, _ := strings.CutPrefix(s, "#")
	var r, g, b uint8
	if n, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b); n != 3 || err != nil || len(hex) != 6 {
		return stdcolor.NRGBA{}, fmt.Errorf("vg-tokens: -seed %q: want #rrggbb", s)
	}
	return stdcolor.NRGBA{R: r, G: g, B: b, A: 0xff}, nil
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
