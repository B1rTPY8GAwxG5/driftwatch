package drift

import (
	"bytes"
	"strings"
	"testing"
)

func cleanDriftResult() DriftResult {
	return DriftResult{
		Service: "api",
		Entries: []DriftEntry{},
	}
}

func driftedDriftResult() DriftResult {
	return DriftResult{
		Service: "api",
		Entries: []DriftEntry{
			{Kind: KindImage, Field: "image", Declared: "v1", Observed: "v2"},
			{Kind: KindReplicas, Field: "replicas", Declared: "2", Observed: "3"},
		},
	}
}

func TestNewMetricsRecorder_NotNil(t *testing.T) {
	r := NewMetricsRecorder(&bytes.Buffer{})
	if r == nil {
		t.Fatal("expected non-nil MetricsRecorder")
	}
}

func TestMetricsRecorder_Record_CleanResult(t *testing.T) {
	var buf bytes.Buffer
	r := NewMetricsRecorder(&buf)

	if err := r.Record(cleanDriftResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	metrics := r.All()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Kind != MetricDriftClean {
		t.Errorf("expected kind %s, got %s", MetricDriftClean, metrics[0].Kind)
	}
	if metrics[0].DriftCount != 0 {
		t.Errorf("expected drift_count 0, got %d", metrics[0].DriftCount)
	}
}

func TestMetricsRecorder_Record_DriftedResult(t *testing.T) {
	var buf bytes.Buffer
	r := NewMetricsRecorder(&buf)

	if err := r.Record(driftedDriftResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	metrics := r.All()
	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Kind != MetricDriftDetected {
		t.Errorf("expected kind %s, got %s", MetricDriftDetected, metrics[0].Kind)
	}
	if metrics[0].DriftCount != 2 {
		t.Errorf("expected drift_count 2, got %d", metrics[0].DriftCount)
	}
}

func TestMetricsRecorder_WritesToWriter(t *testing.T) {
	var buf bytes.Buffer
	r := NewMetricsRecorder(&buf)

	_ = r.Record(driftedDriftResult())

	if !strings.Contains(buf.String(), "service=api") {
		t.Errorf("expected output to contain service name, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), string(MetricDriftDetected)) {
		t.Errorf("expected output to contain metric kind, got: %s", buf.String())
	}
}

func TestMetricsRecorder_Summary(t *testing.T) {
	var buf bytes.Buffer
	r := NewMetricsRecorder(&buf)

	_ = r.Record(cleanDriftResult())
	_ = r.Record(driftedDriftResult())
	_ = r.Record(driftedDriftResult())

	summary := r.Summary()
	if summary[MetricDriftClean] != 1 {
		t.Errorf("expected 1 clean, got %d", summary[MetricDriftClean])
	}
	if summary[MetricDriftDetected] != 2 {
		t.Errorf("expected 2 detected, got %d", summary[MetricDriftDetected])
	}
}

func TestMetricKind_Constants(t *testing.T) {
	if MetricDriftDetected != "drift_detected" {
		t.Errorf("unexpected value for MetricDriftDetected")
	}
	if MetricDriftClean != "drift_clean" {
		t.Errorf("unexpected value for MetricDriftClean")
	}
	if MetricDetectionError != "detection_error" {
		t.Errorf("unexpected value for MetricDetectionError")
	}
}
