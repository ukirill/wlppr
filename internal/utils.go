package internal

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
)

// TODO: struct for global state or settings?
var wlpprData string
var cache string

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

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
	p := filepath.Join(appdata, entry)
	if err := createNotExist(p); err != nil {
		return "", fmt.Errorf("error creatin new folder : %v", err)
	}
	return p, nil
}

// GetCachePath returns path to entry in cache directory. Creates cache directory if not exists
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

func Copy(src, dst string) error {
	if err := FileExist(src); err != nil {
		return err
	}

	fn := filepath.Base(src)
	fulldst := filepath.Join(dst, fn)

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(fulldst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

func FileExist(path string) error {
	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("error on checking file existance : %v", err)
	}
	if !stat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", path)
	}
	return nil
}

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
