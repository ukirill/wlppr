package main

// TODO: better logging
// ? remove fatal everywhere? notification maybe
// ? struct + methods for "stateful tray icon"

import (
	"log"
	"os"

	"github.com/lxn/walk"

	"github.com/ukirill/wlppr-go/providers"
	"github.com/ukirill/wlppr-go/providers/reddit"
	"github.com/ukirill/wlppr-go/switcher"
)

// Program state
var (
	sw *switcher.Switcher
	as *switcher.AutoSwitcher

	// TODO: make switcher able to refresh provs and remove
	provs []providers.Provider
)

func main() {
	f, err := os.OpenFile("wlppr.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("Wlppr starts")

	rd1 := reddit.New("https://www.reddit.com/r/wallpaper/hot/.json?t=month&limit=100")
	rd2 := reddit.New("https://www.reddit.com/r/wallpapers/hot/.json?t=month&limit=100")
	sw = switcher.New(rd1, rd2)
	as = switcher.NewAutoSwitcher(sw, 15)
	provs = []providers.Provider{rd1, rd2}
	log.Println("Providers created")

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
	wlpprAct, err := addNewAction("W&LPPR!", ni.ContextMenu().Actions(), switchHandler(sw))
	if err != nil {
		log.Fatal(err)
	}
	if err = wlpprAct.SetEnabled(false); err != nil {
		log.Fatal(err)
	}

	addMonitorMenu(ni.ContextMenu().Actions())
	addTimeoutMenu(ni.ContextMenu().Actions())

	// Action for refreshing providers sources
	refAct, err := addNewAction("R&efresh source", ni.ContextMenu().Actions(), refreshHandler(provs...))
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
		if err := refreshProviders(rd1, rd2); err != nil {
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

func addMonitorMenu(actions *walk.ActionList) {
	monitorNumMenu, _ := walk.NewMenu()
	monitorNumMenuAct, _ := actions.AddMenu(monitorNumMenu)
	monitorNumMenuAct.SetText("Monitors")
	monitorNumMenuAct.SetToolTip("Set number of monitors")
	oneAct, err := addNewRadioAction("1", monitorNumMenu.Actions(), monitorHandler(sw, 1))
	if err != nil {
		log.Fatal(err)
	}
	oneAct.SetChecked(true)
	_, err = addNewRadioAction("2", monitorNumMenu.Actions(), func() {
		monitorHandler(sw, 2)()
	})
	if err != nil {
		log.Fatal(err)
	}
}

func addTimeoutMenu(actions *walk.ActionList) {
	timeoutMenu, _ := walk.NewMenu()
	timeoutMenuAct, _ := actions.AddMenu(timeoutMenu)
	timeoutMenuAct.SetText("Timeout")
	timeoutMenuAct.SetToolTip("Set timeout for refreshing wallpapers")
	offAct, _ := addNewRadioAction("off", timeoutMenu.Actions(), timeoutHandler(as, 0))
	offAct.SetChecked(true)
	addNewRadioAction("15 min", timeoutMenu.Actions(), timeoutHandler(as, 15))
	addNewRadioAction("1 hour", timeoutMenu.Actions(), timeoutHandler(as, 60))
}
