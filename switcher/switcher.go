package switcher

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"syscall"
	"time"
	"unsafe"

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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// https://msdn.microsoft.com/en-us/library/windows/desktop/ms724947.aspx
var (
	user32               = syscall.NewLazyDLL("user32.dll")
	systemParametersInfo = user32.NewProc("SystemParametersInfoW")
)

// Switcher uses providers to get random wallpaper and place it on desktop
type Switcher struct {
	provs []providers.Provider
}

func New(p ...providers.Provider) *Switcher {
	return &Switcher{provs: p}
}

// Add provider as source for wallpaper
func (s *Switcher) Add(p ...providers.Provider) {
	s.provs = append(s.provs, p...)
}

// Switch to new wallpaper
func (s *Switcher) Switch() error {
	rand.Seed(time.Now().Unix())
	return switchWallpaper(s.provs[rand.Intn(len(s.provs))])
}

// setFromFile sets the wallpaper for the current user.
func setFromFile(filename string) error {
	filenameUTF16, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return err
	}

	systemParametersInfo.Call(
		uintptr(spiSetDeskWallpaper),
		uintptr(uiParam),
		uintptr(unsafe.Pointer(filenameUTF16)),
		uintptr(spifUpdateINIFile|spifSendChange),
	)
	return nil
}

func switchWallpaper(p providers.Provider) error {
	url, err := p.Random()
	if err != nil {
		return fmt.Errorf("error getting random url, might be empty list, try to refresh: %v", err)
	}
	path, err := downloadPic(url)
	if err != nil {
		return fmt.Errorf("erorr while downloading pic: %v", err)
	}

	setFromFile(path)
	return nil
}

func downloadPic(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	fname := randStringBytes(16) + ".jpg"
	p := path.Join("cache", fname)

	// Create the file
	out, err := os.Create(p)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return filepath.Abs(p)
}

func randStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}