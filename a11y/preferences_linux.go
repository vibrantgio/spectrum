package a11y

// linuxSource returns all-false; no reliable cross-desktop accessibility API
// is available without desktop-environment-specific dependencies.
type linuxSource struct{}

func (linuxSource) Read() (A11yPrefs, error) {
	return A11yPrefs{}, nil
}

func defaultSource() Source { return linuxSource{} }
