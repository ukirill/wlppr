package switcher

// TODO: add to switcher
// DONE: 1. Refresh
// 2. Save current to pics (kinda Favs)
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
	"syscall"
	"time"
	"unsafe"

	"github.com/ukirill/wlppr-go/internal"
	"github.com/ukirill/wlppr-go/providers"
)

// https://msdn.microsoft.com/en-us/library/windows/desktop/ms724947.aspx
const (
	spiGetDeskWallpaper = 0x0073
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Switcher uses providers to get random wallpaper and place it on desktop
type Switcher struct {
	provs      []providers.Provider
	resH       int
	resW       int
	MonitorNum int
}

func New(p ...providers.Provider) *Switcher {
	return &Switcher{
		provs:      p,
		resH:       1080,
		resW:       1920,
		MonitorNum: 1,
	}
}

// Add provider as source for wallpaper
func (s *Switcher) Add(p ...providers.Provider) {
	s.provs = append(s.provs, p...)
}

// Switch to new wallpaper
// Receives number of monitors
func (s *Switcher) Switch() error {
	rand.Seed(time.Now().Unix())
	return s.switchWallpaper(s.provs[rand.Intn(len(s.provs))])
}

// setFromFile sets the wallpaper for the current user.
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
	return nil
}

func downloadPic(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	fext := filepath.Ext(url)
	if fext == "" {
		fext = ".jpg"
	}

	fname := randStringBytes(16) + fext
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

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
