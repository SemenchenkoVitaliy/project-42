package cache

import (
	"sync"
	"time"
)

// cacheItem contains cace value and date it was accessed last time
type cacheItem struct {
	value      interface{}
	lastAccess int64
}

// cache struct
type Cache struct {
	cache   map[string]cacheItem
	size    int
	maxSize int
	maxTTL  int64
	sync.Mutex
}

// NewCache accepts max cache size and max age(in seconds) of item in cache
//
// It returns new instance of Cache binded on input parameters
func NewCache(maxSize int, maxTTL int) (c *Cache) {
	c = &Cache{
		cache:   make(map[string]cacheItem, maxSize),
		maxSize: maxSize,
		maxTTL:  int64(maxTTL),
	}
	go func() {
		for now := range time.Tick(time.Second) {
			c.Lock()
			for k, v := range c.cache {
				if now.Unix()-v.lastAccess > c.maxTTL {
					delete(c.cache, k)
				}
			}
			c.Unlock()
		}
	}()
	return c
}

// Add adds new item to cache based on key
func (c *Cache) Add(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()

	if c.size > c.maxSize {
		for k, v := range c.cache {
			if time.Now().Unix()-v.lastAccess > c.maxTTL/2 {
				delete(c.cache, k)
			}
		}
	}
	c.cache[key] = cacheItem{
		value:      value,
		lastAccess: time.Now().Unix(),
	}
	c.size++
}

// Delete deletes item which corresponds to key in cache
func (c *Cache) Delete(key string) {
	c.Lock()
	defer c.Unlock()

	delete(c.cache, key)
	c.size--
}

// Get returns item which corresponds to given key
func (c *Cache) Get(key string) (value interface{}, ok bool) {
	c.Lock()
	defer c.Unlock()

	item, ok := c.cache[key]
	if ok {
		c.cache[key] = cacheItem{
			value:      item.value,
			lastAccess: time.Now().Unix(),
		}
	}
	return item.value, ok
}
