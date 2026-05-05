package drift

import "fmt"

// DriftStatus represents the comparison result between live and declared state.
type DriftStatus string

const (
	StatusMatch   DriftStatus = "match"
	StatusDrifted DriftStatus = "drifted"
	StatusMissing DriftStatus = "missing"
	StatusUnknown DriftStatus = "unknown"
)

// ServiceConfig holds the configuration fields for a single service.
type ServiceConfig struct {
	Name        string            `json:"name"`
	Image       string            `json:"image"`
	Replicas    int               `json:"replicas"`
	Environment map[string]string `json:"environment"`
	Labels      map[string]string `json:"labels"`
}

// DriftResult captures a single field-level drift finding.
type DriftResult struct {
	Service  string
	Field    string
	Declared interface{}
	Live     interface{}
	Status   DriftStatus
}

// Detector compares declared vs live service configurations.
type Detector struct{}

// NewDetector creates a new Detector instance.
func NewDetector() *Detector {
	return &Detector{}
}

// Compare returns a slice of DriftResult entries for each differing field.
func (d *Detector) Compare(declared, live ServiceConfig) []DriftResult {
	var results []DriftResult

	check := func(field string, decl, lv interface{}) {
		if fmt.Sprintf("%v", decl) != fmt.Sprintf("%v", lv) {
			results = append(results, DriftResult{
				Service:  declared.Name,
				Field:    field,
				Declared: decl,
				Live:     lv,
				Status:   StatusDrifted,
			})
		}
	}

	check("image", declared.Image, live.Image)
	check("replicas", declared.Replicas, live.Replicas)

	for k, dv := range declared.Environment {
		lv, ok := live.Environment[k]
		if !ok {
			results = append(results, DriftResult{Service: declared.Name, Field: "env:" + k, Declared: dv, Live: nil, Status: StatusMissing})
		} else if dv != lv {
			check("env:"+k, dv, lv)
		}
	}

	for k, dv := range declared.Labels {
		lv, ok := live.Labels[k]
		if !ok {
			results = append(results, DriftResult{Service: declared.Name, Field: "label:" + k, Declared: dv, Live: nil, Status: StatusMissing})
		} else if dv != lv {
			check("label:"+k, dv, lv)
		}
	}

	return results
}
