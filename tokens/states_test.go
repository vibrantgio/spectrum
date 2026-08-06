package tokens_test

import (
	"image/color"
	"math"
	"testing"

	speccolor "github.com/vibrantgio/spectrum/color"
	"github.com/vibrantgio/spectrum/tokens"
)

// stateSchemes are the two default schemes with the direction their ramps
// run: light luminance falls toward 900, dark rises.
var stateSchemes = []struct {
	name       string
	tok        tokens.ColorTokens
	descending bool
}{
	{"DefaultLight", tokens.DefaultLight, true},
	{"DefaultDark", tokens.DefaultDark, false},
}

var allRoles = []struct {
	name string
	role tokens.Role
}{
	{"Neutral", tokens.RoleNeutral},
	{"Primary", tokens.RolePrimary},
	{"Secondary", tokens.RoleSecondary},
	{"Tertiary", tokens.RoleTertiary},
	{"Error", tokens.RoleError},
	{"Success", tokens.RoleSuccess},
	{"Warning", tokens.RoleWarning},
}

// accentRoles are the roles with a pinned solid fill.
var accentRoles = allRoles[1:]

func rampForRole(t tokens.ColorTokens, role tokens.Role) tokens.Ramp {
	switch role {
	case tokens.RoleNeutral:
		return t.Ramps.Neutral
	case tokens.RolePrimary:
		return t.Ramps.Primary
	case tokens.RoleSecondary:
		return t.Ramps.Secondary
	case tokens.RoleTertiary:
		return t.Ramps.Tertiary
	case tokens.RoleSuccess:
		return t.Ramps.Success
	case tokens.RoleWarning:
		return t.Ramps.Warning
	default:
		return t.Ramps.Error
	}
}

func pinForRole(t tokens.ColorTokens, role tokens.Role) color.NRGBA {
	switch role {
	case tokens.RolePrimary:
		return t.Primary
	case tokens.RoleSecondary:
		return t.Secondary
	case tokens.RoleTertiary:
		return t.Tertiary
	case tokens.RoleSuccess:
		return t.Success
	case tokens.RoleWarning:
		return t.Warning
	default:
		return t.Error
	}
}

// TestTintedStatesStayOnRamp verifies the tinted regime resolves every
// state to an exact ramp step — never a blend or overlay — at the ADR-007
// walk: normal and focus on the ground, hover one step past, pressed,
// selected and dragged two, in both modes with the same index (the
// paired-scale invariant: same step, same job).
func TestTintedStatesStayOnRamp(t *testing.T) {
	for _, s := range stateSchemes {
		for _, r := range allRoles {
			ramp := rampForRole(s.tok, r.role)
			for ground := 100; ground <= 700; ground += 100 {
				checks := []struct {
					state tokens.State
					want  color.NRGBA
				}{
					{tokens.StateNormal, ramp.Step(ground)},
					{tokens.StateFocus, ramp.Step(ground)},
					{tokens.StateHover, ramp.Step(ground + 100)},
					{tokens.StatePressed, ramp.Step(ground + 200)},
					{tokens.StateSelected, ramp.Step(ground + 200)},
					{tokens.StateDragged, ramp.Step(ground + 200)},
				}
				for _, c := range checks {
					if got := s.tok.StateColor(r.role, ground, c.state); got != c.want {
						t.Errorf("%s %s: StateColor(ground %d, state %d) = %v, want ramp step %v",
							s.name, r.name, ground, c.state, got, c.want)
					}
				}
			}
		}
	}
}

// TestTintedWalkClampsAtRampEnd verifies walks past the ramp end clamp to
// step 900: ground 800 pressed resolves to 900, ground 900 stays put.
func TestTintedWalkClampsAtRampEnd(t *testing.T) {
	for _, s := range stateSchemes {
		n := s.tok.Ramps.Neutral
		checks := []struct {
			ground int
			state  tokens.State
			want   color.NRGBA
		}{
			{800, tokens.StateHover, n.Step(900)},
			{800, tokens.StatePressed, n.Step(900)},
			{900, tokens.StateHover, n.Step(900)},
			{900, tokens.StatePressed, n.Step(900)},
		}
		for _, c := range checks {
			if got := s.tok.StateColor(tokens.RoleNeutral, c.ground, c.state); got != c.want {
				t.Errorf("%s: StateColor(Neutral, %d, state %d) = %v, want clamped step 900 %v",
					s.name, c.ground, c.state, got, c.want)
			}
		}
	}
}

