package switcher

// TODO: add to switcher
// DONE: 1. Refresh
// DONE: 2. Save current to pics (kinda Favs)
// DONE: 3. Switch by timeout setting
// DONE: 4. Multimonitor

import (
	"fmt"
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

	"github.com/ukirill/wlppr-go/internal"
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

// Switcher uses providers to get random wallpaper and place it on desktop
type Switcher struct {
	provs      []providers.Provider
	pMux       *sync.Mutex
	resH       int
	resW       int
	current    string
	MonitorNum int
}

func New(p ...providers.Provider) *Switcher {
	return &Switcher{
		provs:      p,
		pMux:       &sync.Mutex{},
		resH:       1080,
		resW:       1920,
		MonitorNum: 1,
	}
}

// Set providers as source for wallpaper
func (s *Switcher) SetProviders(p ...providers.Provider) {
	s.provs = p
}

// AddProvider adds provider p if it isnt added already
func (s *Switcher) AddProvider(p providers.Provider) {
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
func (s *Switcher) RemoveProvider(p providers.Provider) {
	s.pMux.Lock()
	defer s.pMux.Unlock()
	for i, cur := range s.provs {
		if cur == p {
			s.provs = append(s.provs[:i], s.provs[i+1:]...)
		}
	}
}

// Switch to new wallpaper
func (s *Switcher) Switch() error {
	if len(s.provs) == 0 {
		return fmt.Errorf("no providers available to switch")
	}
	rand.Seed(time.Now().Unix())
	return s.switchWallpaper(s.provs[rand.Intn(len(s.provs))])
}

// Refresh update all active providers
func (s *Switcher) Refresh() error {
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
func (s *Switcher) SaveCur(path string) error {
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

func (s *Switcher) switchWallpaper(p providers.Provider) error {
	i := 0
	paths := make([]string, s.MonitorNum)
	for i < s.MonitorNum {
		url, err := p.Random()
		if err != nil {
			return fmt.Errorf("error getting random url, might be empty list, try to refresh: %v", err)
		}
		paths[i], err = downloadPic(url)
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

func downloadPic(url string) (string, error) {
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
	p, err := internal.GetCachePath(fname)
	if err != nil {
		return "", err
	}

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
