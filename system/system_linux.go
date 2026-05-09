package system

// linuxSource returns the zero-value Appearance. Cross-desktop dark-mode
// and accent-colour detection requires a desktop-environment-specific
// dependency (D-Bus to org.freedesktop.appearance, gsettings, kreadconfig5,
// …); we ship a stub here so spectrum/system builds and links everywhere,
// and a later G2.2-linux milestone can implement the live source.
type linuxSource struct{}

func (linuxSource) Read() (Appearance, error) {
	return Appearance{}, nil
}

func defaultSource() Source { return linuxSource{} }