// TestTintedWalkMonotonic verifies the walk is monotonic along the ramp in
// the mode's own direction: on a light ground hover is darker than normal
// and pressed darker than hover; on a dark ground the same index walk is
// lighter — the paired scales make one rule serve both modes.
func TestTintedWalkMonotonic(t *testing.T) {
	for _, s := range stateSchemes {
		for _, r := range allRoles {
			for ground := 100; ground <= 700; ground += 100 {
				normal := relativeLuminance(s.tok.StateColor(r.role, ground, tokens.StateNormal))
				hover := relativeLuminance(s.tok.StateColor(r.role, ground, tokens.StateHover))
				pressed := relativeLuminance(s.tok.StateColor(r.role, ground, tokens.StatePressed))
				ok := normal > hover && hover > pressed
				if !s.descending {
					ok = normal < hover && hover < pressed
				}
				if !ok {
					t.Errorf("%s %s ground %d: luminance walk not monotonic: normal %.4f, hover %.4f, pressed %.4f",
						s.name, r.name, ground, normal, hover, pressed)
				}
			}
		}
	}
}

// TestSolidWalkMonotonicTowardNine verifies the solid-fill regime: from
// the pin, hover and pressed move strictly toward the 900 end of the
// paired scale — darker in light mode, lighter in dark mode — staying
// fully opaque (the tonal solver gamut-maps, so every result is sRGB).
func TestSolidWalkMonotonicTowardNine(t *testing.T) {
	for _, s := range stateSchemes {
		for _, r := range accentRoles {
			pin := s.tok.SolidStateColor(r.role, tokens.StateNormal)
			hover := s.tok.SolidStateColor(r.role, tokens.StateHover)
			pressed := s.tok.SolidStateColor(r.role, tokens.StatePressed)
			if pin != pinForRole(s.tok, r.role) {
				t.Errorf("%s %s: solid normal = %v, want the pin %v", s.name, r.name, pin, pinForRole(s.tok, r.role))
			}
			lPin, _, _ := speccolor.LabFromNRGBA(pin)
			lHover, _, _ := speccolor.LabFromNRGBA(hover)
			lPressed, _, _ := speccolor.LabFromNRGBA(pressed)
			ok := lPin > lHover && lHover > lPressed
			if !s.descending {
				ok = lPin < lHover && lHover < lPressed
			}
			if !ok {
				t.Errorf("%s %s: solid walk L* not monotonic toward 900: pin %.2f, hover %.2f, pressed %.2f",
					s.name, r.name, lPin, lHover, lPressed)
			}
			for state, c := range map[string]color.NRGBA{"hover": hover, "pressed": pressed} {
				if c.A != 0xff {
					t.Errorf("%s %s: solid %s = %v, want an opaque colour", s.name, r.name, state, c)
				}
			}
		}
	}
}

// TestDarkSolidWalkLandsOnPairedDepths verifies "same step, same job" for
// the solid regime: FromSeed pins every dark base at the dark scale's
// step-700 depth (D2.4 — the spike's step-500 depth failed the APCA
// on-colour gate), so its hover must land at step-800 depth and its
// pressed must clamp at step-900 depth — the exact rungs a tinted walk
// from 700 would visit.
func TestDarkSolidWalkLandsOnPairedDepths(t *testing.T) {
	const tol = 1.5 // L*, absorbs 8-bit quantization of pin and ramp
	for _, r := range accentRoles {
		ramp := rampForRole(tokens.DefaultDark, r.role)
		hover := tokens.DefaultDark.SolidStateColor(r.role, tokens.StateHover)
		pressed := tokens.DefaultDark.SolidStateColor(r.role, tokens.StatePressed)
		for _, c := range []struct {
			name string
			got  color.NRGBA
			rung int
		}{
			{"hover", hover, 800},
			{"pressed", pressed, 900},
		} {
			lGot, _, _ := speccolor.LabFromNRGBA(c.got)
			lWant, _, _ := speccolor.LabFromNRGBA(ramp.Step(c.rung))
			if math.Abs(lGot-lWant) > tol {
				t.Errorf("DefaultDark %s: solid %s L* = %.2f, want step-%d depth %.2f ± %.1f",
					r.name, c.name, lGot, c.rung, lWant, tol)
			}
		}
	}
}

