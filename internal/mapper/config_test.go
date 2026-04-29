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