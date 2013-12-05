package main

import (
	mcache "github.com/majek/goplayground/cache"
	"time"
)

type MCache struct {
	mcache.LRUCache
	expiry time.Time
	now time.Time
}

func NewMCache(capacity uint64) *MCache {
	return &MCache{
		LRUCache: *mcache.NewLRUCache(capacity),
		expiry: time.Now().Add(time.Duration(30*time.Second)),
		now: time.Now(),
	}
}

func (c *MCache) Get(key string) (string, bool) {
	v, ok := c.LRUCache.Get(key)
	if !ok {
		return "", false
	}
	return v.(*Value).v, true
}

func (c *MCache) Set(key, value string) {
	c.LRUCache.Set(key, &Value{v: value}, time.Time{})
}

