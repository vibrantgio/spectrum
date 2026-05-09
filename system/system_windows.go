package system

// windowsSource returns the zero-value Appearance. A live implementation
// would read HKEY_CURRENT_USER\Software\Microsoft\Windows\CurrentVersion
// \Themes\Personalize\AppsUseLightTheme via golang.org/x/sys/windows/registry
// and watch it with RegNotifyChangeKeyValue; we ship a stub here so
// spectrum/system builds and links everywhere, and a later G2.2-windows
// milestone can implement the live source.
type windowsSource struct{}

func (windowsSource) Read() (Appearance, error) {
	return Appearance{}, nil
}

func defaultSource() Source { return windowsSource{} }
