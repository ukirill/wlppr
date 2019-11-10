package providers

// Provider provides random wallpapers
type Provider interface {
	Title() string
	// Refresh source of wallpapers
	Refresh() error
	// Random returns url of random wallpaper from source
	Random() (string, error)
}
