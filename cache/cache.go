package cache

import (
	"time"
)

type Cache interface {
	// functions never using current time
	Get(key string) (value interface{}, ok bool)
	GetQuiet(key string) (value interface{}, ok bool)
	Del(key string) (value interface{}, ok bool)
	Clear() int
	Len() int
	Capacity() int

	// use time.Now() if current time is neccessary to expire entries
	Set(key string, value interface{}, expire time.Time)
	GetNotStale(key string) (value interface{}, ok bool)
	Expire() int

	// manually specify time used when neccessary to expire entries
	SetNow(key string, value interface{}, expire time.Time, now time.Time)
	GetNotStaleNow(key string, now time.Time) (value interface{}, ok bool)
	ExpireNow(now time.Time) int
}
