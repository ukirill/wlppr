package local

import (
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"github.com/ukirill/wlppr-go/provider"
)

type Provider struct {
	name  string
	path  string
	files []string
}

// New creates new local directory provider
func New(name, path string) provider.Provider {
	return &Provider{
		name,
		path,
		[]string{},
	}
}

func (f *Provider) Title() string {
	return f.name
}

func (f *Provider) Refresh() error {
	f.files = []string{}
	return filepath.Walk(f.path, walker(&f.files))
}

func (f *Provider) Random() (string, error) {
	rand.Seed(time.Now().Unix())
	return f.files[rand.Intn(len(f.files))], nil
}

func walker(files *[]string) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		ext := filepath.Ext(path)
		if ext != ".jpg" && ext != ".png" {
			return nil
		}
		*files = append(*files, path)
		return nil
	}
}
