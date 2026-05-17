package drift

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// FlushPolicy controls when and how buffered results are flushed.
type FlushPolicy struct {
	MaxSize  int           // flush when buffer reaches this size
	MaxAge   time.Duration // flush when oldest entry exceeds this age
}

// DefaultFlushPolicy returns sensible flush defaults.
func DefaultFlushPolicy() FlushPolicy {
	return FlushPolicy{
		MaxSize: 50,
		MaxAge:  5 * time.Minute,
	}
}

// Flusher buffers DriftResults and flushes them to a writer when a policy
// threshold is met or Flush is called explicitly.
type Flusher struct {
	policy  FlushPolicy
	buf     []DriftResult
	oldest  time.Time
	writer  io.Writer
}

// NewFlusher creates a Flusher with the given policy and destination writer.
// A zero-value policy falls back to DefaultFlushPolicy.
func NewFlusher(policy FlushPolicy, w io.Writer) *Flusher {
	if policy.MaxSize <= 0 {
		policy = DefaultFlushPolicy()
	}
	if policy.MaxAge <= 0 {
		policy = DefaultFlushPolicy()
	}
	return &Flusher{
		policy: policy,
		writer: w,
	}
}

// Add appends a result to the buffer, flushing automatically if a threshold
// is reached.
func (f *Flusher) Add(r DriftResult) error {
	if len(f.buf) == 0 {
		f.oldest = time.Now()
	}
	f.buf = append(f.buf, r)
	if f.shouldFlush() {
		return f.Flush()
	}
	return nil
}

// Len returns the number of buffered results.
func (f *Flusher) Len() int { return len(f.buf) }

// Flush writes all buffered results to the writer and clears the buffer.
func (f *Flusher) Flush() error {
	if len(f.buf) == 0 {
		return nil
	}
	sort.Slice(f.buf, func(i, j int) bool {
		return f.buf[i].Service < f.buf[j].Service
	})
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("--- flush: %d result(s) ---\n", len(f.buf)))
	for _, r := range f.buf {
		drifted := "clean"
		if r.HasDrift() {
			drifted = "drifted"
		}
		sb.WriteString(fmt.Sprintf("  service=%s status=%s entries=%d\n",
			r.Service, drifted, len(r.Entries)))
	}
	_, err := fmt.Fprint(f.writer, sb.String())
	f.buf = f.buf[:0]
	f.oldest = time.Time{}
	return err
}

func (f *Flusher) shouldFlush() bool {
	if len(f.buf) >= f.policy.MaxSize {
		return true
	}
	if !f.oldest.IsZero() && time.Since(f.oldest) >= f.policy.MaxAge {
		return true
	}
	return false
}
