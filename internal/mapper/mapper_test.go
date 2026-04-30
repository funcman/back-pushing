// internal/mapper/mapper_test.go
package mapper

import (
	"context"
	"testing"
)

type mockSource struct {
	rows []map[string]any
}

func (m *mockSource) Read(ctx context.Context) ([]map[string]any, error) {
	return m.rows, nil
}

func (m *mockSource) Close() error {
	return nil
}

func TestMapper_Map(t *testing.T) {
	cfg := &MappingConfig{
		Target: TargetConfig{ObjectType: "Person"},
		Fields: []FieldMapping{
			{Source: "id", Target: "id", Type: "string"},
			{Source: "name", Target: "name", Type: "string"},
			{Source: "org_id", Target: "organization_id", Type: "string", Link: "works_at"},
		},
	}

	mapper := New(cfg)
	source := &mockSource{
		rows: []map[string]any{
			{"id": "1", "name": "Alice", "org_id": "org1"},
			{"id": "2", "name": "Bob", "org_id": "org2"},
		},
	}

	objects, err := mapper.Map(context.Background(), source)
	if err != nil {
		t.Fatalf("Map failed: %v", err)
	}

	if len(objects) != 2 {
		t.Errorf("expected 2 objects, got %d", len(objects))
	}

	if objects[0].Data["name"] != "Alice" {
		t.Errorf("expected name Alice, got %v", objects[0].Data["name"])
	}

	if objects[0].ID != "1" {
		t.Errorf("expected ID 1, got %s", objects[0].ID)
	}

	if len(objects[0].Links) != 1 {
		t.Errorf("expected 1 link, got %d", len(objects[0].Links))
	}

	if objects[0].Links[0].LinkType != "works_at" {
		t.Errorf("expected link type works_at, got %s", objects[0].Links[0].LinkType)
	}
}

func TestMapper_Convert(t *testing.T) {
	cfg := &MappingConfig{}
	mapper := New(cfg)

	tests := []struct {
		input any
		typ   string
		want  any
	}{
		{input: "hello", typ: "string", want: "hello"},
		{input: float64(42), typ: "int", want: 42},
		{input: float64(3.14), typ: "float", want: 3.14},
		{input: nil, typ: "string", want: nil},
	}

	for _, tt := range tests {
		got, err := mapper.convert(tt.input, tt.typ)
		if err != nil {
			t.Errorf("convert(%v, %s) returned error: %v", tt.input, tt.typ, err)
		}
		if got != tt.want {
			t.Errorf("convert(%v, %s) = %v, want %v", tt.input, tt.typ, got, tt.want)
		}
	}
}