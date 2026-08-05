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
