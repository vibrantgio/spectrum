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
appearance a second time. The same stream carries the OS accent colour — an
accent change re-emits the theme just like a dark-mode flip — and while the OS
reports increased contrast, the `Color` observable emits a high-contrast
variant derived from the resolved palette's own seed. The only light/dark
branches left in the seven are the two that pick a chroma syntax theme for a
markdown code block, and they branch on the luminance of the background token
rather than on the OS, because chroma's themes are the one visual thing the
token set does not cover.

The module is deliberately small and, below the `window` package, nearly
Gio-free: `system`, `preferences`, `a11y`, `export` and `color` talk to the OS,
the filesystem and the mathematics and import no UI toolkit, so the runtime is
testable without a display. The one exception is `tokens`, whose `Typography`
owns the system's single `*text.Shaper` and therefore imports Gio's text
machinery.

## Where it sits

Tier 1 of the stack — `mvu → spectrum → prism → pulse → cadence → markdown` —
and since the G-B3 inversion it really is the foundation: the module that owns
the design values everything above is styled from. spectrum imports
[mvu](https://github.com/vibrantgio/mvu) and
[font](https://github.com/vibrantgio/font) — Roboto and Roboto Mono are the
default `Typography`'s faces — and nothing above it, with one recorded
exception: the deprecated `spectrum/transition` alias shim imports
`pulse/transition`, the package's home since the inversion, and F3.3 of the
[org plan](https://github.com/vibrantgio/.github) (planned, not yet landed)
deletes the shim. Everything above imports spectrum — prism, pulse, cadence
and markdown all read `theme` and `tokens` from here, and the
[workbench](https://github.com/vibrantgio/workbench) applications bootstrap
`system` and `window`. The [organization page](https://github.com/vibrantgio)
has the full tier table.

```sh
go get github.com/vibrantgio/spectrum
```

Every module in the organization is on gioui.org v0.10.1,
github.com/reactivego/rx v0.3.0 and Go 1.25.1.

## Packages

| Package | |
| --- | --- |
| `tokens` | The typed design values, all of them: the ADR-007 colour ramps and pins, with `FromSeed` deriving both modes from one seed colour; `Typography` — fifteen MD3 text roles plus `Code`, carrying the faces and the one shared shaper; `Density` (Comfortable 36 dp / Compact 28 dp control heights); `MotionScale` (duration stops, easings, spring presets, and `Reduced()` for the OS reduce-motion preference); the elevation ladder (`SurfaceAt`, levels 0–3); and the 4-pt spacing and named radius scales. |
| `color` | The generative colour engine the palettes are derived with — sRGB ↔ CIELAB and OKLCh conversions and the APCA contrast metric that gates every generated pair. Mathematics only; no colour values live here. |
| `theme` | `Theme`: one `rx.Observable` per token category, so a consumer subscribes to just the categories it reads. `Default()` and `AutoLightDark()` construct one — note `AutoLightDark()` reads the clock (hours 7–17 light), not the OS; `system.LiveTheme` is the real tracker. |
| `system` | The OS appearance — dark mode and accent colour — polled behind a `Source` interface and published as an observable that emits only on change. `Live` gives the raw `Appearance`; `LiveTheme` gives the `theme.Theme` a window wants, with `WithSeed`/`WithPalette` options for branding. Dark mode is read on macOS; the accent is read on all three platforms — macOS's accent choice, the Windows DWM registry value, GNOME's named accent and KDE's `kdeglobals` RGB. |
| `a11y` | OS accessibility preferences — reduce motion, increased contrast, larger text — polled and published as an `rx.Observable[A11yPrefs]` that emits only on change. The composed theme already reflects the first two; macOS and Windows report real preferences, Linux returns all-false. |
| `window` | Pairs an `mvu.Window` with the theme observable that scopes it, and hands that observable to the layer builder. Two windows built with two theme streams render in two different themes in the same process. |
| `preferences` | Persists the user's explicit appearance choice — a theme name plus accessibility overrides — as JSON under the OS config directory, and reads it back at launch. |
| `export` | Serialises a `theme.Theme` emission into the project layout claude.ai/design consumes — `theme.json`, `styles.css`, `readme.md` and the foundation pages. `cmd/vg-tokens` is the command-line front door. |
| `transition` | **Deprecated alias** of [`pulse/transition`](https://github.com/vibrantgio/pulse), where the package moved in the G-B3 inversion. Import the pulse path; F3.3 of the org plan (planned) deletes this shim. |

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

Options on `LiveTheme` (and `FromSourceTheme`) brand the window without giving
up live OS tracking:

```go
// one brand colour; everything else derived, dark mode still live
specsystem.LiveTheme(time.Second, specsystem.WithSeed(brand))

// full control: both schemes supplied, the OS still picks which is live
specsystem.LiveTheme(time.Second, specsystem.WithPalette(light, dark))
```

Precedence, highest first: a palette option pins the pair — the application
chose its brand, the OS accent is ignored. With no palette option the stream
follows the OS accent colour live, each accent becoming the seed of a derived
pair; no accent at all falls back to the default palette. Accessibility
composes on top of whichever palette wins: while the OS reports increased
contrast, `Color` emits a high-contrast variant derived from the resolved
palette's own seed, and while it reports reduced motion, `Motion` emits
`MotionScale.Reduced()` — every duration zero.

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

- **Dark mode is only detected on macOS.** The Linux and Windows sources read
  the *accent* — GNOME's `gsettings` enum and KDE's `kdeglobals` RGB on Linux,
  the DWM registry value on Windows — but not dark mode: `Appearance.Dark`
  stays false there forever. Both files name the API a real implementation
  would use — an `org.freedesktop.appearance` portal read on Linux,
  `AppsUseLightTheme` plus a registry watch on Windows — and neither is
  written, nor claimed by any phase of the current plan.
- **The theme snaps; `transition` is not connected to anything.** The package
  interpolates token sets correctly and is golden-tested, but nothing drives
  it: `LiveTheme` emits the new palette in one step, and no module or
  application imports `pulse/transition` except this repository's own
  deprecated alias. A cross-fade today is the caller's to build out of
  `ColorTokensTween`.
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
- **The newest tag is behind the working tree.** v0.0.15, today's newest tag,
  predates the Roboto Mono faces and the `Typography.Code` style, so a build
  resolved from tags renders code in Roboto until the release in progress
  lands. That release plans v0.1.0 next, and the deprecation sweep after it is
  a breaking release planned at v0.2.0 — the major that deletes the
  `spectrum/transition` shim (with prism's `tokens`, `theme` and `a11y`
  aliases, F3.3) and re-cuts the frozen static `Render` surfaces downstream.
  Planned numbers, not cut tags.

## License

MIT — see [LICENSE](./LICENSE).