// TestDisabledIsOpacity verifies disabled never walks the ramp: it is the
// normal colour at DisabledOpacity — MD3's 38%, alpha 0x61 for an opaque
// input, the value prism already renders.
func TestDisabledIsOpacity(t *testing.T) {
	if tokens.DisabledOpacity != 0.38 {
		t.Errorf("DisabledOpacity = %v, want 0.38", tokens.DisabledOpacity)
	}
	if got := tokens.Disabled(tokens.White); got != (color.NRGBA{0xff, 0xff, 0xff, 0x61}) {
		t.Errorf("Disabled(White) = %v, want alpha 0x61 with RGB untouched", got)
	}
	for _, s := range stateSchemes {
		ground := s.tok.Ramps.Neutral.Step(200)
		want := color.NRGBA{ground.R, ground.G, ground.B, 0x61}
		if got := s.tok.StateColor(tokens.RoleNeutral, 200, tokens.StateDisabled); got != want {
			t.Errorf("%s: disabled tinted = %v, want the ground at 38%% alpha %v", s.name, got, want)
		}
		pin := s.tok.Primary
		want = color.NRGBA{pin.R, pin.G, pin.B, 0x61}
		if got := s.tok.SolidStateColor(tokens.RolePrimary, tokens.StateDisabled); got != want {
			t.Errorf("%s: disabled solid = %v, want the pin at 38%% alpha %v", s.name, got, want)
		}
	}
}

// TestFocusIsTheRing verifies focus keeps the surface colour unchanged and
// the ring colour is Neutral 500 — ADR-007's "strong border, focusable
// edge", which is what the Outline alias resolved to before v0.2.0 deleted
// it. FocusRing() is the name for it now.
func TestFocusIsTheRing(t *testing.T) {
	for _, s := range stateSchemes {
		if got, want := s.tok.FocusRing(), s.tok.Ramps.Neutral.Step(500); got != want {
			t.Errorf("%s: FocusRing() = %v, want Neutral step 500 %v", s.name, got, want)
		}
		if got, want := s.tok.StateColor(tokens.RoleNeutral, 200, tokens.StateFocus), s.tok.Ramps.Neutral.Step(200); got != want {
			t.Errorf("%s: focused tinted = %v, want the unchanged ground %v", s.name, got, want)
		}
		if got, want := s.tok.SolidStateColor(tokens.RolePrimary, tokens.StateFocus), s.tok.Primary; got != want {
			t.Errorf("%s: focused solid = %v, want the unchanged pin %v", s.name, got, want)
		}
	}
}

// TestDraggedFollowsPressed verifies dragged resolves identically to
// pressed in both regimes.
func TestDraggedFollowsPressed(t *testing.T) {
	for _, s := range stateSchemes {
		for _, r := range allRoles {
			for ground := 100; ground <= 900; ground += 100 {
				dragged := s.tok.StateColor(r.role, ground, tokens.StateDragged)
				pressed := s.tok.StateColor(r.role, ground, tokens.StatePressed)
				if dragged != pressed {
					t.Errorf("%s %s ground %d: dragged %v != pressed %v", s.name, r.name, ground, dragged, pressed)
				}
			}
		}
		for _, r := range accentRoles {
			dragged := s.tok.SolidStateColor(r.role, tokens.StateDragged)
			pressed := s.tok.SolidStateColor(r.role, tokens.StatePressed)
			if dragged != pressed {
				t.Errorf("%s %s: solid dragged %v != pressed %v", s.name, r.name, dragged, pressed)
			}
		}
	}
}

// TestStateResolverPanics verifies out-of-vocabulary inputs panic, exactly
// as Ramp.Step does: bad grounds, unknown states and roles, and asking
// Neutral — which has no pinned base — for a solid fill.
func TestStateResolverPanics(t *testing.T) {
	mustPanic := func(name string, f func()) {
		defer func() {
			if recover() == nil {
				t.Errorf("%s: expected panic", name)
			}
		}()
		f()
	}
	tok := tokens.DefaultLight
	mustPanic("ground 150", func() { tok.StateColor(tokens.RoleNeutral, 150, tokens.StateHover) })
	mustPanic("ground 0", func() { tok.StateColor(tokens.RoleNeutral, 0, tokens.StateNormal) })
	mustPanic("unknown state", func() { tok.StateColor(tokens.RoleNeutral, 200, tokens.State(99)) })
	mustPanic("unknown role", func() { tok.StateColor(tokens.Role(99), 200, tokens.StateHover) })
	mustPanic("solid unknown state", func() { tok.SolidStateColor(tokens.RolePrimary, tokens.State(99)) })
	mustPanic("solid neutral", func() { tok.SolidStateColor(tokens.RoleNeutral, tokens.StateHover) })
	mustPanic("solid unknown role", func() { tok.SolidStateColor(tokens.Role(99), tokens.StateHover) })
}
