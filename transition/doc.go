// Package transition is a forwarding shim for
// github.com/vibrantgio/pulse/transition, which is where the colour-token
// tween bridge now lives: it is animation code depending on pulse/tween, and a
// foundation module should not depend on the effects layer (ADR-001). Every
// identifier here is a re-export of its pulse counterpart, so existing imports
// of this path keep compiling unchanged for one release cycle; the shim is
// removed in the next major release of spectrum.
//
// Deprecated: use github.com/vibrantgio/pulse/transition instead.
package transition
