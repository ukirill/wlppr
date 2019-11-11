package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

var wlpprData string
var cache string
var favs string

// GetAppDataDir returns path to wlppr data directory. Creates if it doesnt exist
func GetAppDataDir() (string, error) {
	if wlpprData != "" {
		return wlpprData, nil
	}
	appdata := os.Getenv("APPDATA")
	wlpprData = filepath.Join(appdata, "Wlppr")
	if err := createNotExist(wlpprData); err != nil {
		return "", fmt.Errorf("error creatin data folder : %v", err)
	}
	return wlpprData, nil
}

// GetAppDataPath returns path to dir in Appdata directory. Creates if it doesnt exist
func GetAppDataPath(entry string) (string, error) {
	appdata, err := GetAppDataDir()
	if err != nil {
		return "", err
	}
	p := filepath.Join(appdata, entry)
	if err := createNotExist(p); err != nil {
		return "", fmt.Errorf("error creatin new folder : %v", err)
	}
	return p, nil
}

// GetCachePath returns path to entry in cache directory. Creates cache directory if it doesnt exist
func GetCachePath(entry string) (string, error) {
	if cache != "" {
		return filepath.Join(cache, entry), nil
	}
	c, err := GetAppDataPath("Cache")
	if err != nil {
		return "", err
	}
	cache = c
	return filepath.Join(cache, entry), nil
}

// GetFavDir returns path to Favourite wallpapers directory
func GetFavDir() (string, error) {
	if favs != "" {
		return favs, nil
	}
	var err error
	favs, err = GetAppDataPath("Favs")
	return favs, err
}

func createNotExist(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if os.IsNotExist(err) {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return fmt.Errorf("error creating folder : %v", err)
		}
		return nil
	}
	return err
}
