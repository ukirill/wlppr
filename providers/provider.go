package providers

// Provider provides random wallpapers
type Provider interface {
	// Refresh source of wallpapers
	Refresh() error
	// Random return random wallpaper from source
	Random() (string, error)
}
