// Package window pairs an [mvu.Window] with the rx.Observable[theme.Theme]
// that drives its rendering. Each [Window] holds an independent theme
// stream — emissions on one window's theme do not reach another window's
// stream — which is what makes per-window theme override possible.
//
// This is the runtime surface for the per-window theme override called
// out by Phase 2 of the design. Up to here the theme contract was
// expressed at the component level: every Prism component takes an
// rx.Observable[theme.Theme]. [Window] lifts that contract to the window
// level so that two windows in the same process can render in different
// themes simultaneously (light + dark, brand A + brand B, document A +
// document B...). The "runtime change" is small because rx already gives
// independent subscriptions independent state; the work is to define the
// contract and prove the isolation it implies.
package window

import (
	"gioui.org/layout"

	"github.com/reactivego/rx"

	"github.com/vibrantgio/mvu"
	"github.com/vibrantgio/prism/theme"
)

// Window pairs an [mvu.Window] with the theme observable that scopes
// its rendering. The Theme field is the only path by which the wrapped
// window's content learns of theme changes; constructing two Window
// values with two different Theme observables yields two fully isolated
// theme paths.
type Window struct {
	*mvu.Window
	Theme rx.Observable[theme.Theme]
}

// New wraps an [mvu.Window] with a theme observable. The returned Window
// holds theme by reference; later emissions on theme reach the build
// callback passed to [Window.Render] and to no other Window.
func New(w *mvu.Window, theme rx.Observable[theme.Theme]) *Window {
	return &Window{Window: w, Theme: theme}
}

// Render starts the wrapped [mvu.Window] event loop with layers built
// from this window's theme. The build callback receives this window's
// own Theme observable; sibling windows constructed with their own
// themes do not share state with it.
//
// Render shadows the embedded [mvu.Window.Render]; callers that want
// the raw layer-only signature can still reach it via w.Window.Render.
func (w *Window) Render(build func(theme rx.Observable[theme.Theme]) []rx.Observable[layout.Widget]) rx.Subscription {
	return w.Window.Render(build(w.Theme)...)
}
