package tokens

// ElevationScale holds shadow-depth stops in device-independent pixels,
// following Material Design 3 elevation levels 0–5.
type ElevationScale struct {
	Level0 float32 // 0 dp
	Level1 float32 // 1 dp
	Level2 float32 // 3 dp
	Level3 float32 // 6 dp
	Level4 float32 // 8 dp
	Level5 float32 // 12 dp
}

// Elevation is the default scale instance.
var Elevation = ElevationScale{
	Level0: 0,
	Level1: 1,
	Level2: 3,
	Level3: 6,
	Level4: 8,
	Level5: 12,
}

// ElevationLevel selects an entry on the [ElevationScale] by name.
// The dp value for a given level is read from the [Elevation] instance.
type ElevationLevel int

const (
	Level0 ElevationLevel = iota
	Level1
	Level2
	Level3
	Level4
	Level5
)
