package cache

import (
	"errors"
	"fmt"
	"log"
	"north-polaris/internal/forwarder"
	"sync"
	"time"
)

// cRecord represents a single entry in the cache, containing the cached value and its timestamp.
type cRecord struct {
	tag       string
	record    forwarder.ResourceRecord
	timestamp time.Time
}

// cache represents the in-memory cache, containing a map of records, a mutex for thread safety, and a cleaner for periodic cleanup.
type Cache struct {
	records map[string]*cRecord
	cleaner *cleaner
	mutex   sync.Mutex
	ttl     time.Duration
}

// cleaner is responsible for periodically cleaning up expired records from the cache.
type cleaner struct {
	exitChannel     chan int
	cleanUpInterval time.Duration
}

// NewCache creates a new cache instance with the specified expiration time and cleanup interval, and starts the cleaner goroutine.
func NewCache(expirationTime time.Duration, cleanUpInterval time.Duration) *Cache {
	cache := &Cache{
		records: make(map[string]*cRecord),
		ttl:     expirationTime,
		cleaner: &cleaner{exitChannel: make(chan int), cleanUpInterval: cleanUpInterval},
	}
	go cache.cleaner.Run(cache)
	log.Printf("active cache --> with expiration time: %s and cleanup interval: %s\n\n", expirationTime, cleanUpInterval)
	return cache
}

// Put adds a new record to the cache with the specified key and value, along with the current timestamp.
func (cache *Cache) Put(key string, value forwarder.ResourceRecord) {
	log.Printf("cache put: key: %s\n\n", key)
	cache.mutex.Lock()
	cache.records[key] = &cRecord{tag: fmt.Sprintf("%s-%s", value.GetType(), value.Name), record: value, timestamp: time.Now()}
	cache.mutex.Unlock()
}

// Get retrieves a record from the cache by its key. If the record is not present, it returns an error.
func (cache *Cache) Get(key string) (forwarder.ResourceRecord, error) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	result, present := cache.records[key]
	if !present {
		log.Printf("cache get: key: %s, value: not found\n\n", key)
		return forwarder.ResourceRecord{}, errors.New("cache entry is not present")
	}
	log.Printf("cache get: key [%s]\n\n", key)
	return result.record, nil
}

// GetAll retrieves all records from the cache that match the specified key. It returns a map of matching records or an error if none are found.
func (cache *Cache) GetAll(tag string) ([]forwarder.ResourceRecord, error) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	result := make([]forwarder.ResourceRecord, 0)
	for _, v := range cache.records {
		if v.tag == tag {
			result = append(result, v.record)
		}
	}
	if len(result) == 0 {
		log.Printf("cache get all: tag: %s, value: not found\n\n", tag)
		return nil, errors.New("cache entry is not present")
	}
	log.Printf("cache get all: tag [%s]\n\n", tag)
	return result, nil
}

// Remove removes a record from the cache by its key.
func (cache *Cache) Remove(key string) {
	log.Printf("cache delete: key: %s\n\n", key)
	cache.mutex.Lock()
	delete(cache.records, key)
	cache.mutex.Unlock()
}

// Clean iterates through the cache records and removes any that have expired based on the current time and the cache's TTL.
func (cache *Cache) Clean(currentTime time.Time) {
	log.Printf("running cache cleanup...\n\n")
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	for k, v := range cache.records {
		if currentTime.After(v.timestamp.Add(cache.ttl)) {
			log.Printf("cleaning up %s\n\n", k)
			delete(cache.records, k)
		}
	}
}

// Run starts the cleaner goroutine, which periodically checks for expired cache records and removes them.
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
