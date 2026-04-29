// internal/mapper/config.go
package mapper

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MappingConfig 映射配置
type MappingConfig struct {
	Source struct {
		Type  string `yaml:"type"`  // json | csv | sql
		Path  string `yaml:"path"`
		Query string `yaml:"query"`
	} `yaml:"source"`

	Env map[string]string `yaml:"env"`

	Target struct {
		ObjectType string `yaml:"object_type"`
	} `yaml:"target"`

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