package switcher

import (
	"log"
	"time"

	"github.com/ukirill/wlppr-go/provider"
)

type AutoSwitcher interface {
	Switcher
	SetTimeout(minutes uint)
	Start()
	Stop()
}

type timeSwitcher struct {
	*baseSwitcher
	timeout uint
	cancel  chan interface{}
}

func NewAutoSwitcher(minutes uint, cachePath string, prov ...provider.Provider) AutoSwitcher {
	return &timeSwitcher{
		New(cachePath, prov...),
		minutes,
		make(chan interface{}),
	}
}

func (as *timeSwitcher) SetTimeout(minutes uint) {
	as.timeout = minutes
	as.Stop()
	as.Start()
}

func (as *timeSwitcher) Start() {
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
				if err := as.Switch(); err != nil {
					log.Printf("error on autoswitching wallpaper: %v", err)
				}
			}
		}
	}()
}

func (as *timeSwitcher) Stop() {
	close(as.cancel)
	as.cancel = make(chan interface{})
}
