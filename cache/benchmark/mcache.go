package main

import (
	mcache "github.com/majek/goplayground/cache"
	"github.com/majek/goplayground/cache/lrucache"
	"github.com/majek/goplayground/cache/multilru"
	"time"
)

func makeLRUCache(capacity uint64) mcache.Cache {
	return lrucache.NewLRUCache(uint(capacity))
}
func makeMultiLRU(capacity uint64) mcache.Cache {
	return multilru.NewMultiLRUCache(4, uint(capacity/4), func(capacity uint) mcache.Cache {
		return lrucache.NewLRUCache(capacity)
	})
}

type MCache struct {
	mcache.Cache
	expiry time.Time
	now time.Time
}

type makeCache func(capacity uint64) mcache.Cache

func NewMCache(capacity uint64, newCache makeCache) *MCache {
	return &MCache{
		Cache: newCache(capacity),
		expiry: time.Now().Add(time.Duration(30*time.Second)),
		now: time.Now(),
	}
}

func (c *MCache) Get(key string) (string, bool) {
	v, ok := c.Cache.Get(key)
	if !ok {
		return "", false
	}
	return v.(*Value).v, true
}

func (c *MCache) Set(key, value string) {
	c.Cache.Set(key, &Value{v: value}, time.Time{})
}

