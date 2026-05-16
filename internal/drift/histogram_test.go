package drift

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

var histogramDrifted = DriftResult{
	Service: "svc",
	Entries: []DriftEntry{
		{Kind: KindImage, Field: "image", Declared: "a", Observed: "b"},
	},
}

var histogramClean = DriftResult{Service: "svc", Entries: nil}

func TestNewDriftHistogram_DefaultBucket(t *testing.T) {
	h := NewDriftHistogram(0)
	if h.bucketSize != time.Hour {
		t.Fatalf("expected 1h bucket, got %v", h.bucketSize)
	}
}

func TestNewDriftHistogram_CustomBucket(t *testing.T) {
	h := NewDriftHistogram(15 * time.Minute)
	if h.bucketSize != 15*time.Minute {
		t.Fatalf("expected 15m bucket, got %v", h.bucketSize)
	}
}

func TestDriftHistogram_Record_CleanIgnored(t *testing.T) {
	h := NewDriftHistogram(time.Hour)
	h.Record(histogramClean, time.Now())
	if len(h.Buckets()) != 0 {
		t.Fatal("expected no buckets for clean result")
	}
}

func TestDriftHistogram_Record_DriftedCounted(t *testing.T) {
	h := NewDriftHistogram(time.Hour)
	now := time.Now()
	h.Record(histogramDrifted, now)
	h.Record(histogramDrifted, now)
	buckets := h.Buckets()
	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
	if buckets[0].Count != 2 {
		t.Fatalf("expected count 2, got %d", buckets[0].Count)
	}
}

func TestDriftHistogram_Buckets_Sorted(t *testing.T) {
	h := NewDriftHistogram(time.Hour)
	base := time.Now().Truncate(time.Hour)
	h.Record(histogramDrifted, base.Add(2*time.Hour))
	h.Record(histogramDrifted, base)
	h.Record(histogramDrifted, base.Add(time.Hour))
	buckets := h.Buckets()
	if len(buckets) != 3 {
		t.Fatalf("expected 3 buckets, got %d", len(buckets))
	}
	for i := 1; i < len(buckets); i++ {
		if buckets[i].Label < buckets[i-1].Label {
			t.Fatalf("buckets not sorted at index %d", i)
		}
	}
}

func TestDriftHistogram_WriteTo_NoDrift(t *testing.T) {
	h := NewDriftHistogram(time.Hour)
	var buf bytes.Buffer
	h.WriteTo(&buf)
	if !strings.Contains(buf.String(), "no drift events") {
		t.Fatalf("unexpected output: %s", buf.String())
	}
}

func TestDriftHistogram_WriteTo_WithDrift(t *testing.T) {
	h := NewDriftHistogram(time.Hour)
	h.Record(histogramDrifted, time.Now())
	var buf bytes.Buffer
	h.WriteTo(&buf)
	if !strings.Contains(buf.String(), "█") {
		t.Fatalf("expected bar chart in output, got: %s", buf.String())
	}
}
