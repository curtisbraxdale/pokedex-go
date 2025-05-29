package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	cacheMap map[string]cacheEntry
	mu       sync.Mutex
}

func NewCache(interval time.Duration) *Cache {
	newCache := Cache{cacheMap: make(map[string]cacheEntry), mu: sync.Mutex{}}
	go newCache.reapLoop(interval)
	return &newCache
}

func (c *Cache) Add(key string, val []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	entry := cacheEntry{val: val, createdAt: time.Now()}
	c.cacheMap[key] = entry
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	val, exists := c.cacheMap[key]
	if exists {
		return val.val, true
	} else {
		return nil, false
	}
}

func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.Tick(interval)
	for range ticker {
		c.mu.Lock()
		for key, value := range c.cacheMap {
			if value.createdAt.Before(time.Now().Add(-interval)) {
				delete(c.cacheMap, key)
			}
		}
		c.mu.Unlock()
	}
}
