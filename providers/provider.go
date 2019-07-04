package providers

// WlpprProvider provides random wallpapers
type WlpprProvider interface {
	// Refresh source of wallpapers
	Refresh() error
	// Random return random wallpaper from source
	Random() (string, error)
}
