package ycache

import (
	"7days/ycache/lru"
	"sync"
)

type cache struct {
	mu         sync.Mutex
	lruk       *lru.LRUKCache
	cacheBytes int
}

func (c *cache) add(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lruk == nil {
		c.lruk = lru.NewLRUKCache(c.cacheBytes, 2)
	}

	c.lruk.Add(key, value)
}

func (c *cache) get(key string) (value ByteView, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.lruk == nil {
		return
	}

	if v, ok := c.lruk.Get(key); ok {
		return v.(ByteView), ok
	}

	return
}
