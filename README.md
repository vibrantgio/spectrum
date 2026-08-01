# spectrum

The theme runtime of [Vibrant Gio](https://github.com/vibrantgio), a design
system for native desktop applications on macOS, Windows and Linux, written in
pure Go on [Gio](https://gioui.org). spectrum is the layer that answers one
question — *what does this window look like right now* — and answers it as a
stream, so the answer can change while the application runs.

Following the operating system between light and dark is the kind of thing that
is easy to demo and tedious to actually do: something has to poll the OS,
notice a real change rather than re-emitting the same value, turn it into
design tokens, and get those tokens to every widget on screen without the
application threading a parameter through its whole view tree. spectrum does
that in one line at startup. `system.LiveTheme` publishes the OS appearance as
an `rx.Observable[theme.Theme]`; `window.New` binds that observable to an
[mvu](https://github.com/vibrantgio/mvu) window and hands it to the builder
that constructs the layers. Every [prism](https://github.com/vibrantgio/prism)
component already takes a theme observable as its first argument, so the
appearance change reaches the buttons with no application code at all — which
is why all seven [workbench](https://github.com/vibrantgio/workbench)
applications bootstrap the same two lines and none of them asks the OS about
appearance a second time. The only light/dark branches left in the seven are
the two that pick a chroma syntax theme for a markdown code block, and they
branch on the luminance of the background token rather than on the OS, because
chroma's themes are the one visual thing the token set does not cover.

The module is deliberately small and deliberately Gio-free below the `window`
package: `system` and `preferences` talk to the OS and the filesystem and
import no UI toolkit, so the runtime is testable without a display.

## Where it sits

Tier 1 of the stack — `mvu → spectrum → prism → pulse → cadence → markdown`.
Only the [workbench](https://github.com/vibrantgio/workbench) applications
import spectrum; nothing inside the design system does. The
[organization page](https://github.com/vibrantgio) has the full tier table.

Its own imports are the part worth being honest about. spectrum imports
[mvu](https://github.com/vibrantgio/mvu), which is below it, and then
`theme`, `tokens` and `a11y` from [prism](https://github.com/vibrantgio/prism)
and `tween` from [pulse](https://github.com/vibrantgio/pulse), which are both
above it. The theme runtime therefore depends today on the component library it
exists to theme, and on the effects layer two tiers up. Phase B of the
[org plan](https://github.com/vibrantgio/.github) inverts that: `theme` and
`tokens` move down into this module and `transition` moves up into pulse,
leaving deprecated aliases behind so no downstream repository has to change a
line. After that move spectrum is the foundation — the module that owns the
design values everything above is styled from — rather than a consumer of them,
and it is the natural home for the generative colour engine Phase D builds.

```sh
go get github.com/vibrantgio/spectrum
```

Every module in the organization is on gioui.org v0.10.1,
github.com/reactivego/rx v0.3.0 and Go 1.25.1.

## Packages

| Package | |
| --- | --- |
| `system` | The OS appearance — dark mode and accent — polled behind a `Source` interface and published as an observable that emits only on change. `Live` gives the raw `Appearance`; `LiveTheme` gives the `theme.Theme` a window wants. Real on macOS; a stub on Linux and Windows. |
| `window` | Pairs an `mvu.Window` with the theme observable that scopes it, and hands that observable to the layer builder. Two windows built with two theme streams render in two different themes in the same process. |
| `preferences` | Persists the user's explicit appearance choice — a theme name plus accessibility overrides — as JSON under the OS config directory, and reads it back at launch. |
| `transition` | Interpolates a whole `tokens.ColorTokens` set between two values, so a light-to-dark flip can cross-fade rather than snap. Moves to `pulse/transition` in Phase B. |

## Usage

The whole bootstrap, from `main.go` in
[workbench/todos](https://github.com/vibrantgio/workbench/tree/master/todos) —
the smallest complete Vibrant Gio application. Two of these lines are spectrum:

```go
mvuWin := mvu.NewWindow(
	app.Title("Todos"),
	app.Size(unit.Dp(650), unit.Dp(600)),
)
w := specwin.New(mvuWin, specsystem.LiveTheme(time.Second))

models, runner := mvu.Loop(mvuWin.Messages(), Init, Update)
defer func() { runner.Unsubscribe(); runner.Wait() }()
modelObs := models.Publish().AutoConnect(modelObsConsumers)

if err := w.Render(buildLayers(modelObs)).Wait(); err != nil {
	fmt.Fprintln(os.Stderr, "todos:", err)
	os.Exit(1)
}
```

One second is the intended poll interval — the OS caches these values and will
not report a toggle much sooner.

`Render` is where the theme becomes the application's. It calls the build
function with this window's own theme observable and renders the layers that
come back, so the observable is a parameter rather than a global — this is
`view.go` from the same app:

```go
func buildLayers(modelObs rx.Observable[Model]) func(th rx.Observable[theme.Theme]) []rx.Observable[layout.Widget] {
	return func(th rx.Observable[theme.Theme]) []rx.Observable[layout.Widget] {
		return []rx.Observable[layout.Widget]{
			BackdropLayer(th),
			ContentLayer(th, modelObs),
		}
	}
}
```

From there `th` goes straight into prism and cadence components, which take it
as their first argument. A layer that needs the resolved values rather than the
`Theme` subscribes to the category it reads — each `LiveTheme` emission is a
static snapshot, every field an `rx.Of`, so the inner observable resolves
synchronously:

```go
themes := rx.SwitchMap(th, func(t theme.Theme) rx.Observable[themed] {
	return rx.Map(t.Color, func(c tokens.ColorTokens) themed {
		return themed{prism: t, palette: PaletteFrom(c)}
	})
})
```

To test any of this without an OS, implement `system.Source` and use
`FromSource` or `FromSourceTheme`; that is the whole test seam, and it is what
this module's own tests drive.

## For coding assistants

Read the canonical guide before writing code against this module — the module
inventory with current tags, the application skeleton, MVU and rx semantics,
typography, and the pitfalls that are not guessable:

<https://raw.githubusercontent.com/vibrantgio/.github/master/llms.txt>

[`AGENTS.md`](./AGENTS.md) in this repository has the build, test and
golden-image commands. The golden line there is exact and both halves of it
matter — `-golden.update` must follow the package list, and the list cannot be
replaced by `./...`.

## Status

Honest about what does not work yet:

- **You cannot supply your own colours.** `LiveTheme` and `FromSourceTheme`
  hardcode `tokens.DefaultLight` and `tokens.DefaultDark`; there is no seed, no
  palette parameter and no injection point anywhere in the stack. Branding an
  application today means building the `theme.Theme` yourself and giving up OS
  dark-mode tracking to do it — the two are mutually exclusive. Phase D is the
  generative colour engine, and D3.1 is the injection point that makes a custom
  palette and live light/dark switching coexist.
- **Dark mode is only detected on macOS.** `system_linux.go` and
  `system_windows.go` return the zero `Appearance` forever, so `Live` there
  emits light mode once and never again. Both files name the API a real
  implementation would use — an `org.freedesktop.appearance` D-Bus/gsettings
  read on Linux, `AppsUseLightTheme` plus `RegNotifyChangeKeyValue` on Windows —
  and neither is written. Phase D schedules the Windows and Linux *accent*
  sources; the dark-mode readers those two platforms need are not claimed by
  any phase of the current plan.
- **The accent colour is read and thrown away.** macOS `AppleAccentColor`
  (−1..7) is polled, throttled to one exec per ten seconds, and carried on
  `Appearance.AccentIndex` — and no consumer maps it to a colour, so the tinting
  it exists for does not happen. Phase D wires it.
- **The theme snaps; `transition` is not connected to anything.** The package
  interpolates token sets correctly and is golden-tested at frames 0/15/30, but
  nothing drives it: `LiveTheme` emits the new palette in one step, and no
  module or application in the organization imports `spectrum/transition`. A
  cross-fade today is the caller's to build out of `ColorTokensTween`. The
  package also moves to `pulse/transition` in Phase B, behind an alias.
- **`preferences` persists a choice nothing reads.** No module or application
  imports it, and there is no mapping from the stored theme name to a
  `theme.Theme` — the string round-trips to disk and stops there, as do the
  stored a11y overrides. `Observe` is also read-once-and-complete, so it does
  not notify anyone of a later `Save`.
- **The theme observable is cold, and sharing it costs OS calls.** Every
  subscription starts its own poller, so handing one `LiveTheme` stream to *n*
  consumers polls *n* times per interval — on macOS, *n* `defaults` fork+execs
  per second. Every workbench application does exactly this today with two
  layers. `rx` Publish/AutoConnect fixes it at the call site; nothing in the
  current plan changes the default.
- **The layering is inverted.** spectrum imports `prism/theme`, `prism/tokens`,
  `prism/a11y` and `pulse/tween`, so tier 1 reaches into tiers 2 and 3. Phase B
  moves `theme` and `tokens` down here and `transition` up into pulse, with
  aliases left behind.

## License

MIT — see [LICENSE](./LICENSE).
