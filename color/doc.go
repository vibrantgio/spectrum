// Package color is the generative colour engine the palettes are derived
// with. It converts between sRGB and the perceptual spaces the derivation
// runs in — no colour values live here, only the mathematics.
//
// The tone axis is CIELAB L*, exactly as Material Design 3 defines tone:
// Lab and RGB convert an sRGB triple to L*a*b* and back through XYZ under
// standard illuminant D65. Hue and chroma come from OKLCh: OKLab, OKLCh
// and their inverses live in oklab.go, after Björn Ottosson's reference
// formulation; the two spaces share the sRGB linearisation but nothing
// else.
//
// The package name collides with image/color in an importer's import list;
// alias one of them — spectrum code aliases the standard library one.
package color
