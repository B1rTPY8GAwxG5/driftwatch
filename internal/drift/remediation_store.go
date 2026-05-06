package drift

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RemediationStore persists remediation plans to disk.
type RemediationStore struct {
	dir string
}

// NewRemediationStore creates a RemediationStore rooted at dir.
func NewRemediationStore(dir string) (*RemediationStore, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("remediation store: mkdir %s: %w", dir, err)
	}
	return &RemediationStore{dir: dir}, nil
}

type storedPlan struct {
	Service   string              `json:"service"`
	Actions   []RemediationAction `json:"actions"`
	SavedAt   time.Time           `json:"saved_at"`
}

// Save writes the plan to disk keyed by service name.
func (s *RemediationStore) Save(plan RemediationPlan) error {
	sp := storedPlan{
		Service: plan.Service,
		Actions: plan.Actions,
		SavedAt: time.Now().UTC(),
	}
	data, err := json.MarshalIndent(sp, "", "  ")
	if err != nil {
		return fmt.Errorf("remediation store: marshal: %w", err)
	}
	return os.WriteFile(s.path(plan.Service), data, 0o644)
}

// Load retrieves the plan for the given service.
func (s *RemediationStore) Load(service string) (RemediationPlan, error) {
	data, err := os.ReadFile(s.path(service))
	if err != nil {
		return RemediationPlan{}, fmt.Errorf("remediation store: load %s: %w", service, err)
	}
	var sp storedPlan
	if err := json.Unmarshal(data, &sp); err != nil {
		return RemediationPlan{}, fmt.Errorf("remediation store: unmarshal: %w", err)
	}
	return RemediationPlan{Service: sp.Service, Actions: sp.Actions}, nil
}

// Exists returns true when a plan for service is stored.
func (s *RemediationStore) Exists(service string) bool {
	_, err := os.Stat(s.path(service))
	return err == nil
}

func (s *RemediationStore) path(service string) string {
	return filepath.Join(s.dir, service+".json")
}
