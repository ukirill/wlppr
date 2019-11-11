// Package moviemania provides service
// for getting new wallpapers from Moviemania.io
package moviemania

import (
	"math/rand"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/anaskhan96/soup"
)

// Provider provides random wallpaper posters from Moviemania.io
type Provider struct {
	siteurl string
	// Map of genre: base_url_to_genre_pics
	genreurls map[string]string
	// List of all pics urls
	allpics []string
	// Map of all pics grouped by genre
	genrepics map[string][]string
	mu        sync.Mutex
}

// New Moviemania provider
func New() *Provider {
	return &Provider{
		siteurl:   "https://www.moviemania.io",
		allpics:   make([]string, 0, 500),
		genreurls: map[string]string{},
		genrepics: map[string][]string{},
	}
}

func (m *Provider) Title() string {
	return "Moviemania"
}

// Refresh wallpapers data from Moviemania.io
func (m *Provider) Refresh() error {
	if err := m.refreshGenreUrls(); err != nil {
		return err
	}
	if err := m.refreshPics(); err != nil {
		return err
	}
	m.allpics = removeDuplicatesUnordered(m.allpics)
	return nil
}

// Random gets random picture (wallpaper) from source,
// returns path to downloaded file
func (m *Provider) Random() (string, error) {
	// Get random picture
	rand.Seed(time.Now().Unix())
	return getURL(m.siteurl, m.allpics[rand.Intn(len(m.allpics))], "1920x1080")
}

func (m *Provider) refreshGenreUrls() error {
	u, _ := getURL(m.siteurl, "desktop")

	resp, err := soup.Get(u)
	if err != nil {
		return err
	}

	doc := soup.HTMLParse(resp)
	links := doc.FindAll("a", "class", "menu-item")
	for _, a := range links {
		href := a.Attrs()["href"]
		if strings.Contains(href, "genre") {
			e := strings.Split(href, "/")
			genre := e[len(e)-1]
			genreurl, _ := getURL(m.siteurl, href)
			m.genreurls[genre] = genreurl + "?offset="
		}
	}

	return nil
}

func (m *Provider) refreshPics() error {
	var wg sync.WaitGroup
	for genre, url := range m.genreurls {
		m.genrepics[genre] = []string{}
		wg.Add(1)
		go func(url, genre string) {
			defer wg.Done()
			m.getGenrePics(genre, url)
		}(url, genre)
	}
	wg.Wait()
	return nil
}

// getGenrePics get all links to pics by genre
func (m *Provider) getGenrePics(genre, url string) error {
	offset := 0
	for {
		genreurl := url + strconv.Itoa(offset)
		resp, err := soup.Get(genreurl)

		if err != nil {
			return err
		}

		doc := soup.HTMLParse(resp)
		links := doc.FindAll("a", "class", "wallpaper")
		if len(links) == 0 {
			return nil
		}

		pics := []string{}
		for _, l := range links {
			p := strings.Replace(strings.Split(l.Attrs()["href"], "-")[0], "wallpaper", "download", 1)
			pics = append(pics, p)
		}

		m.addPicURL(genre, pics)
		offset += len(links)
	}
}

func (m *Provider) addPicURL(genre string, pics []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.genrepics[genre] = append(m.genrepics[genre], pics...)
	m.allpics = append(m.allpics, pics...)
}

func getURL(first string, el ...string) (string, error) {
	u, err := url.Parse(first)
	if err != nil {
		return "", err
	}
	if len(el) > 0 {
		u.Path = path.Join(u.Path, path.Join(el...))
	}
	return u.String(), nil
}

func removeDuplicatesUnordered(elements []string) []string {
	encountered := map[string]bool{}

	// Create a map of all unique elements.
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := []string{}
	for key := range encountered {
		result = append(result, key)
	}
	return result
}
