package drift

import "time"

// CachedDetector wraps a Detector with a Cache to avoid redundant comparisons.
type CachedDetector struct {
	detector *Detector
	cache    *Cache
}

// NewCachedDetector creates a CachedDetector with the given TTL.
func NewCachedDetector(d *Detector, ttl time.Duration) *CachedDetector {
	return &CachedDetector{
		detector: d,
		cache:    NewCache(ttl),
	}
}

// Compare returns a cached DriftResult if available and valid, otherwise
// delegates to the underlying Detector and caches the result.
func (cd *CachedDetector) Compare(spec ServiceSpec, observed ServiceSpec) DriftResult {
	key := spec.Name
	if entry, ok := cd.cache.Get(key); ok {
		return entry.Result
	}
	result := cd.detector.Compare(spec, observed)
	cd.cache.Set(key, result)
	return result
}

// Invalidate removes the cached result for the given service name.
func (cd *CachedDetector) Invalidate(service string) {
	cd.cache.Invalidate(service)
}

// Flush clears all cached results.
func (cd *CachedDetector) Flush() {
	cd.cache.Flush()
}

// CacheSize returns the number of cached entries.
func (cd *CachedDetector) CacheSize() int {
	return cd.cache.Size()
}
