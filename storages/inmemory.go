// Package storages store rate counters.
package storages

import (
	"sync"
	"time"
)

// NewInMemory is a constructor to InMemory struct.
func NewInMemory() *InMemory {
	inmem := &InMemory{
		items: map[string]*InMemoryItem{},
	}
	inmem.startCleanupTimer()
	return inmem
}

// InMemory is a synchronised map of items that auto-expires.
type InMemory struct {
	mutex sync.RWMutex
	items map[string]*InMemoryItem
}

// IncrBy creates a new item on map or increment existing onr by num.
func (inmem *InMemory) IncrBy(key string, num int64, ttl time.Duration) {
	existing, found := inmem.GetItem(key)
	if found {
		inmem.mutex.Lock()
		existing.IncrBy(num)
		inmem.mutex.Unlock()

	} else {
		inmem.mutex.Lock()
		inmem.items[key] = NewInMemoryItem(num, ttl)
		inmem.mutex.Unlock()
	}
}

// Get a count from map.
func (inmem *InMemory) Get(key string) (count int64, found bool) {
	item, found := inmem.GetItem(key)
	if found {
		return item.count, found
	} else {
		return int64(-1), found
	}
}

// GetItem InMemoryItem struct from map.
func (inmem *InMemory) GetItem(key string) (item *InMemoryItem, found bool) {
	var exists bool

	inmem.mutex.Lock()
	item, exists = inmem.items[key]
	if !exists || item.expired() {
		found = false
	} else {
		item.touch()
		found = true
	}
	inmem.mutex.Unlock()
	return
}

func (inmem *InMemory) cleanup() {
	inmem.mutex.Lock()
	for key, item := range inmem.items {
		if item.expired() {
			delete(inmem.items, key)
		}
	}
	inmem.mutex.Unlock()
}

func (inmem *InMemory) startCleanupTimer() {
	duration := time.Second
	ticker := time.Tick(duration)
	go (func() {
		for {
			select {
			case <-ticker:
				inmem.cleanup()
			}
		}
	})()
}

// NewInMemoryItem is a constructor to InMemoryItem struct.
func NewInMemoryItem(num int64, ttl time.Duration) *InMemoryItem {
	item := &InMemoryItem{count: num, ttl: ttl}
	item.touch()
	return item
}

// InMemoryItem represents a single record.
type InMemoryItem struct {
	sync.RWMutex
	count   int64
	ttl     time.Duration
	expires *time.Time
}

func (item *InMemoryItem) IncrBy(num int64) {
	item.Lock()
	item.count = item.count + num
	item.Unlock()
}

func (item *InMemoryItem) touch() {
	item.Lock()
	expiration := time.Now().Add(item.ttl)
	item.expires = &expiration
	item.Unlock()
}

func (item *InMemoryItem) expired() bool {
	var value bool
	item.RLock()
	if item.expires == nil {
		value = true
	} else {
		value = item.expires.Before(time.Now())
	}
	item.RUnlock()
	return value
}
