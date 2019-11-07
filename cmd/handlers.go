package main

import (
	"log"

	"github.com/lxn/walk"
	"golang.org/x/sync/errgroup"

	"github.com/ukirill/wlppr-go/providers"
	"github.com/ukirill/wlppr-go/switcher"
)

func monitorHandler(sw *switcher.Switcher, n int) walk.EventHandler {
	return func() {
		sw.MonitorNum = n
	}
}

func timeoutHandler(as *switcher.AutoSwitcher, minutes uint) walk.EventHandler {
	return func() {
		as.SetTimeout(minutes)
	}
}

func exitHandler() {
	walk.App().Exit(0)
}

func favHandler(sw *switcher.Switcher, path string) walk.EventHandler {
	return func() {
		if err := sw.SaveCur(path); err != nil {
			log.Printf("error on saving fav wlppr : %v", err)
		}
	}
}

func refreshHandler(provs ...providers.Provider) walk.EventHandler {
	return func() {
		if err := refreshProviders(provs...); err != nil {
			log.Printf("error while refreshing providers : %v", err)
		}
	}
}

func refreshProviders(provs ...providers.Provider) error {
	g := errgroup.Group{}
	for _, p := range provs {
		p := p
		g.Go(func() error {
			return p.Refresh()
		})
	}
	return g.Wait()
}

func switchHandler(sw *switcher.Switcher) walk.EventHandler {
	//go sw.Switch()
	return func() {
		if err := sw.Switch(); err != nil {
			log.Printf("error switching wlppr : %v", err)
		}
	}
}
