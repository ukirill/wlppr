package switcher

import (
	"fmt"
	"github.com/ukirill/wlppr-go/internal"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sync/errgroup"

	"github.com/ukirill/wlppr-go/providers"
)

// https://msdn.microsoft.com/en-us/library/windows/desktop/ms724947.aspx
const (
	//spiGetDeskWallpaper = 0x0073
	spiSetDeskWallpaper = 0x0014

	uiParam = 0x0000

	spifUpdateINIFile = 0x01
	spifSendChange    = 0x02
)

// https://msdn.microsoft.com/en-us/library/windows/desktop/ms724947.aspx
var (
	user32               = syscall.NewLazyDLL("user32.dll")
	systemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

type Switcher interface {
	Providers() []providers.Provider
	SetProviders(p ...providers.Provider)
	AddProvider(p providers.Provider)
	RemoveProvider(p providers.Provider)
	SetDispNum(n int)
	Switch() error
	Refresh() error
	SaveCur(path string) error
}

// baseSwitcher uses providers to get random wallpaper and place it on desktop
type baseSwitcher struct {
	provs     []providers.Provider
	pMux      *sync.Mutex
	resH      int
	resW      int
	current   string
	dispnum   int
	cachePath string
}

func New(cachePath string, p ...providers.Provider) *baseSwitcher {
	return &baseSwitcher{
		provs:     p,
		pMux:      &sync.Mutex{},
		resH:      1080,
		resW:      1920,
		dispnum:   1,
		cachePath: cachePath,
	}
}

// Set providers as source for wallpaper
func (s *baseSwitcher) SetProviders(p ...providers.Provider) {
	s.provs = p
}

func (s *baseSwitcher) Providers() []providers.Provider {
	return s.provs
}

// AddProvider adds provider p if it isnt added already
func (s *baseSwitcher) AddProvider(p providers.Provider) {
	s.pMux.Lock()
	defer s.pMux.Unlock()
	for _, cur := range s.provs {
		if cur == p {
			return
		}
	}
	s.provs = append(s.provs, p)
}

// RemoveProvider remove provider instance p from the list of available
func (s *baseSwitcher) RemoveProvider(p providers.Provider) {
	s.pMux.Lock()
	defer s.pMux.Unlock()
	for i, cur := range s.provs {
		if cur == p {
			s.provs = append(s.provs[:i], s.provs[i+1:]...)
		}
	}
}

func (s *baseSwitcher) SetDispNum(n int) {
	s.dispnum = n
}

// Switch to new wallpaper
func (s *baseSwitcher) Switch() error {
	if len(s.provs) == 0 {
		return fmt.Errorf("no providers available to switch")
	}
	rand.Seed(time.Now().Unix())
	return s.switchWallpaper(s.provs[rand.Intn(len(s.provs))])
}

// Refresh update all active providers
func (s *baseSwitcher) Refresh() error {
	if len(s.provs) == 0 {
		return nil
	}
	g := errgroup.Group{}
	for _, p := range s.provs {
		p := p
		g.Go(func() error {
			return p.Refresh()
		})
	}
	return g.Wait()
}

// SaveCur saves current wallpaper to the specified local path.
// Path should exist
func (s *baseSwitcher) SaveCur(path string) error {
	if s.current == "" {
		return fmt.Errorf("no current wallpaper to save")
	}
	return internal.Copy(s.current, path)
}

// setFromFile sets the wallpaper for the current user
func setFromFile(filename string) error {
	filenameUTF16, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return err
	}

	if _, _, err := systemParametersInfo.Call(
		uintptr(spiSetDeskWallpaper),
		uintptr(uiParam),
		uintptr(unsafe.Pointer(filenameUTF16)),
		uintptr(spifUpdateINIFile|spifSendChange),
	); err != nil {
		// TODO: Always timeout error. Need filter
		log.Print(err)
	}
	return nil
}

func (s *baseSwitcher) switchWallpaper(p providers.Provider) error {
	i := 0
	paths := make([]string, s.dispnum)
	for i < s.dispnum {
		url, err := p.Random()
		if err != nil {
			return fmt.Errorf("error getting random url, might be empty list, try to refresh: %v", err)
		}
		paths[i], err = downloadPic(url, s.cachePath)
		if err != nil {
			return fmt.Errorf("erorr while downloading pic: %v", err)
		}
		i++
	}

	img, err := s.mergeImage(paths)
	if err != nil {
		return fmt.Errorf("error while post-processing images: %v", err)
	}
	if err := setFromFile(img); err != nil {
		return fmt.Errorf("error setting wallpaper from file: %v", err)
	}
	s.current = img
	return nil
}

func downloadPic(url, dest string) (string, error) {
	// check if file is already local
	if internal.FileExist(url) == nil {
		return url, nil
	}

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	fext := filepath.Ext(url)
	if fext == "" {
		fext = ".jpg"
	}

	fname := internal.RandStringBytes(16) + fext
	p := filepath.Join(dest, fname)

	// Create the file
	out, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	if _, err = io.Copy(out, resp.Body); err != nil {
		return "", err
	}
	return filepath.Abs(p)
}
