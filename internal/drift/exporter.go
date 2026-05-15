package drift

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// ExportFormat defines the output format for drift exports.
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatText ExportFormat = "text"
)

// ExportRecord represents a single exported drift result with metadata.
type ExportRecord struct {
	Service   string       `json:"service"`
	Timestamp time.Time    `json:"timestamp"`
	HasDrift  bool         `json:"has_drift"`
	Entries   []DriftEntry `json:"entries"`
	Score     int          `json:"score"`
}

// Exporter writes drift results to an output sink in a specified format.
type Exporter struct {
	format ExportFormat
	writer io.Writer
}

// NewExporter creates a new Exporter for the given format and writer.
// Returns an error if the format is unsupported.
func NewExporter(format ExportFormat, w io.Writer) (*Exporter, error) {
	if w == nil {
		return nil, fmt.Errorf("exporter: writer must not be nil")
	}
	switch format {
	case ExportFormatJSON, ExportFormatText:
		// valid
	default:
		return nil, fmt.Errorf("exporter: unsupported format %q", format)
	}
	return &Exporter{format: format, writer: w}, nil
}

// Export writes the given DriftResult as an ExportRecord to the exporter's writer.
func (e *Exporter) Export(result DriftResult) error {
	scored := ScoreResult(result)
	rec := ExportRecord{
		Service:   result.Service,
		Timestamp: time.Now().UTC(),
		HasDrift:  result.HasDrift(),
		Entries:   result.Entries,
		Score:     scored.Score,
	}
	switch e.format {
	case ExportFormatJSON:
		return e.writeJSON(rec)
	case ExportFormatText:
		return e.writeText(rec)
	}
	return nil
}

// ExportAll writes multiple DriftResults to the exporter's writer in sequence.
// It returns the first error encountered, along with the number of records
// successfully written before the failure.
func (e *Exporter) ExportAll(results []DriftResult) (int, error) {
	for i, result := range results {
		if err := e.Export(result); err != nil {
			return i, fmt.Errorf("exporter: failed on record %d (service %q): %w", i, result.Service, err)
		}
	}
	return len(results), nil
}

func (e *Exporter) writeJSON(rec ExportRecord) error {
	enc := json.NewEncoder(e.writer)
	enc.SetIndent("", "  ")
	return enc.Encode(rec)
}

func (e *Exporter) writeText(rec ExportRecord) error {
	status := "clean"
	if rec.HasDrift {
		status = "drifted"
	}
	_, err := fmt.Fprintf(e.writer, "[%s] service=%s status=%s score=%d entries=%d\n",
		rec.Timestamp.Format(time.RFC3339), rec.Service, status, rec.Score, len(rec.Entries))
	return err
}
