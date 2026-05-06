package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// PolicyFile is the top-level structure for a policy YAML file.
type PolicyFile struct {
	Policies []Policy `yaml:"policies"`
}

// LoadPolicy reads a Policy by name from a YAML file.
func LoadPolicy(path, name string) (*Policy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("policy: read file: %w", err)
	}
	return LoadPolicyFromBytes(data, name)
}

// LoadPolicyFromBytes parses policy YAML and returns the named policy.
func LoadPolicyFromBytes(data []byte, name string) (*Policy, error) {
	var pf PolicyFile
	if err := yaml.Unmarshal(data, &pf); err != nil {
		return nil, fmt.Errorf("policy: unmarshal: %w", err)
	}
	for i := range pf.Policies {
		if pf.Policies[i].Name == name {
			return &pf.Policies[i], nil
		}
	}
	return nil, fmt.Errorf("policy: %q not found", name)
}
