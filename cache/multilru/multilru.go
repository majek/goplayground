package multilru

import (
	"github.com/majek/goplayground/cache"
	"hash"
	"hash/fnv"
	"time"
)

type MultiLRUCache struct {
	buckets uint
	cache   []cache.Cache
	hash    hash.Hash
}

type MakeCache func(capacity uint) cache.Cache

func (m *MultiLRUCache) Init(buckets, bucket_capacity uint, newCache MakeCache) {
	mlru := &MultiLRUCache{
		buckets: buckets,
		cache:   make([]cache.Cache, buckets),
	}
	for i := uint(0); i < buckets; i++ {
		mlru.cache[i] = newCache(bucket_capacity)
	}
}

func NewMultiLRUCache(buckets, bucket_capacity uint, newCache MakeCache) *MultiLRUCache {
	mlru := &MultiLRUCache{}
	mlru.Init(buckets, bucket_capacity, newCache)
	return mlru
}

func (m *MultiLRUCache) bucketNo(key string) uint {
	h := fnv.New32a() // Arbitrary choice. Any fast hash will do.
	h.Write([]byte(key))
	return uint(h.Sum32()) % m.buckets
}

func (m *MultiLRUCache) Set(key string, value interface{}, expire time.Time) {
	m.cache[m.bucketNo(key)].Set(key, value, expire)
}

func (m *MultiLRUCache) SetNow(key string, value interface{}, expire time.Time, now time.Time) {
	m.cache[m.bucketNo(key)].SetNow(key, value, expire, now)
}

func (m *MultiLRUCache) Get(key string) (value interface{}, ok bool) {
	return m.cache[m.bucketNo(key)].Get(key)
}

func (m *MultiLRUCache) GetQuiet(key string) (value interface{}, ok bool) {
	return m.cache[m.bucketNo(key)].Get(key)
}

func (m *MultiLRUCache) GetNotStale(key string) (value interface{}, ok bool) {
	return m.cache[m.bucketNo(key)].GetNotStale(key)
}

func (m *MultiLRUCache) GetNotStaleNow(key string, now time.Time) (value interface{}, ok bool) {
	return m.cache[m.bucketNo(key)].GetNotStaleNow(key, now)
}

func (m *MultiLRUCache) Del(key string) (value interface{}, ok bool) {
	return m.cache[m.bucketNo(key)].Del(key)
}

func (m *MultiLRUCache) Clear() int {
	var s int
	for _, c := range m.cache {
		s += c.Clear()
	}
	return s
}

func (m *MultiLRUCache) Len() int {
	var s int
	for _, c := range m.cache {
		s += c.Len()
	}
	return s
}

func (m *MultiLRUCache) Capacity() int {
	var s int
	for _, c := range m.cache {
		s += c.Capacity()
	}
	return s
}

func (m *MultiLRUCache) Expire() int {
	var s int
	for _, c := range m.cache {
		s += c.Expire()
	}
	return s
}
