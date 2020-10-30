package main

import (
	"os"

	"gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	Name        string  `yaml:"name" json:"name"`
	Environment string  `yaml:"environment" json:"environment"`
	Port        string  `yaml:"port" json:"port"`
	Proxies     []Proxy `yaml:"proxies" json:"proxies"`
}

func readConfig(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var cfg Config

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
