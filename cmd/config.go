package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Providers []struct {
		name string `yaml:"name"`
		url  string `yaml:"port"`
	} `yaml:"providers"`

	Timers []struct {
		name    string `yaml:"name"`
		minutes uint   `yaml:"minutes"`
	} `yaml:"timers"`
}

func readFile(cfg *Config) error {
	f, err := os.Open("config.yml")
	if err != nil {
		return err
	}

	decoder := yaml.NewDecoder(f)
	if err = decoder.Decode(cfg); err != nil {
		return err
	}
	return nil
}
