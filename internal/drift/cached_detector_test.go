package drift

import (
	"testing"
	"time"
)

func TestNewCachedDetector_NotNil(t *testing.T) {
	cd := NewCachedDetector(NewDetector(), 1*time.Minute)
	if cd == nil {
		t.Fatal("expected non-nil CachedDetector")
	}
}

func TestCachedDetector_Compare_CachesMiss(t *testing.T) {
	cd := NewCachedDetector(NewDetector(), 1*time.Minute)

	spec := baseService()
	observed := baseService()

	_ = cd.Compare(spec, observed)
	if cd.CacheSize() != 1 {
		t.Errorf("expected cache size 1 after first compare, got %d", cd.CacheSize())
	}
}

func TestCachedDetector_Compare_ReturnsFromCache(t *testing.T) {
	cd := NewCachedDetector(NewDetector(), 1*time.Minute)

	spec := baseService()
	observed := baseService()

	r1 := cd.Compare(spec, observed)
	r2 := cd.Compare(spec, observed)

	if r1.Service != r2.Service {
		t.Errorf("expected same result from cache, got %v vs %v", r1, r2)
	}
	// Size should still be 1 (no duplicate entries)
	if cd.CacheSize() != 1 {
		t.Errorf("expected cache size 1, got %d", cd.CacheSize())
	}
}

func TestCachedDetector_Invalidate(t *testing.T) {
	cd := NewCachedDetector(NewDetector(), 1*time.Minute)

	spec := baseService()
	observed := baseService()
	_ = cd.Compare(spec, observed)

	cd.Invalidate(spec.Name)
	if cd.CacheSize() != 0 {
		t.Errorf("expected cache size 0 after invalidation, got %d", cd.CacheSize())
	}
}

func TestCachedDetector_Flush(t *testing.T) {
	cd := NewCachedDetector(NewDetector(), 1*time.Minute)

	s1 := baseService()
	s2 := baseService()
	s2.Name = "other-service"

	_ = cd.Compare(s1, s1)
	_ = cd.Compare(s2, s2)

	cd.Flush()
	if cd.CacheSize() != 0 {
		t.Errorf("expected empty cache after flush, got %d", cd.CacheSize())
	}
}

func TestCachedDetector_Compare_ExpiredEntryRecomputes(t *testing.T) {
	cd := NewCachedDetector(NewDetector(), 1*time.Millisecond)

	spec := baseService()
	observed := baseService()
	_ = cd.Compare(spec, observed)

	time.Sleep(5 * time.Millisecond)

	// After TTL, the cache should miss and recompute
	_ = cd.Compare(spec, observed)
	if cd.CacheSize() != 1 {
		t.Errorf("expected cache size 1 after recompute, got %d", cd.CacheSize())
	}
}
