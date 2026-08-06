package tokens

import (
	"sync"

	"gioui.org/font"
	"gioui.org/text"
	"github.com/vibrantgio/font/roboto"
	"github.com/vibrantgio/font/robotomono"
)

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
	Typeface string
	Weight   int
	Size     float32

	// LineHeight is the height of one line box in dp — the whole box, not the
	// gap between lines, and not a multiplier. It means what CSS line-height
	// means: text in this role occupies LineHeight per line whatever its
	// glyphs measure, with the leading split evenly above and below the ink.
	// spectrum/export writes exactly this number into
	// `--font-<role>-line-height`, so the design-surface mirror and the Gio
	// rendering are stating the same fact.
	//
	// Handing it to gioui.org/widget.Label is not enough to get that, which is
	// the trap this comment exists for. Gio spends the line height on the gap
	// to the *next* line and gives the first line its own ascent plus descent,
	// so a MaxLines:1 label — nearly every control in this system — reports
	// the same size at any line height at all, and wrapped text lands one
	// deficit short of a whole multiple. Lay text out through
	// spectrum/typeset, which wraps widget.Label and adds the missing leading;
	// components in this organization all do.
	LineHeight float32

	Tracking float32
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

	// Faces is the font collection both shapers build from. Every Typeface a
	// role names must appear in it, or text in that role falls back to
	// whatever face the shaper picks instead. Resolution is by Typeface name,
	// so the order of entries only matters as fallback for text that names no
	// typeface at all — the first faces are the default family.
	//
	// Use WithFaces to add to it; assigning here after either shaper has been
	// built has no effect on that shaper.
	Faces []font.FontFace

	// shaper and pinnedShaper are the caches behind Shaper and
	// DeterministicShaper. They are separate fields because they hold
	// differently configured shapers, and handing back the wrong one would
	// make a test's pinned faces silently machine-dependent — or an
	// application's text silently unresolvable. Both are guarded by shaperMu.
	shaper       *text.Shaper
	pinnedShaper *text.Shaper
}

// shaperMu guards the lazily built shaper caches of every Typography value.
var shaperMu sync.Mutex

// Shaper returns the shaper applications should draw with: Faces first, then
// the platform's own fonts for anything Faces cannot serve. It is built on the
// first call, cached in the receiver, and safe for concurrent use from any
// number of goroutines.
//
// The fallback is the point. Faces is Roboto and Roboto Mono, which between
// them carry no arrow, no box-drawing character and no dingbat, so a shaper
// confined to them draws a missing-glyph box — tofu — for text a real
// application genuinely receives. Leaving the system fonts on means text
// resolves: all of it, including the glyphs no embedded face was ever going to
// have.
//
// What the platform serves therefore varies by machine, which is exactly why
// tests must not use this one. See DeterministicShaper.
//
// Copying a Typography value before its first shaper call is fine — the caches
// are per-copy, and each copy lazily builds its own. Copies made after the
// first call share the already built shaper, so change Faces through WithFaces
// rather than in place once shaping has started.
func (t *Typography) Shaper() *text.Shaper {
	shaperMu.Lock()
	defer shaperMu.Unlock()
	if t.shaper == nil {
		t.shaper = text.NewShaper(text.WithCollection(t.Faces))
	}
	return t.shaper
}

// DeterministicShaper returns the shaper golden tests must draw with: Faces
// and nothing else, system fonts off, so that the same text shapes to the same
// pixels on every machine. It is built on the first call, cached in the
// receiver separately from Shaper's, and safe for concurrent use.
//
// A test that pins its faces this way says what it wants, which is stricter
// than inheriting a default — it cannot drift when the default changes, and it
// cannot pass here and fail on a machine with a different font set.
//
// The pinning is real: a rune outside Faces shapes to the missing-glyph glyph
// rather than to whatever the machine happens to own. A test that legitimately
// draws such a rune adds the face that carries it instead of reaching for the
// platform's:
//
//	typ := tokens.DefaultTypography.WithFaces(notosansmono.FontFace())
//	shaper := typ.DeterministicShaper()
//
// That keeps the test deterministic without making it blind. It does not make
// symbols acceptable in a golden image: the face serving a symbol is the
// machine-dependent thing goldens exist to avoid, so symbol coverage is
// asserted at the glyph level — the shaper resolved this rune to a real face —
// and never as pixels.
func (t *Typography) DeterministicShaper() *text.Shaper {
	shaperMu.Lock()
	defer shaperMu.Unlock()
	if t.pinnedShaper == nil {
		t.pinnedShaper = text.NewShaper(text.NoSystemFonts(), text.WithCollection(t.Faces))
	}
	return t.pinnedShaper
}

// WithFaces returns a copy of t whose collection is Faces followed by extra,
// with both shaper caches cleared so the copy builds its own from the wider
// collection. The receiver is untouched, and neither slice is aliased.
//
// It is the one line an application adds a face with, rather than rebuilding
// the collection by hand:
//
//	typ := tokens.DefaultTypography.WithFaces(notosansmono.FontFace())
//
// Two callers want this. An application that cannot rely on system fonts — a
// container, a kiosk, anything shipping its own world — appends the optional
// symbol face, since the fallback Shaper counts on is not there to be had. And
// a test appends whatever face its subject legitimately draws from, then pins
// it with DeterministicShaper.
//
// The extra faces go last, so they serve as fallback without displacing the
// default family: text naming no typeface still resolves to Faces[0]'s.
//
// Call it while wiring, before shaping starts. Like every copy of a Typography
// it reads the shaper caches, so it is not safe to call concurrently with
// Shaper or DeterministicShaper on the same value.
func (t Typography) WithFaces(extra ...font.FontFace) Typography {
	faces := make([]font.FontFace, 0, len(t.Faces)+len(extra))
	faces = append(faces, t.Faces...)
	faces = append(faces, extra...)
	t.Faces = faces
	t.shaper = nil
	t.pinnedShaper = nil
	return t
}

// DefaultTypography is the canonical MD3 typography: Roboto throughout, the
// Material Design 3 sizes, and the official MD3 line heights and tracking. Display, Headline, Title Large and Body roles are regular weight;
// Title Medium/Small and the Label roles are medium. Code is BodyMedium's
// metrics on Roboto Mono, Roboto's companion mono face (G-F0); Faces carries
// the twelve Roboto faces first — the default family for text that names no
// typeface — then the four Roboto Mono faces Code resolves against.
//
// Sixteen faces, and no symbol face: font/notosansmono is deliberately absent,
// because Shaper's system fallback already covers what it carries and more.
// Add it with WithFaces where there is no system to fall back on.
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
