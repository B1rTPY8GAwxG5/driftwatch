package drift

import (
	"context"
	"log"
	"time"
)

// WatchConfig holds configuration for the drift watcher.
type WatchConfig struct {
	Interval  time.Duration
	SpecPath  string
	ServiceFn func() (*ServiceSpec, error)
	OnDrift   func(DriftResult)
}

// Watcher periodically checks for configuration drift.
type Watcher struct {
	cfg      WatchConfig
	detector *Detector
	stop     chan struct{}
}

// NewWatcher creates a new Watcher with the given configuration.
func NewWatcher(cfg WatchConfig) *Watcher {
	return &Watcher{
		cfg:      cfg,
		detector: NewDetector(),
		stop:     make(chan struct{}),
	}
}

// Start begins periodic drift detection. Blocks until ctx is cancelled or Stop is called.
func (w *Watcher) Start(ctx context.Context) {
	ticker := time.NewTicker(w.cfg.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.check()
		case <-w.stop:
			return
		case <-ctx.Done():
			return
		}
	}
}

// Stop signals the watcher to cease polling.
func (w *Watcher) Stop() {
	close(w.stop)
}

func (w *Watcher) check() {
	spec, err := LoadSpec(w.cfg.SpecPath)
	if err != nil {
		log.Printf("driftwatch: failed to load spec: %v", err)
		return
	}

	live, err := w.cfg.ServiceFn()
	if err != nil {
		log.Printf("driftwatch: failed to fetch live service: %v", err)
		return
	}

	result := w.detector.Compare(spec, live)
	if result.HasDrift() && w.cfg.OnDrift != nil {
		w.cfg.OnDrift(result)
	}
}
