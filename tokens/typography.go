package tokens

import (
	"sync"

	"gioui.org/font"
	"gioui.org/text"
	"github.com/vibrantgio/font/roboto"
	"github.com/vibrantgio/font/robotomono"
)

// TypeScale holds font-size stops for each Material Design 3 type role,
// expressed in device-independent pixels (dp).
type TypeScale struct {
	DisplayLarge  float32 // 57 dp
	DisplayMedium float32 // 45 dp
	DisplaySmall  float32 // 36 dp

	HeadlineLarge  float32 // 32 dp
	HeadlineMedium float32 // 28 dp
	HeadlineSmall  float32 // 24 dp

	TitleLarge  float32 // 22 dp
	TitleMedium float32 // 16 dp
	TitleSmall  float32 // 14 dp

	LabelLarge  float32 // 14 dp
	LabelMedium float32 // 12 dp
	LabelSmall  float32 // 11 dp

	BodyLarge  float32 // 16 dp
	BodyMedium float32 // 14 dp
	BodySmall  float32 // 12 dp
}

// CSS-style numeric font weights used by the default typography. The values
// follow the OpenType/CSS convention where regular is 400, so a zero weight
// always means "unset".
const (
	WeightRegular = 400
	WeightMedium  = 500
)

// FontWeight converts a CSS-style numeric weight, where regular is 400, to
// gioui.org's font.Weight, which counts in steps of 100 from regular at 0:
// FontWeight(400) is font.Normal and FontWeight(500) is font.Medium.
func FontWeight(weight int) font.Weight {
	return font.Weight(weight - WeightRegular)
}

// TextStyle describes one Material Design 3 type role: the typeface to shape
// with and its metrics. Size, LineHeight and Tracking (letter spacing) are in
// device-independent pixels (dp); Weight is a CSS-style numeric weight where
// regular is 400 and medium is 500.
type TextStyle struct {
	Typeface   string
	Weight     int
	Size       float32
	LineHeight float32
	Tracking   float32
}

// Typography holds one TextStyle per Material Design 3 type role.
type Typography struct {
	DisplayLarge  TextStyle
	DisplayMedium TextStyle
	DisplaySmall  TextStyle

	HeadlineLarge  TextStyle
	HeadlineMedium TextStyle
	HeadlineSmall  TextStyle

	TitleLarge  TextStyle
	TitleMedium TextStyle
	TitleSmall  TextStyle

	LabelLarge  TextStyle
	LabelMedium TextStyle
	LabelSmall  TextStyle

	BodyLarge  TextStyle
	BodyMedium TextStyle
	BodySmall  TextStyle

	// Code is the monospace style code renders in — markdown code blocks and
	// inline code spans. It is not one of the fifteen MD3 roles: MD3's 5×3
	// grid has no code role, so Code sits outside the grid as a sixteenth
	// style, carrying a body role's metrics on the mono face.
	Code TextStyle

	// Faces is the font collection Shaper builds from. Every Typeface a role
	// names must appear in it, or text in that role falls back to whatever
	// face the shaper picks instead. Resolution is by Typeface name, so the
	// order of entries only matters as fallback for text that names no
	// typeface at all — the first faces are the default family.
	Faces []font.FontFace

	// shaper is the cache behind Shaper. Guarded by shaperMu.
	shaper *text.Shaper
}

// shaperMu guards the lazily built shaper cache of every Typography value.
var shaperMu sync.Mutex

// Shaper returns the text shaper for Faces, building it on the first call and
// caching it in the receiver. The shaper is built with system fonts excluded,
// so text resolves only against Faces. It is safe for concurrent use from any
// number of goroutines.
//
// Copying a Typography value before its first Shaper call is fine — the cache
// is per-copy, and each copy lazily builds its own shaper. Copies made after
// the first call share the already built shaper, so change Faces only before
// shaping starts.
func (t *Typography) Shaper() *text.Shaper {
	shaperMu.Lock()
	defer shaperMu.Unlock()
	if t.shaper == nil {
		t.shaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(t.Faces))
	}
	return t.shaper
}

