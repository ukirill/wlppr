package switcher

import (
	"log"
	"time"

	"github.com/ukirill/wlppr-go/provider"
)

type SwitchEventHandle func(error)

type AutoSwitcher interface {
	Switcher
	SetTimeout(minutes uint)
	Start(handle SwitchEventHandle)
	Stop()
}

type timeSwitcher struct {
	*baseSwitcher
	timeout uint
	cancel  chan interface{}
	handle  SwitchEventHandle
}

func NewAutoSwitcher(minutes uint, cachePath string, handle SwitchEventHandle, prov ...provider.Provider) AutoSwitcher {
	return &timeSwitcher{
		baseSwitcher: New(cachePath, prov...),
		timeout:      minutes,
		cancel:       make(chan interface{}),
		handle:       handle,
	}
}

func (as *timeSwitcher) SetTimeout(minutes uint) {
	as.timeout = minutes
	as.Stop()
	as.Start(as.handle)
}

func (as *timeSwitcher) Start(handle SwitchEventHandle) {
	if as.timeout == 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(time.Duration(as.timeout) * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-as.cancel:
				return
			case <-ticker.C:
				err := as.Switch()
				if err != nil {
					log.Printf("error on autoswitching wallpaper: %v", err)
				}
				handle(err)
			}
		}
	}()
}

func (as *timeSwitcher) Stop() {
	close(as.cancel)
	as.cancel = make(chan interface{})
}
