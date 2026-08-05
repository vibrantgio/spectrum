package tokens

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
}

// DefaultTypography is the canonical MD3 typography: Roboto throughout, the
// same sizes as DefaultTypeScale, and the official MD3 line heights and
// tracking. Display, Headline, Title Large and Body roles are regular weight;
// Title Medium/Small and the Label roles are medium.
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
