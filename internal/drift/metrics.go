package drift

import (
	"fmt"
	"io"
	"time"
)

// MetricKind identifies the type of metric being recorded.
type MetricKind string

const (
	MetricDriftDetected  MetricKind = "drift_detected"
	MetricDriftClean     MetricKind = "drift_clean"
	MetricDetectionError MetricKind = "detection_error"
)

// Metric represents a single recorded observation from a drift check.
type Metric struct {
	Service   string
	Kind      MetricKind
	DriftCount int
	Timestamp time.Time
}

// String returns a human-readable representation of the metric.
func (m Metric) String() string {
	return fmt.Sprintf("[%s] service=%s kind=%s drift_count=%d",
		m.Timestamp.Format(time.RFC3339), m.Service, m.Kind, m.DriftCount)
}

// MetricsRecorder collects metrics from drift detection runs.
type MetricsRecorder struct {
	w       io.Writer
	metrics []Metric
}

// NewMetricsRecorder creates a MetricsRecorder that writes observations to w.
func NewMetricsRecorder(w io.Writer) *MetricsRecorder {
	return &MetricsRecorder{w: w}
}

// Record stores a metric derived from a DriftResult and writes it to the writer.
func (r *MetricsRecorder) Record(result DriftResult) error {
	kind := MetricDriftClean
	if result.HasDrift() {
		kind = MetricDriftDetected
	}

	m := Metric{
		Service:    result.Service,
		Kind:       kind,
		DriftCount: len(result.Entries),
		Timestamp:  time.Now().UTC(),
	}

	r.metrics = append(r.metrics, m)

	_, err := fmt.Fprintln(r.w, m.String())
	return err
}

// All returns all recorded metrics.
func (r *MetricsRecorder) All() []Metric {
	out := make([]Metric, len(r.metrics))
	copy(out, r.metrics)
	return out
}

// Summary returns counts of each MetricKind observed.
func (r *MetricsRecorder) Summary() map[MetricKind]int {
	counts := make(map[MetricKind]int)
	for _, m := range r.metrics {
		counts[m.Kind]++
	}
	return counts
}
