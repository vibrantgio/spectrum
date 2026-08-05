package tokens

// RadiusScale holds border-radius stops in device-independent pixels,
// mirroring the Tailwind border-radius naming convention.
type RadiusScale struct {
	None float32 // 0 dp
	Sm   float32 // 2 dp
	Base float32 // 4 dp
	Md   float32 // 6 dp
	Lg   float32 // 8 dp
	Xl   float32 // 12 dp
	Xl2  float32 // 16 dp
	Xl3  float32 // 24 dp
	Full float32 // 9999 dp — pill / full-circle
}

// Radius is the default scale instance.
var Radius = RadiusScale{
	None: 0,
	Sm:   2,
	Base: 4,
	Md:   6,
	Lg:   8,
	Xl:   12,
	Xl2:  16,
	Xl3:  24,
	Full: 9999,
}
