package drift

import (
	"sync"
	"time"
)

// Debouncer suppresses rapid repeated drift results for the same service,
// only forwarding a result once the signal has been stable for the configured
// quiet period.
type Debouncer struct {
	mu      sync.Mutex
	quiet   time.Duration
	timers  map[string]*time.Timer
	pending map[string]DriftResult
	forward func(DriftResult)
}

// NewDebouncer creates a Debouncer that waits for quiet before forwarding.
// forward is called at most once per quiet period per service.
func NewDebouncer(quiet time.Duration, forward func(DriftResult)) *Debouncer {
	if quiet <= 0 {
		quiet = 5 * time.Second
	}
	return &Debouncer{
		quiet:   quiet,
		timers:  make(map[string]*time.Timer),
		pending: make(map[string]DriftResult),
		forward: forward,
	}
}

// Submit schedules result for forwarding after the quiet period. If another
// result for the same service arrives before the timer fires, the timer is
// reset and the newer result replaces the pending one.
func (d *Debouncer) Submit(result DriftResult) {
	d.mu.Lock()
	defer d.mu.Unlock()

	key := result.Service
	d.pending[key] = result

	if t, ok := d.timers[key]; ok {
		t.Reset(d.quiet)
		return
	}

	d.timers[key] = time.AfterFunc(d.quiet, func() {
		d.mu.Lock()
		r := d.pending[key]
		delete(d.timers, key)
		delete(d.pending, key)
		d.mu.Unlock()
		d.forward(r)
	})
}

// Flush immediately forwards all pending results and cancels outstanding timers.
func (d *Debouncer) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	for key, t := range d.timers {
		t.Stop()
		delete(d.timers, key)
		d.forward(d.pending[key])
		delete(d.pending, key)
	}
}

// Pending returns the number of results currently waiting to be forwarded.
func (d *Debouncer) Pending() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.pending)
}
