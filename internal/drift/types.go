package drift

// DriftKind identifies the category of a detected drift.
type DriftKind string

const (
	KindImage       DriftKind = "image"
	KindReplicas    DriftKind = "replicas"
	KindEnvVar      DriftKind = "env_var"
	KindMissingEnv  DriftKind = "missing_env"
	KindExtraEnv    DriftKind = "extra_env"
)

// DriftEntry describes a single field that has drifted.
type DriftEntry struct {
	Kind     DriftKind   `json:"kind"`
	Field    string      `json:"field"`
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
}

// HasDrift returns true when the entry represents a real difference.
func (e DriftEntry) HasDrift() bool {
	return e.Expected != e.Actual
}

// DriftReport is the result of comparing a live service against its spec.
type DriftReport struct {
	Service string       `json:"service"`
	Entries []DriftEntry `json:"entries"`
}

// HasDrift returns true when at least one entry was recorded.
func (r *DriftReport) HasDrift() bool {
	return len(r.Entries) > 0
}

// ServiceSpec is the declared desired state loaded from IaC definitions.
type ServiceSpec struct {
	Name     string            `yaml:"name"     json:"name"`
	Image    string            `yaml:"image"    json:"image"`
	Replicas int               `yaml:"replicas" json:"replicas"`
	Env      map[string]string `yaml:"env"      json:"env"`
}

// ApplyDefaults fills in zero-value fields with sensible defaults.
func (s *ServiceSpec) ApplyDefaults() {
	if s.Replicas == 0 {
		s.Replicas = 1
	}
	if s.Env == nil {
		s.Env = make(map[string]string)
	}
}
