package drift

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Snapshot captures a point-in-time record of a drift detection result.
type Snapshot struct {
	Service   string       `json:"service"`
	Timestamp time.Time    `json:"timestamp"`
	Result    DriftResult  `json:"result"`
}

// SnapshotStore persists and retrieves snapshots.
type SnapshotStore struct {
	dir string
}

// NewSnapshotStore creates a SnapshotStore backed by the given directory.
func NewSnapshotStore(dir string) (*SnapshotStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("snapshot: create dir: %w", err)
	}
	return &SnapshotStore{dir: dir}, nil
}

// Save writes a snapshot for the given result to disk.
func (s *SnapshotStore) Save(result DriftResult) error {
	snap := Snapshot{
		Service:   result.Service,
		Timestamp: time.Now().UTC(),
		Result:    result,
	}
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return fmt.Errorf("snapshot: marshal: %w", err)
	}
	path := s.filePath(result.Service)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("snapshot: write: %w", err)
	}
	return nil
}

// Load reads the latest snapshot for the named service.
func (s *SnapshotStore) Load(service string) (Snapshot, error) {
	path := s.filePath(service)
	data, err := os.ReadFile(path)
	if err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: read: %w", err)
	}
	var snap Snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return Snapshot{}, fmt.Errorf("snapshot: unmarshal: %w", err)
	}
	return snap, nil
}

// Exists reports whether a snapshot exists for the named service.
func (s *SnapshotStore) Exists(service string) bool {
	_, err := os.Stat(s.filePath(service))
	return err == nil
}

func (s *SnapshotStore) filePath(service string) string {
	return fmt.Sprintf("%s/%s.json", s.dir, service)
}
