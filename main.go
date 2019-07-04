package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/getlantern/systray"
	mm "github.com/ukirill/wlppr-go/providers/moviemania"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("resources/icon.ico"))
	systray.SetTitle("Wlppr!")
	m := mm.New()
	refresh(m)
	n := systray.AddMenuItem("New wlppr", "Get new random wallpaper")
	r := systray.AddMenuItem("Refresh wlpprs' list", "Refresh wallpapers list from Moviemania.io")
	r.Hide()
	q := systray.AddMenuItem("Quit", "Quit the app")
	fmt.Println("ready")
	d := time.Hour * 2
	t := time.NewTimer(d)
	for {
		select {
		case <-n.ClickedCh:
			go switchWallpaper(m)
			if !t.Stop() {
				<-t.C
			}
			t.Reset(d)
		case <-t.C:
			go switchWallpaper(m)
			t.Reset(d)
		case <-r.ClickedCh:
			go func() {
				refresh(m)
			}()
		case <-q.ClickedCh:
			systray.Quit()
		}

	}
}

func onExit() {

}

func refresh(m *mm.Provider) {
	systray.SetTooltip("Initialize wlppr list")
	m.Refresh()
	systray.SetTooltip("Wlppr switcher")
}

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}
