// LRU cache data structure
//
// Features:
//
//  - Avoids dynamic memory allocations. All memory is allocated
//    on creation.
//  - Access is O(1). Modification O(n*log(n)).
//  - Multithreading supported using a mutex lock.
//
// Every element in the cache is linked to three data structures:
// `table` map, `priorityQueue` ordered by expiry and `lruList`
// ordered by decreasing popularity.

package cache

import (
	"container/heap"
	"github.com/majek/goplayground/cache/list"
	"sync"
	"time"
)

type entry struct {
	element list.Element // list element. value is a pointer to this entry
	key     string       // key is a key!
	value   interface{}  // value must not be nil
	expire  time.Time    // time when the item is expired. it's okay to be stale.
	index   int          // index for priority queue needs. -1 if entry is free
}

type LRUCache struct {
	lock          sync.Mutex
	table         map[string]*entry
	priorityQueue PriorityQueue
	lruList       list.List // every entry is either used and resides in lruList
	freeList      list.List // or free and is linked to freeList
}

// Create new LRU cache instance. Allocate all the needed memory.
func NewLRUCache(capacity int) *LRUCache {
	b := &LRUCache{
		table:         make(map[string]*entry, capacity),
		priorityQueue: make([]*entry, 0, capacity),
	}
	b.lruList.Init()
	b.freeList.Init()
	heap.Init(&b.priorityQueue)

	// Reserve all the entries in one giant continous block of memory
	arrayOfEntries := make([]entry, capacity)
	for i := 0; i < capacity; i++ {
		e := &arrayOfEntries[i]
		e.element.Value = e
		e.index = -1
		b.freeList.PushElementBack(&e.element)
	}
	return b
}

// Give me the entry with lowest expiry field if it's before now.
func (b *LRUCache) expiredEntry(now time.Time) *entry {
	e := b.priorityQueue[0]
	if e.expire.Before(now) {
		return e
	}
	return nil
}

// Give me the least loved used entry.
func (b *LRUCache) leastUsedEntry() *entry {
	return b.lruList.Back().Value.(*entry)
}

func (b *LRUCache) freeSomeEntry(now time.Time) (e *entry, used bool) {
	if b.freeList.Len() > 0 {
		return b.freeList.PopFront().(*entry), false
	}

	e = b.expiredEntry(now)
	if e != nil {
		return e, true
	}
	return b.leastUsedEntry(), true
}

// Move entry from used/lru list to a free list. Clear the entry as well.
func (b *LRUCache) removeEntry(e *entry) {
	heap.Remove(&b.priorityQueue, e.index)
	b.lruList.Remove(&e.element)
	b.freeList.PushElementFront(&e.element)
	delete(b.table, e.key)
	e.key = ""
	e.value = nil
}

func (b *LRUCache) insertEntry(e *entry) {
	heap.Push(&b.priorityQueue, e)
	b.freeList.Remove(&e.element)
	b.lruList.PushElementFront(&e.element)
	b.table[e.key] = e
}

func (b *LRUCache) touchEntry(e *entry) {
	b.lruList.Remove(&e.element)
	b.lruList.PushElementFront(&e.element)
}

// Add an item to the cache overwriting existing one if it
// exists. Allows specifing current time required to expire an
// item when no more slots are used. Value must not be
// nil. O(log(n))
func (b *LRUCache) SetNow(key string, value interface{}, expire time.Time, now *time.Time) {
	if value == nil {
		panic("Value must not be nil!")
	}

	b.lock.Lock()
	defer b.lock.Unlock()

	var used bool

	e := b.table[key]
	if e != nil {
		used = true
	} else {
		var xnow time.Time
		if now == nil {
			xnow = time.Now()
		} else {
			xnow = *now
		}
		e, used = b.freeSomeEntry(xnow)
	}
	if used {
		b.removeEntry(e)
	}

	e.key = key
	e.value = value
	e.expire = expire
	b.insertEntry(e)
}

// Add an item to the cache overwriting existing one if it
// exists. O(log(n))
func (b *LRUCache) Set(key string, value interface{}, expire time.Time) {
	b.SetNow(key, value, expire, nil)
}

// Get a key from the cache, possibly stale. Update its LRU score. O(1)
func (b *LRUCache) Get(key string) interface{} {
	b.lock.Lock()
	defer b.lock.Unlock()

	e := b.table[key]
	if e == nil {
		return nil
	}

	b.touchEntry(e)
	return e.value
}

// Get a key from the cache, possibly stale. Don't modify its LRU score. O(1)
func (b *LRUCache) GetQuiet(key string) interface{} {
	b.lock.Lock()
	defer b.lock.Unlock()

	e := b.table[key]
	if e == nil {
		return nil
	}

	return e.value
}

// Get a key from the cache, make sure it's not stale. Update its
// LRU score. O(n*log(n))
func (b *LRUCache) GetNotStale(key string) interface{} {
	return b.GetNotStaleNow(key, time.Now())
}

func (b *LRUCache) GetNotStaleNow(key string, now time.Time) interface{} {
	b.lock.Lock()
	defer b.lock.Unlock()

	e := b.table[key]
	if e == nil {
		return nil
	}

	if e.expire.Before(now) {
		b.removeEntry(e)
		return nil
	}

	b.touchEntry(e)
	return e.value
}

// Get and remove a key from the cache. O(log(n))
func (b *LRUCache) Del(key string) interface{} {
	b.lock.Lock()
	defer b.lock.Unlock()

	e := b.table[key]
	if e == nil {
		return nil
	}

	value := e.value
	b.removeEntry(e)
	return value
}

// Evict all items from the cache. O(n*log(n))
func (b *LRUCache) Clear() int {
	b.lock.Lock()
	defer b.lock.Unlock()

	l := b.lruList.Len()
	// This could be reduced to O(n).
	for i := 0; i < l; i++ {
		b.removeEntry(b.priorityQueue[0])
	}
	return l
}

// Evict all the expired items. O(n*log(n))
func (b *LRUCache) Expire() int {
	return b.ExpireNow(time.Now())
}

// Evict items that expire before `now`. O(n*log(n))
func (b *LRUCache) ExpireNow(now time.Time) int {
	b.lock.Lock()
	defer b.lock.Unlock()

	i := 0
	for {
		e := b.expiredEntry(now)
		if e == nil {
			break
		}
		b.removeEntry(e)
		i += 1
	}
	return i
}

// Evict all the expired items. O(n*log(n))
func (b *LRUCache) Len() int {
	return b.lruList.Len()
}
