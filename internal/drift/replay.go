package drift

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// ReplayEvent represents a single recorded drift event for replay.
type ReplayEvent struct {
	Timestamp time.Time  `json:"timestamp" yaml:"timestamp"`
	Result    DriftResult `json:"result"    yaml:"result"`
}

// Replayer replays recorded drift events through a handler for historical analysis.
type Replayer struct {
	events  []ReplayEvent
	handler func(ReplayEvent) error
	speed   float64 // 0 means instant; >0 scales original timing
}

// NewReplayer constructs a Replayer with the given events and handler.
// speed controls playback: 0 = instant, 1.0 = real-time, 0.5 = half-speed.
func NewReplayer(events []ReplayEvent, handler func(ReplayEvent) error, speed float64) *Replayer {
	sorted := make([]ReplayEvent, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Timestamp.Before(sorted[j].Timestamp)
	})
	if speed < 0 {
		speed = 0
	}
	return &Replayer{events: sorted, handler: handler, speed: speed}
}

// Run replays all events in chronological order.
func (r *Replayer) Run() error {
	for i, ev := range r.events {
		if r.speed > 0 && i > 0 {
			gap := r.events[i].Timestamp.Sub(r.events[i-1].Timestamp)
			scaled := time.Duration(float64(gap) / r.speed)
			time.Sleep(scaled)
		}
		if err := r.handler(ev); err != nil {
			return fmt.Errorf("replay handler error at %s: %w", ev.Timestamp.Format(time.RFC3339), err)
		}
	}
	return nil
}

// Len returns the number of events queued for replay.
func (r *Replayer) Len() int { return len(r.events) }

// WriteReplaySummary writes a human-readable summary of replay events to w.
func WriteReplaySummary(events []ReplayEvent, w io.Writer) error {
	_, err := fmt.Fprintf(w, "Replay: %d event(s)\n", len(events))
	if err != nil {
		return err
	}
	for _, ev := range events {
		status := "clean"
		if ev.Result.HasDrift() {
			status = fmt.Sprintf("drifted (%d entries)", len(ev.Result.Entries))
		}
		_, err = fmt.Fprintf(w, "  [%s] %s — %s\n",
			ev.Timestamp.Format(time.RFC3339), ev.Result.Service, status)
		if err != nil {
			return err
		}
	}
	return nil
}