// DefaultTypography is the canonical MD3 typography: Roboto throughout, the
// same sizes as DefaultTypeScale, and the official MD3 line heights and
// tracking. Display, Headline, Title Large and Body roles are regular weight;
// Title Medium/Small and the Label roles are medium. Code is BodyMedium's
// metrics on Roboto Mono, Roboto's companion mono face (G-F0); Faces carries
// the twelve Roboto faces first — the default family for text that names no
// typeface — then the four Roboto Mono faces Code resolves against.
var DefaultTypography = Typography{
	DisplayLarge:  TextStyle{Typeface: "Roboto", Weight: WeightRegular, Size: 57, LineHeight: 64, Tracking: -0.25},
	DisplayMedium: TextStyle{Typeface: "Roboto", Weight: WeightRegular, Size: 45, LineHeight: 52, Tracking: 0},
	DisplaySmall:  TextStyle{Typeface: "Roboto", Weight: WeightRegular, Size: 36, LineHeight: 44, Tracking: 0},

	HeadlineLarge:  TextStyle{Typeface: "Roboto", Weight: WeightRegular, Size: 32, LineHeight: 40, Tracking: 0},
	HeadlineMedium: TextStyle{Typeface: "Roboto", Weight: WeightRegular, Size: 28, LineHeight: 36, Tracking: 0},
	HeadlineSmall:  TextStyle{Typeface: "Roboto", Weight: WeightRegular, Size: 24, LineHeight: 32, Tracking: 0},

	TitleLarge:  TextStyle{Typeface: "Roboto", Weight: WeightRegular, Size: 22, LineHeight: 28, Tracking: 0},
	TitleMedium: TextStyle{Typeface: "Roboto", Weight: WeightMedium, Size: 16, LineHeight: 24, Tracking: 0.15},
	TitleSmall:  TextStyle{Typeface: "Roboto", Weight: WeightMedium, Size: 14, LineHeight: 20, Tracking: 0.1},

	LabelLarge:  TextStyle{Typeface: "Roboto", Weight: WeightMedium, Size: 14, LineHeight: 20, Tracking: 0.1},
	LabelMedium: TextStyle{Typeface: "Roboto", Weight: WeightMedium, Size: 12, LineHeight: 16, Tracking: 0.5},
	LabelSmall:  TextStyle{Typeface: "Roboto", Weight: WeightMedium, Size: 11, LineHeight: 16, Tracking: 0.5},

	BodyLarge:  TextStyle{Typeface: "Roboto", Weight: WeightRegular, Size: 16, LineHeight: 24, Tracking: 0.5},
	BodyMedium: TextStyle{Typeface: "Roboto", Weight: WeightRegular, Size: 14, LineHeight: 20, Tracking: 0.25},
	BodySmall:  TextStyle{Typeface: "Roboto", Weight: WeightRegular, Size: 12, LineHeight: 16, Tracking: 0.4},

	Code: TextStyle{Typeface: "Roboto Mono", Weight: WeightRegular, Size: 14, LineHeight: 20, Tracking: 0.25},

	Faces: append(roboto.FontFaces(), robotomono.FontFaces()...),
}

// DefaultTypeScale is the canonical MD3 type scale.
var DefaultTypeScale = TypeScale{
	DisplayLarge:  57,
	DisplayMedium: 45,
	DisplaySmall:  36,

	HeadlineLarge:  32,
	HeadlineMedium: 28,
	HeadlineSmall:  24,

	TitleLarge:  22,
	TitleMedium: 16,
	TitleSmall:  14,

	LabelLarge:  14,
	LabelMedium: 12,
	LabelSmall:  11,

	BodyLarge:  16,
	BodyMedium: 14,
	BodySmall:  12,
}
