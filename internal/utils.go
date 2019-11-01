package internal

import (
	"fmt"
	"os"
	"path/filepath"
)

var wlpprData string
var cache string

// GetAppDataDir returns path to wlppr data directory. Creates it if not exists
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

// GetAppDataPath returns path to entry in Appdata directory
func GetAppDataPath(entry string) (string, error) {
	appdata, err := GetAppDataDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(appdata, entry), nil
}

// GetCachePath returns path to entry in cache directory. Creates cache directory if not exists
func GetCachePath(entry string) (string, error) {
	if cache != "" {
		return filepath.Join(cache, entry), nil
	}
	c, err := GetAppDataPath("cache")
	if err != nil {
		return "", err
	}
	if err := createNotExist(c); err != nil {
		return "", fmt.Errorf("error creating cache folder : %v", err)
	}
	cache = c
	return filepath.Join(cache, entry), nil
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
