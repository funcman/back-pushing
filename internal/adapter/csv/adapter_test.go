// internal/adapter/csv/adapter_test.go
package csv

import (
    "context"
    "os"
    "testing"
)

func TestAdapter_Read(t *testing.T) {
    tests := []struct {
        name     string
        data     string
        expected map[string]any
    }{
        {
            name: "basic CSV with header",
            data: "id,name,email\n1,Alice,alice@example.com\n2,Bob,bob@example.com",
            expected: map[string]any{"name": "Alice", "email": "alice@example.com"},
        },
        {
            name: "CSV without header",
            data: "1,Alice,alice@example.com\n2,Bob,bob@example.com",
            expected: map[string]any{"col0": "1", "col1": "Alice"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmp := t.TempDir()
            path := tmp + "/test.csv"
            os.WriteFile(path, []byte(tt.data), 0644)

            a := New(path)
            rows, err := a.Read(context.Background())
            if err != nil {
                t.Fatalf("Read failed: %v", err)
            }

            if tt.name == "basic CSV with header" {
                if len(rows) != 2 {
                    t.Errorf("expected 2 rows, got %d", len(rows))
                }
                if rows[0]["name"] != "Alice" {
                    t.Errorf("expected name Alice, got %v", rows[0]["name"])
                }
            }
        })
    }
}

func TestAdapter_Delimiter(t *testing.T) {
    tmp := t.TempDir()
    path := tmp + "/test.tsv"
    data := "id\tname\n1\tAlice"
    os.WriteFile(path, []byte(data), 0644)

    a := New(path, WithDelimiter('\t'))
    rows, err := a.Read(context.Background())
    if err != nil {
        t.Fatalf("Read failed: %v", err)
    }

    if rows[0]["name"] != "Alice" {
        t.Errorf("expected name Alice, got %v", rows[0]["name"])
    }
}