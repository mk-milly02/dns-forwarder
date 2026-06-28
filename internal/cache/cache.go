package cache

import (
	"errors"
	"log"
	"sync"
	"time"
)

// cEntry represents a single entry in the cache, containing the cached value and its timestamp.
type cEntry struct {
	entry     any
	timestamp time.Time
}

// cache represents the in-memory cache, containing a map of entries, a mutex for thread safety, and a cleaner for periodic cleanup.
type Cache struct {
	entries map[string]*cEntry
	cleaner *cleaner
	mutex   sync.Mutex
	ttl     time.Duration
}

// cleaner is responsible for periodically cleaning up expired entries from the cache.
type cleaner struct {
	exitChannel     chan int
	cleanUpInterval time.Duration
}

// NewCache creates a new cache instance with the specified expiration time and cleanup interval, and starts the cleaner goroutine.
func NewCache(expirationTime time.Duration, cleanUpInterval time.Duration) *Cache {
	cache := &Cache{
		entries: make(map[string]*cEntry),
		ttl:     expirationTime,
		cleaner: &cleaner{exitChannel: make(chan int), cleanUpInterval: cleanUpInterval},
	}
	go cache.cleaner.Run(cache)
	log.Printf("active cache -> with expiration time: %s and cleanup interval: %s\n", expirationTime, cleanUpInterval)
	return cache
}

// Put adds a new entry to the cache with the specified key and value, along with the current timestamp.
func (cache *Cache) Put(key string, value any) {
	log.Printf("cache put: key: %s, value: %+v\n", key, value)
	cache.mutex.Lock()
	cache.entries[key] = &cEntry{entry: value, timestamp: time.Now()}
	cache.mutex.Unlock()
}

// Get retrieves an entry from the cache by its key. If the entry is not present, it returns an error.
func (cache *Cache) Get(key string) (any, error) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	result, present := cache.entries[key]
	if !present {
		return nil, errors.New("cache entry is not present")
	}
	return result.entry, nil
}

// Remove removes an entry from the cache by its key.
func (cache *Cache) Remove(key string) {
	log.Printf("cache delete: key: %s\n", key)
	cache.mutex.Lock()
	delete(cache.entries, key)
	cache.mutex.Unlock()
}

// Clean iterates through the cache entries and removes any that have expired based on the current time and the cache's TTL.
func (cache *Cache) Clean(currentTime time.Time) {
	log.Println("running cache cleanup...")
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	for k, v := range cache.entries {
		if currentTime.After(v.timestamp.Add(cache.ttl)) {
			log.Printf("cleaning up %s\n", k)
			delete(cache.entries, k)
		}
	}
}

// Run starts the cleaner goroutine, which periodically checks for expired cache entries and removes them.
func (cleaner *cleaner) Run(c *Cache) {
    ticker := time.NewTicker(cleaner.cleanUpInterval)
    for {
        select {
        case currentTime := <-ticker.C:
            c.Clean(currentTime)
        case <-cleaner.exitChannel:
            return
        }
    }
}