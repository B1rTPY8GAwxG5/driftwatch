package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ProjectionConfig holds the YAML-decoded configuration for a Projection.
type ProjectionConfig struct {
	Fields []string `yaml:"fields"`
}

// LoadProjectionConfig reads a ProjectionConfig from a YAML file.
func LoadProjectionConfig(path string) (*ProjectionConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("projection: read file: %w", err)
	}
	return LoadProjectionConfigFromBytes(data)
}

// LoadProjectionConfigFromBytes parses a ProjectionConfig from raw YAML bytes.
func LoadProjectionConfigFromBytes(data []byte) (*ProjectionConfig, error) {
	var cfg ProjectionConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("projection: parse yaml: %w", err)
	}
	return &cfg, nil
}

// BuildProjection constructs a Projection from a ProjectionConfig.
// Unknown field names are silently ignored.
func BuildProjection(cfg *ProjectionConfig) *Projection {
	known := map[string]ProjectionField{
		"service":  ProjectionFieldService,
		"kind":     ProjectionFieldKind,
		"field":    ProjectionFieldField,
		"expected": ProjectionFieldExpected,
		"actual":   ProjectionFieldActual,
		"drifted":  ProjectionFieldDrifted,
		"time":     ProjectionFieldTime,
	}
	var fields []ProjectionField
	for _, name := range cfg.Fields {
		if f, ok := known[name]; ok {
			fields = append(fields, f)
		}
	}
	return NewProjection(fields...)
}
