package server

import (
	"errors"
	"fmt"
	"log"
	"north-polaris/internal/dns"
	"sync"
	"time"
)

// record represents a single entry in the cache, containing the cached value and its timestamp.
type record struct {
	tag       string
	record    dns.ResourceRecord
	timestamp time.Time
}

// cache represents the in-memory cache, containing a map of records, a mutex for thread safety, and a cleaner for periodic cleanup.
type Cache struct {
	records map[string]*record
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
		records: make(map[string]*record),
		ttl:     expirationTime,
		cleaner: &cleaner{exitChannel: make(chan int), cleanUpInterval: cleanUpInterval},
	}
	go cache.cleaner.Run(cache)
	log.Printf("active cache --> with expiration time: %s and cleanup interval: %s\n", expirationTime, cleanUpInterval)
	return cache
}

// Put adds a new record to the cache with the specified key and value, along with the current timestamp.
func (c *Cache) Put(key string, value dns.ResourceRecord) {
	log.Printf("cache put: key >> %s\n", key)
	c.mutex.Lock()
	c.records[key] = &record{tag: fmt.Sprintf("%s-%s", value.GetType(), value.Name), record: value, timestamp: time.Now()}
	c.mutex.Unlock()
}

// Get retrieves a record from the cache by its key. If the record is not present, it returns an error.
func (c *Cache) Get(key string) (dns.ResourceRecord, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	result, present := c.records[key]
	if !present {
		log.Printf("cache get: key: %s, value: not found\n", key)
		return dns.ResourceRecord{}, errors.New("cache entry is not present")
	}
	log.Printf("cache get: key [%s]\n", key)
	return result.record, nil
}

// GetAll retrieves all records from the cache that match the specified key. It returns a map of matching records or an error if none are found.
func (c *Cache) GetAll(tag string) ([]dns.ResourceRecord, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	result := make([]dns.ResourceRecord, 0)
	for _, v := range c.records {
		if v.tag == tag {
			result = append(result, v.record)
		}
	}
	if len(result) == 0 {
		log.Printf("cache get all: tag: %s, value: not found\n", tag)
		return nil, errors.New("cache entry is not present")
	}
	log.Printf("cache get all: tag [%s]\n", tag)
	return result, nil
}

// Remove removes a record from the cache by its key.
func (c *Cache) Remove(key string) {
	log.Printf("cache delete: key: %s\n", key)
	c.mutex.Lock()
	delete(c.records, key)
	c.mutex.Unlock()
}

// Clean iterates through the cache records and removes any that have expired based on the current time and the cache's TTL.
func (c *Cache) Clean(currentTime time.Time) {
	log.Printf("running cache cleanup...\n")
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k, v := range c.records {
		if currentTime.After(v.timestamp.Add(c.ttl)) {
			log.Printf("cleaning up %s\n", k)
			delete(c.records, k)
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
