package drift

import (
	"sync"
	"time"
)

// BatchPolicy controls how results are collected before flushing.
type BatchPolicy struct {
	MaxSize  int
	MaxWait  time.Duration
}

// DefaultBatchPolicy returns sensible defaults for batching.
func DefaultBatchPolicy() BatchPolicy {
	return BatchPolicy{
		MaxSize: 20,
		MaxWait: 5 * time.Second,
	}
}

// Batcher accumulates DriftResults and flushes them when a size or
// time threshold is reached.
type Batcher struct {
	mu      sync.Mutex
	policy  BatchPolicy
	buf     []DriftResult
	flushFn func([]DriftResult)
	stopCh  chan struct{}
}

// NewBatcher creates a Batcher with the given policy and flush callback.
// If policy is zero-valued, DefaultBatchPolicy is used.
func NewBatcher(policy BatchPolicy, flushFn func([]DriftResult)) *Batcher {
	if policy.MaxSize <= 0 || policy.MaxWait <= 0 {
		policy = DefaultBatchPolicy()
	}
	if flushFn == nil {
		flushFn = func([]DriftResult) {}
	}
	b := &Batcher{
		policy:  policy,
		flushFn: flushFn,
		stopCh:  make(chan struct{}),
	}
	go b.ticker()
	return b
}

// Add appends a result to the current batch, flushing if MaxSize is reached.
func (b *Batcher) Add(r DriftResult) {
	b.mu.Lock()
	b.buf = append(b.buf, r)
	should := len(b.buf) >= b.policy.MaxSize
	var batch []DriftResult
	if should {
		batch = b.drain()
	}
	b.mu.Unlock()
	if batch != nil {
		b.flushFn(batch)
	}
}

// Flush forces an immediate flush of buffered results.
func (b *Batcher) Flush() {
	b.mu.Lock()
	batch := b.drain()
	b.mu.Unlock()
	if len(batch) > 0 {
		b.flushFn(batch)
	}
}

// Stop halts the background ticker and flushes remaining results.
func (b *Batcher) Stop() {
	close(b.stopCh)
	b.Flush()
}

// Len returns the number of buffered results.
func (b *Batcher) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.buf)
}

func (b *Batcher) drain() []DriftResult {
	if len(b.buf) == 0 {
		return nil
	}
	out := make([]DriftResult, len(b.buf))
	copy(out, b.buf)
	b.buf = b.buf[:0]
	return out
}

func (b *Batcher) ticker() {
	t := time.NewTicker(b.policy.MaxWait)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			b.Flush()
		case <-b.stopCh:
			return
		}
	}
}
