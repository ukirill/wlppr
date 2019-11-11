package main

import (
	"log"

	"github.com/lxn/walk"
	"golang.org/x/sync/errgroup"

	"github.com/ukirill/wlppr-go/provider"
	"github.com/ukirill/wlppr-go/switcher"
)

func monitorHandler(sw switcher.Switcher, n int) walk.EventHandler {
	return func() {
		sw.SetDispNum(n)
	}
}

func timeoutHandler(sw switcher.AutoSwitcher, minutes uint) walk.EventHandler {
	return func() {
		sw.SetTimeout(minutes)
	}
}

func exitHandler() {
	walk.App().Exit(0)
}

func favHandler(sw switcher.Switcher, path string) walk.EventHandler {
	return func() {
		if err := sw.SaveCur(path); err != nil {
			log.Printf("error on saving fav wlppr : %v", err)
		}
	}
}

func refreshHandler(sw switcher.Switcher) walk.EventHandler {
	return func() {
		if err := sw.Refresh(); err != nil {
			log.Printf("error while refreshing providers : %v", err)
		}
	}
}

func provHandler(sw switcher.Switcher, p provider.Provider, state bool) walk.EventHandler {
	if state {
		return func() { sw.AddProvider(p) }
	}
	return func() { sw.RemoveProvider(p) }
}

func refreshProviders(provs ...provider.Provider) error {
	g := errgroup.Group{}
	for _, p := range provs {
		p := p
		g.Go(func() error {
			return p.Refresh()
		})
	}
	return g.Wait()
}

func switchHandler(sw switcher.Switcher) walk.EventHandler {
	return func() {
		if err := sw.Switch(); err != nil {
			log.Printf("error switching wlppr : %v", err)
		}
	}
}
