package drift

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MapperConfig holds YAML-decoded configuration for a Mapper.
type MapperConfig struct {
	Mode string `yaml:"mode"`
}

// LoadMapperConfig reads a MapperConfig from a YAML file at path.
func LoadMapperConfig(path string) (*MapperConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("mapper: read file: %w", err)
	}
	return LoadMapperConfigFromBytes(data)
}

// LoadMapperConfigFromBytes parses a MapperConfig from raw YAML bytes.
func LoadMapperConfigFromBytes(data []byte) (*MapperConfig, error) {
	var cfg MapperConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("mapper: parse yaml: %w", err)
	}
	if cfg.Mode == "" {
		cfg.Mode = string(MapByService)
	}
	return &cfg, nil
}

// BuildMapper constructs a Mapper from a MapperConfig.
func BuildMapper(cfg *MapperConfig) *Mapper {
	return NewMapper(MapMode(cfg.Mode))
}
