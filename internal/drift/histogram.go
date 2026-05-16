package drift

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// HistogramBucket holds the count of drift events within a time window.
type HistogramBucket struct {
	Label string
	Count int
}

// DriftHistogram tracks drift event counts bucketed by a configurable duration.
type DriftHistogram struct {
	bucketSize time.Duration
	buckets    map[int64]int // unix epoch / bucketSize -> count
}

// NewDriftHistogram creates a DriftHistogram with the given bucket size.
// A zero or negative bucket size defaults to one hour.
func NewDriftHistogram(bucketSize time.Duration) *DriftHistogram {
	if bucketSize <= 0 {
		bucketSize = time.Hour
	}
	return &DriftHistogram{
		bucketSize: bucketSize,
		buckets:    make(map[int64]int),
	}
}

// Record increments the bucket corresponding to the given timestamp if the
// result contains drift.
func (h *DriftHistogram) Record(result DriftResult, at time.Time) {
	if !result.HasDrift() {
		return
	}
	key := at.UnixNano() / int64(h.bucketSize)
	h.buckets[key]++
}

// Buckets returns a sorted slice of HistogramBucket values.
func (h *DriftHistogram) Buckets() []HistogramBucket {
	keys := make([]int64, 0, len(h.buckets))
	for k := range h.buckets {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	out := make([]HistogramBucket, 0, len(keys))
	for _, k := range keys {
		t := time.Unix(0, k*int64(h.bucketSize)).UTC()
		out = append(out, HistogramBucket{
			Label: t.Format("2006-01-02T15:04"),
			Count: h.buckets[k],
		})
	}
	return out
}

// WriteTo writes a human-readable histogram to w.
func (h *DriftHistogram) WriteTo(w io.Writer) (int64, error) {
	buckets := h.Buckets()
	if len(buckets) == 0 {
		n, err := fmt.Fprintln(w, "histogram: no drift events recorded")
		return int64(n), err
	}
	var total int64
	for _, b := range buckets {
		n, err := fmt.Fprintf(w, "  %s  %s\n", b.Label, renderBar(b.Count))
		total += int64(n)
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

func renderBar(count int) string {
	bar := ""
	for i := 0; i < count; i++ {
		bar += "█"
	}
	return fmt.Sprintf("%s (%d)", bar, count)
}
