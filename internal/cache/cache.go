package cache

import (
	"time"
)

type Cache struct {
	timeout time.Duration
	store   map[string]cacheElement
}

func New(timeout time.Duration) *Cache {
	return &Cache{
		timeout: timeout,
		store:   map[string]cacheElement{},
	}
}

func (c *Cache) Add(key string, data []byte) {
	c.store[key] = cacheElement{
		cachedAt: time.Now(),
		data:     data,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
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

func (c *Cache) Purge() {
	c.store = make(map[string]cacheElement)
}

type cacheElement struct {
	data     []byte
	cachedAt time.Time
}
