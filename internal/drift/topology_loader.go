package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// topologyNodeConfig is the YAML representation of a topology node.
type topologyNodeConfig struct {
	Service      string            `yaml:"service"`
	Dependencies []string          `yaml:"dependencies"`
	Labels       map[string]string `yaml:"labels"`
}

// topologyConfig is the root YAML structure for a topology file.
type topologyConfig struct {
	Nodes []topologyNodeConfig `yaml:"nodes"`
}

// LoadTopologyConfig reads a topology YAML file from the given path.
func LoadTopologyConfig(path string) (*TopologyGraph, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("topology: read file: %w", err)
	}
	return LoadTopologyConfigFromBytes(data)
}

// LoadTopologyConfigFromBytes parses a topology YAML payload and returns a
// populated TopologyGraph.
func LoadTopologyConfigFromBytes(data []byte) (*TopologyGraph, error) {
	var cfg topologyConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("topology: parse yaml: %w", err)
	}
	g := NewTopologyGraph()
	for _, n := range cfg.Nodes {
		if n.Service == "" {
			return nil, fmt.Errorf("topology: node missing required 'service' field")
		}
		g.AddNode(TopologyNode{
			Service:      n.Service,
			Dependencies: n.Dependencies,
			Labels:       n.Labels,
		})
	}
	return g, nil
}
