package drift

import (
	"fmt"
	"io"
	"sort"
	"time"
)

// HeatmapCell holds the drift count for a service at a given hour bucket.
type HeatmapCell struct {
	Service string
	Hour    time.Time
	Count   int
}

// DriftHeatmap tracks drift event frequency per service across hourly buckets.
type DriftHeatmap struct {
	cells map[string]map[time.Time]int
}

// NewDriftHeatmap creates an empty DriftHeatmap.
func NewDriftHeatmap() *DriftHeatmap {
	return &DriftHeatmap{
		cells: make(map[string]map[time.Time]int),
	}
}

// Record registers a drift result in the heatmap if it has drift.
func (h *DriftHeatmap) Record(result DriftResult) {
	if !result.HasDrift() {
		return
	}
	bucket := result.CheckedAt.UTC().Truncate(time.Hour)
	if _, ok := h.cells[result.Service]; !ok {
		h.cells[result.Service] = make(map[time.Time]int)
	}
	h.cells[result.Service][bucket]++
}

// Cells returns all heatmap cells sorted by service then hour.
func (h *DriftHeatmap) Cells() []HeatmapCell {
	var out []HeatmapCell
	for svc, buckets := range h.cells {
		for hour, count := range buckets {
			out = append(out, HeatmapCell{Service: svc, Hour: hour, Count: count})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Service != out[j].Service {
			return out[i].Service < out[j].Service
		}
		return out[i].Hour.Before(out[j].Hour)
	})
	return out
}

// WriteTo renders the heatmap as a text table to w.
func (h *DriftHeatmap) WriteTo(w io.Writer) error {
	cells := h.Cells()
	if len(cells) == 0 {
		_, err := fmt.Fprintln(w, "heatmap: no drift recorded")
		return err
	}
	_, err := fmt.Fprintf(w, "%-30s %-20s %s\n", "SERVICE", "HOUR (UTC)", "COUNT")
	if err != nil {
		return err
	}
	for _, c := range cells {
		_, err = fmt.Fprintf(w, "%-30s %-20s %d\n",
			c.Service, c.Hour.Format("2006-01-02T15:04"), c.Count)
		if err != nil {
			return err
		}
	}
	return nil
}
