package main

import (
	"net/http"
	"sync"
	"time"
)

// TO-DO: Backend to persistent and shared store

// Cache stuct
type Cache struct {
	Items map[string]Item
	mu    *sync.RWMutex
}

// Item struct
type Item struct {
	ProxyResponse ProxyResponse
	Expiration    int64
}

// ProxyResponse stuct
type ProxyResponse struct {
	Content    []byte
	StatusCode int
	Header     http.Header
}

// Expired returns true if the item has expired.
func (item Item) Expired() bool {
	if item.Expiration == 0 {
		return false
	}

	return time.Now().UnixNano() > item.Expiration
}

// InitializeCache creates a new in memory storage
func InitializeCache() *Cache {
	return &Cache{
		Items: make(map[string]Item),
		mu:    &sync.RWMutex{},
	}
}

// Get a cached content by key
func (c Cache) Get(key string) *ProxyResponse {
	c.mu.RLock()

	defer c.mu.RUnlock()

	item := c.Items[key]

	if item.Expired() {
		delete(c.Items, key)

		return nil
	}

	return &item.ProxyResponse
}

// Set a cached content by key
func (c Cache) Set(key string, proxyResponse ProxyResponse, duration time.Duration) {
	c.mu.Lock()

	defer c.mu.Unlock()

	c.Items[key] = Item{
		ProxyResponse: proxyResponse,
		Expiration:    time.Now().Add(duration).UnixNano(),
	}
}
