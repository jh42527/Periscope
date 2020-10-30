package main

import (
	"errors"
)

// TO-DO: Add support for auth headers on target
// LONG-TERM:
// Add support for auth headers on target, target server using k/v store
// Add support for virtual hosts / access control list
// Add support for rate limiting

// Proxy type
type Proxy struct {
	Name   string `yaml:"name" json:"name"`
	Method string `yaml:"method" json:"method"`
	Path   string `yaml:"path" json:"path"`
	Target string `yaml:"target" json:"target"`
	Cache  string `yaml:"cache" json:"cache"`
}

// FindProxy function
func FindProxy(proxies []Proxy, path string) (*Proxy, error) {
	for _, proxy := range proxies {
		if proxy.Path == path {
			return &proxy, nil
		}
	}

	return nil, errors.New("Proxy Not Found")
}
