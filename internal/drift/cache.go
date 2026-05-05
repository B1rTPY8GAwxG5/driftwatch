package drift

import (
	"sync"
	"time"
)

// CacheEntry holds a cached DriftResult with a timestamp.
type CacheEntry struct {
	Result    DriftResult
	CachedAt  time.Time
}

// Cache stores recent drift results keyed by service name.
type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
	ttl     time.Duration
}

// NewCache creates a Cache with the given TTL duration.
func NewCache(ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	return &Cache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
	}
}

// Set stores a DriftResult for the given service name.
func (c *Cache) Set(service string, result DriftResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[service] = CacheEntry{
		Result:   result,
		CachedAt: time.Now(),
	}
}

// Get retrieves a cached DriftResult. Returns the entry and whether it was
// found and still valid (within TTL).
func (c *Cache) Get(service string) (CacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[service]
	if !ok {
		return CacheEntry{}, false
	}
	if time.Since(entry.CachedAt) > c.ttl {
		return CacheEntry{}, false
	}
	return entry, true
}

// Invalidate removes the cached entry for the given service.
func (c *Cache) Invalidate(service string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, service)
}

// Flush removes all cached entries.
func (c *Cache) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]CacheEntry)
}

// Size returns the number of entries currently in the cache.
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}
