package main

//TODO: better logging
// ? remove fatal everywhere? notification maybe
// ? struct + methods for "statful tray icon"

import (
	"log"

	"github.com/lxn/walk"
	"github.com/ukirill/wlppr-go/providers"
	"github.com/ukirill/wlppr-go/providers/moviemania"
	"github.com/ukirill/wlppr-go/providers/reddit"
	"github.com/ukirill/wlppr-go/switcher"
	"golang.org/x/sync/errgroup"
)

var sw *switcher.Switcher

// TODO: make switcher able to refresh provs and remove
var provs []providers.Provider

func main() {
	mm := moviemania.New()
	rd1 := reddit.New("https://www.reddit.com/r/wallpaper/top/.json?t=month&limit=100")
	rd2 := reddit.New("https://www.reddit.com/r/wallpapers/top/.json?t=month&limit=100")
	sw = switcher.New(mm, rd1, rd2)
	provs = []providers.Provider{mm, rd1, rd2}
	mw, err := walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}

	// We load our icon from a file.
	icon, err := walk.Resources.Icon("../resources/icon.ico")
	if err != nil {
		log.Fatal(err)
	}

	// Create the notify icon and make sure we clean it up on exit.
	ni, err := walk.NewNotifyIcon(mw)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := ni.Dispose(); err != nil {
			log.Print(err)
		}
	}()

	// Set the icon and a tool tip text.
	if err := ni.SetIcon(icon); err != nil {
		log.Fatal(err)
	}
	if err := ni.SetToolTip("WLPPR!"); err != nil {
		log.Fatal(err)
	}

	wlpprAct, err := addNewAction("W&LPPR!", ni, switchHandler)
	if err != nil {
		log.Fatal(err)
	}
	if err := wlpprAct.SetEnabled(false); err != nil {
		log.Fatal(err)
	}

	refAct, err := addNewAction("R&efresh source", ni, refreshHandler)
	if err != nil {
		log.Fatal(err)
	}
	if err := refAct.SetEnabled(false); err != nil {
		log.Fatal(err)
	}

	if _, err := addNewAction("E&xit", ni, exitHandler); err != nil {
		log.Fatal(err)
	}

	go func() {
		// The notify icon is hidden initially, so we have to make it visible.
		if err := ni.SetVisible(true); err != nil {
			log.Fatal(err)
		}

		// Now that the icon is visible, we can bring up an info balloon.
		if err := ni.ShowInfo("Wlppr is starting", "Functions will be available after providers init'd"); err != nil {
			log.Fatal(err)
		}
		if err := refreshProviders(provs...); err != nil {
			log.Fatal(err)
		}
		if err := ni.ShowInfo("Wlppr init'd", "Use context menu on tray icon"); err != nil {
			log.Fatal(err)
		}

		if err := wlpprAct.SetEnabled(true); err != nil {
			log.Fatal(err)
		}
		if err := refAct.SetEnabled(true); err != nil {
			log.Fatal(err)
		}
	}()

	// Run the message loop.
	mw.Run()
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

func addNewAction(name string, ni *walk.NotifyIcon, handler walk.EventHandler) (*walk.Action, error) {
	actions := ni.ContextMenu().Actions()
	action := walk.NewAction()
	if err := action.SetText(name); err != nil {
		return nil, err
	}
	action.Triggered().Attach(func() { handler() })
	if err := actions.Add(action); err != nil {
		return nil, err
	}

	return action, nil
}

func exitHandler() {
	walk.App().Exit(0)
}

func refreshHandler() {
	if err := refreshProviders(provs...); err != nil {
		log.Fatal(err)
	}
}

func switchHandler() {
	if err := sw.Switch(); err != nil {
		log.Fatal(err)
	}
}
