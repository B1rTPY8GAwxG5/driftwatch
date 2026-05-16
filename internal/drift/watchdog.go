package drift

import (
	"fmt"
	"io"
	"sync"
	"time"
)

// WatchdogPolicy controls when the watchdog fires.
type WatchdogPolicy struct {
	// MaxSilence is the maximum duration without a successful check before
	// the watchdog considers the detector stalled.
	MaxSilence time.Duration
	// CheckInterval is how often the watchdog inspects the last-seen time.
	CheckInterval time.Duration
}

// DefaultWatchdogPolicy returns a sensible WatchdogPolicy.
func DefaultWatchdogPolicy() WatchdogPolicy {
	return WatchdogPolicy{
		MaxSilence:    5 * time.Minute,
		CheckInterval: 30 * time.Second,
	}
}

// Watchdog monitors detector liveness and writes an alert when the detector
// has not reported within MaxSilence.
type Watchdog struct {
	mu       sync.Mutex
	policy   WatchdogPolicy
	lastSeen time.Time
	out      io.Writer
	stop     chan struct{}
}

// NewWatchdog creates a Watchdog with the given policy that writes stall
// alerts to out.
func NewWatchdog(policy WatchdogPolicy, out io.Writer) *Watchdog {
	if policy.MaxSilence <= 0 {
		policy.MaxSilence = DefaultWatchdogPolicy().MaxSilence
	}
	if policy.CheckInterval <= 0 {
		policy.CheckInterval = DefaultWatchdogPolicy().CheckInterval
	}
	return &Watchdog{
		policy:   policy,
		out:      out,
		lastSeen: time.Now(),
		stop:     make(chan struct{}),
	}
}

// Ping records that the detector is alive. Call this after every successful
// detection cycle.
func (w *Watchdog) Ping() {
	w.mu.Lock()
	w.lastSeen = time.Now()
	w.mu.Unlock()
}

// Start begins background monitoring. It is safe to call Start once.
func (w *Watchdog) Start() {
	go func() {
		ticker := time.NewTicker(w.policy.CheckInterval)
		defer ticker.Stop()
		for {
			select {
			case <-w.stop:
				return
			case <-ticker.C:
				w.mu.Lock()
				since := time.Since(w.lastSeen)
				w.mu.Unlock()
				if since > w.policy.MaxSilence {
					fmt.Fprintf(w.out, "[watchdog] detector stalled: no ping for %s\n", since.Round(time.Second))
				}
			}
		}
	}()
}

// Stop shuts down the background monitor.
func (w *Watchdog) Stop() {
	close(w.stop)
}

// Stalled reports whether the detector is currently considered stalled.
func (w *Watchdog) Stalled() bool {
	w.mu.Lock()
	defer w.mu.Unlock()
	return time.Since(w.lastSeen) > w.policy.MaxSilence
}
