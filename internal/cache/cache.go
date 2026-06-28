package cache

import (
	"errors"
	"fmt"
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
type cache struct {
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

// cacheService provides a higher-level interface for interacting with the cache, allowing for resource-specific caching.
type cacheService struct {
	cache *cache
}

// new creates a new cache instance with the specified expiration time and cleanup interval, and starts the cleaner goroutine.
func new(expirationTime time.Duration, cleanUpInterval time.Duration) *cache {
	cache := &cache{
		entries: make(map[string]*cEntry),
		ttl:     expirationTime,
		cleaner: &cleaner{exitChannel: make(chan int), cleanUpInterval: cleanUpInterval},
	}
	go cache.cleaner.Run(cache)
	return cache
}

// put adds a new entry to the cache with the specified key and value, along with the current timestamp.
func (cache *cache) put(key string, value any) {
	log.Printf("cache put: key: %s, value: %+v\n", key, value)
	cache.mutex.Lock()
	cache.entries[key] = &cEntry{entry: value, timestamp: time.Now()}
	cache.mutex.Unlock()
}

// get retrieves an entry from the cache by its key. If the entry is not present, it returns an error.
func (cache *cache) get(key string) (any, error) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	result, present := cache.entries[key]
	if !present {
		return nil, errors.New("cache entry is not present")
	}
	return result.entry, nil
}

// delete removes an entry from the cache by its key.
func (cache *cache) delete(key string) {
	log.Printf("cache delete: key: %s\n", key)
	cache.mutex.Lock()
	delete(cache.entries, key)
	cache.mutex.Unlock()
}

// clean iterates through the cache entries and removes any that have expired based on the current time and the cache's TTL.
func (cache *cache) clean(currentTime time.Time) {
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

// CacheResource checks if a resource is present in the cache. If it is, it returns the cached value. 
// If not, it calls the provided function to retrieve the resource, caches it, and then returns the value.
func (cs *cacheService) CacheResource(f func() (any, error), resource string, key string) (any, error) {
   fullKey := fmt.Sprintf("%s-%s", resource, key)
   cached, cacheErr := cs.cache.get(fullKey)
   if cacheErr == nil {
    log.Printf("cache hit: key: %s, value: %+v\n", fullKey, cached)
    return cached, nil
   }
   fnRes, fnErr := f()
   if fnErr == nil {
    cs.cache.put(fullKey, fnRes)
   }
   return fnRes, fnErr
}

// Run starts the cleaner goroutine, which periodically checks for expired cache entries and removes them.
func (cleaner *cleaner) Run(c *cache) {
    ticker := time.NewTicker(cleaner.cleanUpInterval)
    for {
        select {
        case currentTime := <-ticker.C:
            c.clean(currentTime)
        case <-cleaner.exitChannel:
            return
        }
    }
}