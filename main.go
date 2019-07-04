package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/getlantern/systray"
	"github.com/ukirill/wlppr-go/providers"
	"github.com/ukirill/wlppr-go/providers/moviemania"
	"github.com/ukirill/wlppr-go/providers/reddit"
	"github.com/ukirill/wlppr-go/switcher"
	"golang.org/x/sync/errgroup"
)

func main() {
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetIcon(getIcon("resources/icon.ico"))
	systray.SetTitle("Wlppr!")
	mm := moviemania.New()
	rd := reddit.New()
	if err := refreshProviders(mm, rd); err != nil {
		// TODO: retry logic
		log.Fatal(err)
	}
	sw := switcher.New(mm, rd)

	wlppr := systray.AddMenuItem("New wlppr", "Get new random wallpaper")
	refresh := systray.AddMenuItem("Refresh wlpprs' list", "Refresh wallpapers list from Moviemania.io")
	refresh.Hide()
	quit := systray.AddMenuItem("Quit", "Quit the app")
	fmt.Println("ready")

	d := time.Hour * 2
	timer := time.NewTimer(d)
	for {
		select {
		case <-wlppr.ClickedCh:
			// TODO: handle error
			go sw.Switch()
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(d)
		case <-timer.C:
			go sw.Switch()
			timer.Reset(d)
		case <-refresh.ClickedCh:
			go func() {
				refreshProviders(mm, rd)
			}()
		case <-quit.ClickedCh:
			systray.Quit()
		}
	}
}

func onExit() {

}

func refreshProviders(provs ...providers.Provider) error {
	systray.SetTooltip("Initialize wlppr list")
	defer systray.SetTooltip("Wlppr switcher")

	g := errgroup.Group{}
	for _, p := range provs {
		p := p
		g.Go(func() error {
			log.Print(p)
			return p.Refresh()
		})
	}
	return g.Wait()
}

func getIcon(s string) []byte {
	b, err := ioutil.ReadFile(s)
	if err != nil {
		fmt.Print(err)
	}
	return b
}
