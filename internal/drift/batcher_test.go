package drift

import (
	"sync"
	"testing"
	"time"
)

func TestDefaultBatchPolicy_Values(t *testing.T) {
	p := DefaultBatchPolicy()
	if p.MaxSize != 20 {
		t.Errorf("expected MaxSize 20, got %d", p.MaxSize)
	}
	if p.MaxWait != 5*time.Second {
		t.Errorf("expected MaxWait 5s, got %v", p.MaxWait)
	}
}

func TestNewBatcher_NotNil(t *testing.T) {
	b := NewBatcher(DefaultBatchPolicy(), nil)
	defer b.Stop()
	if b == nil {
		t.Fatal("expected non-nil Batcher")
	}
}

func TestNewBatcher_ZeroPolicy_UsesDefaults(t *testing.T) {
	b := NewBatcher(BatchPolicy{}, nil)
	defer b.Stop()
	if b.policy.MaxSize != 20 {
		t.Errorf("expected default MaxSize 20, got %d", b.policy.MaxSize)
	}
}

func TestBatcher_Len_InitiallyZero(t *testing.T) {
	b := NewBatcher(DefaultBatchPolicy(), nil)
	defer b.Stop()
	if b.Len() != 0 {
		t.Errorf("expected 0, got %d", b.Len())
	}
}

func TestBatcher_Add_IncrementsLen(t *testing.T) {
	b := NewBatcher(BatchPolicy{MaxSize: 10, MaxWait: time.Minute}, nil)
	defer b.Stop()
	b.Add(DriftResult{Service: "svc-a"})
	b.Add(DriftResult{Service: "svc-b"})
	if b.Len() != 2 {
		t.Errorf("expected 2, got %d", b.Len())
	}
}

func TestBatcher_Flush_DrainsBuf(t *testing.T) {
	var mu sync.Mutex
	var flushed []DriftResult
	b := NewBatcher(BatchPolicy{MaxSize: 10, MaxWait: time.Minute}, func(batch []DriftResult) {
		mu.Lock()
		flushed = append(flushed, batch...)
		mu.Unlock()
	})
	defer b.Stop()
	b.Add(DriftResult{Service: "svc-a"})
	b.Flush()
	mu.Lock()
	defer mu.Unlock()
	if len(flushed) != 1 {
		t.Errorf("expected 1 flushed, got %d", len(flushed))
	}
	if b.Len() != 0 {
		t.Errorf("expected buf empty after flush")
	}
}

func TestBatcher_Add_FlushesAtMaxSize(t *testing.T) {
	var mu sync.Mutex
	var flushed []DriftResult
	b := NewBatcher(BatchPolicy{MaxSize: 3, MaxWait: time.Minute}, func(batch []DriftResult) {
		mu.Lock()
		flushed = append(flushed, batch...)
		mu.Unlock()
	})
	defer b.Stop()
	for i := 0; i < 3; i++ {
		b.Add(DriftResult{Service: "svc"})
	}
	time.Sleep(20 * time.Millisecond)
	mu.Lock()
	defer mu.Unlock()
	if len(flushed) != 3 {
		t.Errorf("expected 3 flushed at MaxSize, got %d", len(flushed))
	}
}

func TestBatcher_Stop_FlushesRemaining(t *testing.T) {
	var mu sync.Mutex
	var flushed []DriftResult
	b := NewBatcher(BatchPolicy{MaxSize: 10, MaxWait: time.Minute}, func(batch []DriftResult) {
		mu.Lock()
		flushed = append(flushed, batch...)
		mu.Unlock()
	})
	b.Add(DriftResult{Service: "svc-x"})
	b.Stop()
	mu.Lock()
	defer mu.Unlock()
	if len(flushed) != 1 {
		t.Errorf("expected 1 flushed on Stop, got %d", len(flushed))
	}
}
