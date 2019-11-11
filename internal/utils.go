package internal

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
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
