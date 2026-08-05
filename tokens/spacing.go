package tokens

// SpacingScale holds named stops on the 4-pt grid, Tailwind-aligned.
// Field names match the Tailwind spacing key; values are device-independent pixels.
type SpacingScale struct {
	S0  float32 // 0 dp
	S1  float32 // 4 dp
	S2  float32 // 8 dp
	S3  float32 // 12 dp
	S4  float32 // 16 dp
	S5  float32 // 20 dp
	S6  float32 // 24 dp
	S8  float32 // 32 dp
	S10 float32 // 40 dp
	S12 float32 // 48 dp
	S16 float32 // 64 dp
	S20 float32 // 80 dp
	S24 float32 // 96 dp
}

// Spacing is the default scale instance.
var Spacing = SpacingScale{
	S0:  0,
	S1:  4,
	S2:  8,
	S3:  12,
	S4:  16,
	S5:  20,
	S6:  24,
	S8:  32,
	S10: 40,
	S12: 48,
	S16: 64,
	S20: 80,
	S24: 96,
}
