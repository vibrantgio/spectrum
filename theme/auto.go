package theme

import (
	"time"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/tokens"
)

// AutoLightDark returns an Observable that emits a new Theme every minute.
// Hours 7–17 (inclusive) use DefaultLight; all other hours use DefaultDark.
func AutoLightDark() rx.Observable[Theme] {
	return rx.Map(rx.Ticker(0, time.Minute), func(t time.Time) Theme {
		colors := tokens.DefaultLight
		if t.Hour() <= 6 || t.Hour() >= 18 {
			colors = tokens.DefaultDark
		}
		return Theme{
			Color:     rx.Of(colors),
			Type:      rx.Of(tokens.DefaultTypeScale),
			Motion:    rx.Of(tokens.Motion),
			Spacing:   rx.Of(tokens.Spacing),
			Radius:    rx.Of(tokens.Radius),
			Elevation: rx.Of(tokens.Elevation),
		}
	})
}
