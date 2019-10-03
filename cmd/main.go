package main

//TODO: better logging
// ? remove fatal everywhere? notification maybe
// ? struct + methods for "statful tray icon"

import (
	"log"

	"github.com/lxn/walk"
	"github.com/ukirill/wlppr-go/providers"

	//"github.com/ukirill/wlppr-go/providers/moviemania"
	"github.com/ukirill/wlppr-go/providers/reddit"
	"github.com/ukirill/wlppr-go/switcher"
	"golang.org/x/sync/errgroup"
)

var sw *switcher.Switcher

// TODO: make switcher able to refresh provs and remove
var provs []providers.Provider

func main() {
	//mm := moviemania.New()
	rd1 := reddit.New("https://www.reddit.com/r/wallpaper/hot/.json?t=year&limit=100")
	rd2 := reddit.New("https://www.reddit.com/r/wallpapers/hot/.json?t=month&limit=100")
	sw = switcher.New(rd1, rd2)
	provs = []providers.Provider{rd1, rd2}
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

	// Action for switching wlpprs
	wlpprAct, err := addNewAction("W&LPPR!", ni.ContextMenu().Actions(), switchHandler)
	if err != nil {
		log.Fatal(err)
	}
	if err = wlpprAct.SetEnabled(false); err != nil {
		log.Fatal(err)
	}

	addMonitorMenu(ni.ContextMenu().Actions())

	// Action for refreshing providers sources
	refAct, err := addNewAction("R&efresh source", ni.ContextMenu().Actions(), refreshHandler)
	if err != nil {
		log.Fatal(err)
	}
	if err := refAct.SetEnabled(false); err != nil {
		log.Fatal(err)
	}

	// Action for exit
	if _, err := addNewAction("E&xit", ni.ContextMenu().Actions(), exitHandler); err != nil {
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

func addNewAction(name string, actions *walk.ActionList, handler walk.EventHandler) (*walk.Action, error) {
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

func addMonitorMenu(actions *walk.ActionList) {

	monitorNumMenu, _ := walk.NewMenu()
	monitorNumMenuAct, _ := actions.AddMenu(monitorNumMenu)
	monitorNumMenuAct.SetText("Monitors")
	monitorNumMenuAct.SetToolTip("Set number of monitors")
	oneAct, err := addNewAction("1", monitorNumMenu.Actions(), func() {})
	if err != nil {
		log.Fatal(err)
	}
	oneAct.SetCheckable(true)
	oneAct.SetChecked(true)
	oneAct.Triggered().Attach(monitorHandler(1))
	twoAct, err := addNewAction("2", monitorNumMenu.Actions(), func() {})
	if err != nil {
		log.Fatal(err)
	}
	twoAct.SetCheckable(true)
	twoAct.Triggered().Attach(monitorHandler(2))
	oneAct.Triggered().Attach(func() {
		oneAct.SetChecked(true)
		twoAct.SetChecked(false)
	})
	twoAct.Triggered().Attach(func() {
		oneAct.SetChecked(false)
		twoAct.SetChecked(true)
	})
}

func monitorHandler(n int) walk.EventHandler {
	return func() {
		sw.MonitorNum = n
	}
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
	go sw.Switch()
	// if err := sw.Switch(); err != nil {
	// 	log.Fatal(err)
	// }
}
