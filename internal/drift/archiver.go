package drift

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ArchivePolicy controls how results are archived.
type ArchivePolicy struct {
	// MaxAge is the maximum age of archived results before they are pruned.
	MaxAge time.Duration
	// Dir is the directory where archives are stored.
	Dir string
}

// DefaultArchivePolicy returns a policy with sensible defaults.
func DefaultArchivePolicy() ArchivePolicy {
	return ArchivePolicy{
		MaxAge: 30 * 24 * time.Hour,
		Dir:    "data/archive",
	}
}

// Archiver persists DriftResult values to disk and can retrieve them by date.
type Archiver struct {
	policy ArchivePolicy
}

// NewArchiver creates an Archiver, ensuring the archive directory exists.
func NewArchiver(policy ArchivePolicy) (*Archiver, error) {
	if policy.Dir == "" {
		policy.Dir = DefaultArchivePolicy().Dir
	}
	if policy.MaxAge <= 0 {
		policy.MaxAge = DefaultArchivePolicy().MaxAge
	}
	if err := os.MkdirAll(policy.Dir, 0o755); err != nil {
		return nil, fmt.Errorf("archiver: create dir: %w", err)
	}
	return &Archiver{policy: policy}, nil
}

// Archive writes result to a timestamped JSON file under the archive directory.
func (a *Archiver) Archive(result DriftResult) error {
	ts := time.Now().UTC().Format("20060102T150405Z")
	name := fmt.Sprintf("%s-%s.json", result.Service, ts)
	path := filepath.Join(a.policy.Dir, name)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("archiver: create file: %w", err)
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(result)
}

// Prune removes archived files older than the policy MaxAge.
func (a *Archiver) Prune() (int, error) {
	entries, err := os.ReadDir(a.policy.Dir)
	if err != nil {
		return 0, fmt.Errorf("archiver: read dir: %w", err)
	}
	cutoff := time.Now().UTC().Add(-a.policy.MaxAge)
	var removed int
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filepath.Join(a.policy.Dir, e.Name()))
			removed++
		}
	}
	return removed, nil
}

// Count returns the number of archived files currently on disk.
func (a *Archiver) Count() (int, error) {
	entries, err := os.ReadDir(a.policy.Dir)
	if err != nil {
		return 0, fmt.Errorf("archiver: read dir: %w", err)
	}
	var n int
	for _, e := range entries {
		if !e.IsDir() {
			n++
		}
	}
	return n, nil
}
