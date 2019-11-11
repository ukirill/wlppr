package main

// TODO: better logging
// ? remove fatal everywhere? notification maybe

import (
	"log"
	"os"
	"path/filepath"

	"github.com/lxn/walk"

	"github.com/ukirill/wlppr-go/cmd/internal"
	"github.com/ukirill/wlppr-go/providers"
	"github.com/ukirill/wlppr-go/providers/local"
	"github.com/ukirill/wlppr-go/providers/reddit"
	"github.com/ukirill/wlppr-go/switcher"
)

var (
	sw switcher.AutoSwitcher
)

func main() {
	appData, err := internal.GetAppDataDir()
	if err != nil {
		log.Fatalf("error on getting AppData directory : %v", err)
	}
	logfile := filepath.Join(appData, "wlppr.log")
	f, err := os.OpenFile(logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening file: %v", err)
	}
	defer f.Close()

	log.SetOutput(f)
	log.Println("Wlppr starts")

	rd1 := reddit.New("Reddit wallpaper", "https://www.reddit.com/r/wallpaper/hot/.json?t=month&limit=100")
	rd2 := reddit.New("Reddit wallpapers", "https://www.reddit.com/r/wallpapers/hot/.json?t=month&limit=100")

	favPath, _ := internal.GetFavDir()
	fav := local.New("Favourites", favPath)

	cachePath, _ := internal.GetCachePath("")
	sw = switcher.NewAutoSwitcher(0, cachePath, rd1, rd2, fav)
	log.Println("Switcher created")

	mw, err := walk.NewMainWindow()
	if err != nil {
		log.Fatal(err)
	}

	// We load our icon from embedded resource
	icon, err := walk.Resources.Icon("10")
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
	hp := func(action *walk.Action) walk.EventHandler {
		return switchHandler(sw)
	}
	wlpprAct, err := addNewAction("W&LPPR!", ni.ContextMenu().Actions(), hp)
	if err != nil {
		log.Fatal(err)
	}
	wlpprAct.SetToolTip("Get the new one!")
	if err = wlpprAct.SetEnabled(false); err != nil {
		log.Fatal(err)
	}

	addMonitorMenu(ni.ContextMenu().Actions())
	addTimeoutMenu(ni.ContextMenu().Actions())
	addProviderMenu(ni.ContextMenu().Actions(), sw.Providers()...)

	// Action for refreshing providers sources
	hp = func(action *walk.Action) walk.EventHandler {
		return refreshHandler(sw)
	}
	refAct, err := addNewAction("R&efresh source", ni.ContextMenu().Actions(), hp)
	if err != nil {
		log.Fatal(err)
	}
	if err := refAct.SetEnabled(false); err != nil {
		log.Fatal(err)
	}

	// Action for save fav
	hp = func(action *walk.Action) walk.EventHandler {
		return favHandler(sw, favPath)
	}
	if err != nil {
		log.Fatal(err)
	}
	if _, err := addNewAction("Save to favs", ni.ContextMenu().Actions(), hp); err != nil {
		log.Fatal(err)
	}

	// Action for exit
	hp = func(action *walk.Action) walk.EventHandler {
		return exitHandler
	}
	if _, err := addNewAction("E&xit", ni.ContextMenu().Actions(), hp); err != nil {
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

	func() {
		defer logPanic()

		// Run the message loop.
		mw.Run()
	}()
}

func logPanic() {
	if p := recover(); p != nil {
		log.Fatalf("[PANIC] program panic : %v", p)
	}
}

func addMonitorMenu(actions *walk.ActionList) {
	monitorNumMenu, _ := walk.NewMenu()
	monitorNumMenuAct, _ := actions.AddMenu(monitorNumMenu)
	monitorNumMenuAct.SetText("Monitors")
	monitorNumMenuAct.SetToolTip("Set number of monitors")
	oneAct, err := addNewRadioAction("1", monitorNumMenu.Actions(), monitorHandler(sw, 1), nil)
	if err != nil {
		log.Fatal(err)
	}
	oneAct.SetChecked(true)
	_, err = addNewRadioAction("2", monitorNumMenu.Actions(), monitorHandler(sw, 2), nil)
	if err != nil {
		log.Fatal(err)
	}
}

func addTimeoutMenu(actions *walk.ActionList) {
	timeoutMenu, _ := walk.NewMenu()
	timeoutMenuAct, _ := actions.AddMenu(timeoutMenu)
	timeoutMenuAct.SetText("Timeout")
	timeoutMenuAct.SetToolTip("Set timeout for refreshing wallpapers")
	offAct, _ := addNewRadioAction("off", timeoutMenu.Actions(), timeoutHandler(sw, 0), nil)
	offAct.SetChecked(true)
	addNewRadioAction("15 min", timeoutMenu.Actions(), timeoutHandler(sw, 15), nil)
	addNewRadioAction("1 hour", timeoutMenu.Actions(), timeoutHandler(sw, 60), nil)
}

func addProviderMenu(actions *walk.ActionList, provs ...providers.Provider) {
	provMenu, _ := walk.NewMenu()
	provMenuAction, _ := actions.AddMenu(provMenu)
	provMenuAction.SetText("Sources")
	provMenuAction.SetToolTip("Choose Wlppr sources")
	for _, p := range provs {
		addNewCheckableAction(p.Title(), provMenu.Actions(), true,
			provHandler(sw, p, true), provHandler(sw, p, false))
	}
}
