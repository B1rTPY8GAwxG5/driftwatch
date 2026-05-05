package drift

import (
	"testing"
	"time"
)

func cachedResult(service string, drifted bool) DriftResult {
	entries := []DriftEntry{}
	if drifted {
		entries = append(entries, DriftEntry{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"})
	}
	return DriftResult{Service: service, Entries: entries}
}

func TestNewCache_DefaultTTL(t *testing.T) {
	c := NewCache(0)
	if c.ttl != 5*time.Minute {
		t.Errorf("expected default TTL 5m, got %v", c.ttl)
	}
}

func TestCache_SetAndGet_Valid(t *testing.T) {
	c := NewCache(1 * time.Minute)
	r := cachedResult("svc-a", false)
	c.Set("svc-a", r)

	entry, ok := c.Get("svc-a")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if entry.Result.Service != "svc-a" {
		t.Errorf("unexpected service: %s", entry.Result.Service)
	}
}

func TestCache_Get_Expired(t *testing.T) {
	c := NewCache(1 * time.Millisecond)
	c.Set("svc-b", cachedResult("svc-b", true))
	time.Sleep(5 * time.Millisecond)

	_, ok := c.Get("svc-b")
	if ok {
		t.Error("expected cache miss for expired entry")
	}
}

func TestCache_Get_Missing(t *testing.T) {
	c := NewCache(1 * time.Minute)
	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected cache miss for unknown service")
	}
}

func TestCache_Invalidate(t *testing.T) {
	c := NewCache(1 * time.Minute)
	c.Set("svc-c", cachedResult("svc-c", false))
	c.Invalidate("svc-c")

	_, ok := c.Get("svc-c")
	if ok {
		t.Error("expected cache miss after invalidation")
	}
}

func TestCache_Flush(t *testing.T) {
	c := NewCache(1 * time.Minute)
	c.Set("svc-d", cachedResult("svc-d", false))
	c.Set("svc-e", cachedResult("svc-e", true))
	c.Flush()

	if c.Size() != 0 {
		t.Errorf("expected empty cache after flush, got size %d", c.Size())
	}
}

func TestCache_Size(t *testing.T) {
	c := NewCache(1 * time.Minute)
	if c.Size() != 0 {
		t.Errorf("expected size 0, got %d", c.Size())
	}
	c.Set("svc-f", cachedResult("svc-f", false))
	if c.Size() != 1 {
		t.Errorf("expected size 1, got %d", c.Size())
	}
}
