package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ServiceSpec represents the declared infrastructure-as-code definition of a service.
type ServiceSpec struct {
	Name     string            `yaml:"name"`
	Image    string            `yaml:"image"`
	Replicas int               `yaml:"replicas"`
	Env      map[string]string `yaml:"env"`
	Ports    []int             `yaml:"ports"`
}

// DeployedService represents the live state of a deployed service.
type DeployedService struct {
	Name     string
	Image    string
	Replicas int
	Env      map[string]string
	Ports    []int
}

// LoadSpec reads a ServiceSpec from a YAML file at the given path.
func LoadSpec(path string) (*ServiceSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading spec file %q: %w", path, err)
	}

	var spec ServiceSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing spec file %q: %w", path, err)
	}

	if spec.Name == "" {
		return nil, fmt.Errorf("spec file %q: missing required field 'name'", path)
	}
	if spec.Image == "" {
		return nil, fmt.Errorf("spec file %q: missing required field 'image'", path)
	}

	return &spec, nil
}

// LoadSpecFromBytes parses a ServiceSpec from raw YAML bytes.
func LoadSpecFromBytes(data []byte) (*ServiceSpec, error) {
	var spec ServiceSpec
	if err := yaml.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing spec: %w", err)
	}
	return &spec, nil
}
