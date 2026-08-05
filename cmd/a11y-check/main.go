// a11y-check prints OS accessibility preferences once per second.
// Run with: go run github.com/vibrantgio/spectrum/cmd/a11y-check
// Manual test: toggle "Reduce Motion" in System Settings > Accessibility > Display
// and confirm the change appears within ~1 second.
package main

import (
	"fmt"
	"time"

	"github.com/reactivego/rx"
	"github.com/vibrantgio/spectrum/a11y"
)

func main() {
	fmt.Println("Watching accessibility preferences (Ctrl-C to stop)...")
	fmt.Println("Toggle Settings > Accessibility > Display > Reduce Motion to verify propagation.")

	a11y.Live(time.Second).Subscribe(rx.GoroutineContext(), func(p a11y.A11yPrefs, err error, done bool) {
		if !done {
			fmt.Printf("[%s] ReduceMotion=%-5v  HighContrast=%-5v  IncreaseTextSize=%-5v\n",
				time.Now().Format("15:04:05"),
				p.ReduceMotion, p.HighContrast, p.IncreaseTextSize)
		}
	}).Wait()
}
