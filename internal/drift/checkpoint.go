package drift

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Checkpoint records the last successful run state for a service.
type Checkpoint struct {
	Service   string    `json:"service"`
	RunAt     time.Time `json:"run_at"`
	Drifted   bool      `json:"drifted"`
	EntryCount int      `json:"entry_count"`
}

// CheckpointStore persists and retrieves run checkpoints.
type CheckpointStore struct {
	dir string
}

// NewCheckpointStore creates a CheckpointStore backed by the given directory.
func NewCheckpointStore(dir string) (*CheckpointStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("checkpoint: create dir: %w", err)
	}
	return &CheckpointStore{dir: dir}, nil
}

func (s *CheckpointStore) path(service string) string {
	return filepath.Join(s.dir, service+".checkpoint.json")
}

// Save writes a checkpoint for the given result.
func (s *CheckpointStore) Save(result DriftResult) error {
	cp := Checkpoint{
		Service:    result.Service,
		RunAt:      time.Now().UTC(),
		Drifted:    result.HasDrift(),
		EntryCount: len(result.Entries),
	}
	data, err := json.MarshalIndent(cp, "", "  ")
	if err != nil {
		return fmt.Errorf("checkpoint: marshal: %w", err)
	}
	return os.WriteFile(s.path(result.Service), data, 0o644)
}

// Load retrieves the last checkpoint for a service.
func (s *CheckpointStore) Load(service string) (Checkpoint, error) {
	data, err := os.ReadFile(s.path(service))
	if err != nil {
		if os.IsNotExist(err) {
			return Checkpoint{}, fmt.Errorf("checkpoint: not found: %s", service)
		}
		return Checkpoint{}, fmt.Errorf("checkpoint: read: %w", err)
	}
	var cp Checkpoint
	if err := json.Unmarshal(data, &cp); err != nil {
		return Checkpoint{}, fmt.Errorf("checkpoint: unmarshal: %w", err)
	}
	return cp, nil
}

// Exists reports whether a checkpoint exists for the given service.
func (s *CheckpointStore) Exists(service string) bool {
	_, err := os.Stat(s.path(service))
	return err == nil
}

// Delete removes the checkpoint for the given service.
func (s *CheckpointStore) Delete(service string) error {
	err := os.Remove(s.path(service))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("checkpoint: delete: %w", err)
	}
	return nil
}
