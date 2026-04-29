// internal/mapper/config.go
package mapper

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// TargetConfig 目标对象配置
type TargetConfig struct {
	ObjectType string `yaml:"object_type"`
}

// MappingConfig 映射配置
type MappingConfig struct {
	Source struct {
		Type  string `yaml:"type"`  // json | csv | sql
		Path  string `yaml:"path"`
		Query string `yaml:"query"`
	} `yaml:"source"`

	Env map[string]string `yaml:"env"`

	Target TargetConfig `yaml:"target"`

	Fields []FieldMapping `yaml:"fields"`
}

// FieldMapping 字段映射
type FieldMapping struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
	Type   string `yaml:"type"`  // string | int | float | bool | datetime
	Link   string `yaml:"link,omitempty"`
}

// LoadConfig 加载映射配置
func LoadConfig(path string) (*MappingConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	var cfg MappingConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return &cfg, nil
}

// Validate validates the mapping configuration
func (c *MappingConfig) Validate() error {
	if c.Source.Type == "" {
		return fmt.Errorf("source.type is required")
	}
	if len(c.Fields) == 0 {
		return fmt.Errorf("fields cannot be empty")
	}
	return nil
}