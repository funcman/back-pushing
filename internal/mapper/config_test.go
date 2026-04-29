// internal/mapper/config_test.go
package mapper

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/mapping.yaml"

	yaml := `
source:
  type: csv
  path: ./data/persons.csv

env:
  DB_URL: DATABASE_URL

target:
  object_type: Person

fields:
  - source: id
    target: id
    type: string
  - source: name
    target: name
    type: string
  - source: org_id
    target: organization_id
    type: string
    link: works_at
`
	os.WriteFile(path, []byte(yaml), 0644)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Source.Type != "csv" {
		t.Errorf("expected type csv, got %s", cfg.Source.Type)
	}

	if cfg.Source.Path != "./data/persons.csv" {
		t.Errorf("expected path ./data/persons.csv, got %s", cfg.Source.Path)
	}

	if cfg.Target.ObjectType != "Person" {
		t.Errorf("expected object_type Person, got %s", cfg.Target.ObjectType)
	}

	if cfg.Env["DB_URL"] != "DATABASE_URL" {
		t.Errorf("expected DB_URL env var, got %s", cfg.Env["DB_URL"])
	}

	if len(cfg.Fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(cfg.Fields))
	}

	if cfg.Fields[2].Link != "works_at" {
		t.Errorf("expected link works_at, got %s", cfg.Fields[2].Link)
	}
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := LoadConfig("/nonexistent/path.yaml")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestLoadConfig_MalformedYAML(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/malformed.yaml"

	// Malformed YAML - invalid syntax
	yaml := `
source:
  type: csv
  path: ./data
  invalid: yaml: content
`
	os.WriteFile(path, []byte(yaml), 0644)

	_, err := LoadConfig(path)
	if err == nil {
		t.Error("expected error for malformed YAML")
	}
}

func TestLoadConfig_EmptyFields(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/empty_fields.yaml"

	// Empty Fields slice - per requirements, this is valid in our design
	yaml := `
source:
  type: csv
  path: ./data/persons.csv

target:
  object_type: Person

fields: []
`
	os.WriteFile(path, []byte(yaml), 0644)

	cfg, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Empty fields is allowed by LoadConfig, but Validate should catch it
	if len(cfg.Fields) != 0 {
		t.Errorf("expected 0 fields, got %d", len(cfg.Fields))
	}
}

func TestMappingConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  MappingConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			config: MappingConfig{
				Source: struct {
					Type  string `yaml:"type"`
					Path  string `yaml:"path"`
					Query string `yaml:"query"`
				}{
					Type: "csv",
					Path: "./data/test.csv",
				},
				Fields: []FieldMapping{
					{Source: "id", Target: "id", Type: "string"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing source type",
			config: MappingConfig{
				Source: struct {
					Type  string `yaml:"type"`
					Path  string `yaml:"path"`
					Query string `yaml:"query"`
				}{
					Type: "",
					Path: "./data/test.csv",
				},
				Fields: []FieldMapping{
					{Source: "id", Target: "id", Type: "string"},
				},
			},
			wantErr: true,
			errMsg:  "source.type is required",
		},
		{
			name: "empty fields",
			config: MappingConfig{
				Source: struct {
					Type  string `yaml:"type"`
					Path  string `yaml:"path"`
					Query string `yaml:"query"`
				}{
					Type: "csv",
					Path: "./data/test.csv",
				},
				Fields: []FieldMapping{},
			},
			wantErr: true,
			errMsg:  "fields cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.errMsg {
				t.Errorf("Validate() error = %v, want %v", err, tt.errMsg)
			}
		})
	}
}