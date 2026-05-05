package drift

// ServiceSpec defines the desired state of a service as declared in IaC.
type ServiceSpec struct {
	Name     string            `yaml:"name"`
	Image    string            `yaml:"image"`
	Replicas int               `yaml:"replicas"`
	Env      map[string]string `yaml:"env"`
	Ports    []int             `yaml:"ports"`
}

// ServiceState represents the actual running state of a deployed service.
type ServiceState struct {
	Name     string
	Image    string
	Replicas int
	Env      map[string]string
	Ports    []int
}

// DriftKind categorizes the type of detected drift.
type DriftKind string

const (
	DriftKindImage    DriftKind = "image"
	DriftKindReplicas DriftKind = "replicas"
	DriftKindEnv      DriftKind = "env"
	DriftKindPort     DriftKind = "port"
	DriftKindMissing  DriftKind = "missing"
)

// DriftEntry describes a single drift finding between spec and live state.
type DriftEntry struct {
	Kind     DriftKind
	Field    string
	Expected string
	Actual   string
}

// DriftResult holds all drift entries found for a given service.
type DriftResult struct {
	ServiceName string
	Entries     []DriftEntry
}

// HasDrift returns true if any drift entries were recorded.
func (r DriftResult) HasDrift() bool {
	return len(r.Entries) > 0
}
