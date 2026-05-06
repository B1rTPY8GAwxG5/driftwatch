package drift

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Baseline represents a saved reference state for a service spec.
type Baseline struct {
	Service   string      `json:"service"`
	Spec      ServiceSpec `json:"spec"`
	CreatedAt time.Time   `json:"created_at"`
}

// BaselineStore persists and retrieves baselines for services.
type BaselineStore struct {
	dir string
}

// NewBaselineStore creates a BaselineStore rooted at dir.
func NewBaselineStore(dir string) (*BaselineStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("baseline: create dir: %w", err)
	}
	return &BaselineStore{dir: dir}, nil
}

func (s *BaselineStore) path(service string) string {
	return filepath.Join(s.dir, service+".baseline.json")
}

// Save writes the given spec as the baseline for service.
func (s *BaselineStore) Save(service string, spec ServiceSpec) error {
	b := Baseline{
		Service:   service,
		Spec:      spec,
		CreatedAt: time.Now().UTC(),
	}
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("baseline: marshal: %w", err)
	}
	if err := os.WriteFile(s.path(service), data, 0o644); err != nil {
		return fmt.Errorf("baseline: write: %w", err)
	}
	return nil
}

// Load retrieves the baseline for service. Returns an error if not found.
func (s *BaselineStore) Load(service string) (*Baseline, error) {
	data, err := os.ReadFile(s.path(service))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("baseline: not found: %s", service)
		}
		return nil, fmt.Errorf("baseline: read: %w", err)
	}
	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("baseline: unmarshal: %w", err)
	}
	return &b, nil
}

// Exists reports whether a baseline exists for service.
func (s *BaselineStore) Exists(service string) bool {
	_, err := os.Stat(s.path(service))
	return err == nil
}

// Delete removes the baseline for service.
func (s *BaselineStore) Delete(service string) error {
	if err := os.Remove(s.path(service)); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("baseline: delete: %w", err)
	}
	return nil
}
