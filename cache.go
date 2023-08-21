package main

import (
	"time"
)

type cache struct {
	timeout time.Duration
	store   map[string]cacheElement
}

func NewCache(timeout time.Duration) *cache {
	return &cache{
		timeout: timeout,
		store:   map[string]cacheElement{},
	}
}

func (c *cache) Add(key string, data []byte) {
	c.store[key] = cacheElement{
		cachedAt: time.Now(),
		data:     data,
	}
}

func (c *cache) Get(key string) ([]byte, bool) {
	elem, ok := c.store[key]
	if !ok {
		return nil, false
	}
	age := time.Since(elem.cachedAt)
	if age >= c.timeout {
		delete(c.store, key)
		return nil, false
	}
	return elem.data, true
}

func (c *cache) Purge() {
	c.store = make(map[string]cacheElement)
}

type cacheElement struct {
	data     []byte
	cachedAt time.Time
}
