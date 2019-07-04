// Package reddit provides service
// for getting new wallpapers from Reddit
package reddit

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

// Provider provides new wallpapers from subreddits
type Provider struct {
	siteurl string
	newpics []string
}

const (
	useragent string = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:68.0) Gecko/20100101 Firefox/68.0"
)

// New Reddit wallpapers provider
func New() *Provider {
	return &Provider{
		siteurl: "https://www.reddit.com/r/wallpaper/top/.json?t=month&limit=100",
	}
}

// Refresh lists of subreddit wallpapers
func (r *Provider) Refresh() error {
	client := http.DefaultClient

	req, err := http.NewRequest("GET", r.siteurl, nil)
	if err != nil {
		return fmt.Errorf("error while creating request: %v", err)
	}
	req.Header.Add("User-Agent", useragent)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error while sending request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("not successfull request, code: %v, %v", resp.StatusCode, resp.Status)
	}

	v := &response{}
	json.NewDecoder(resp.Body).Decode(v)
	for _, item := range v.Data.Children {
		if item.Data.Ups < 10 || strings.Contains(item.Data.Title, "1920") {
			continue
		}
		r.newpics = append(r.newpics, item.Data.URL)
	}

	return nil
}

// Random gets random picture (wallpaper) from source,
// returns path to downloaded file
func (r *Provider) Random() (string, error) {
	rand.Seed(time.Now().Unix())
	return r.newpics[rand.Intn(len(r.newpics))], nil
}
