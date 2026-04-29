// internal/adapter/csv/adapter_test.go
package csv

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestAdapter_Read(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		expectError bool
		expectLen   int
		opts        []Option
		checkRow    func(t *testing.T, rows []map[string]any)
	}{
		{
			name:        "basic CSV with header",
			data:        "id,name,email\n1,Alice,alice@example.com\n2,Bob,bob@example.com",
			expectError: false,
			expectLen:   2,
			checkRow: func(t *testing.T, rows []map[string]any) {
				if rows[0]["name"] != "Alice" {
					t.Errorf("expected name Alice, got %v", rows[0]["name"])
				}
				if rows[0]["email"] != "alice@example.com" {
					t.Errorf("expected email alice@example.com, got %v", rows[0]["email"])
				}
			},
		},
		{
			name:        "CSV without header",
			data:        "1,Alice,alice@example.com\n2,Bob,bob@example.com",
			expectError: false,
			expectLen:   2,
			opts:        []Option{func(a *Adapter) { a.hasHeader = false }},
			checkRow: func(t *testing.T, rows []map[string]any) {
				if rows[0]["col0"] != "1" {
					t.Errorf("expected col0=1, got %v", rows[0]["col0"])
				}
				if rows[0]["col1"] != "Alice" {
					t.Errorf("expected col1=Alice, got %v", rows[0]["col1"])
				}
			},
		},
		{
			name:        "non-existent file",
			data:        "",
			expectError: true,
			expectLen:   0,
			opts:        []Option{func(a *Adapter) { a.path = "/nonexistent/file.csv" }},
		},
		{
			name:        "empty file",
			data:        "",
			expectError: false,
			expectLen:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp := t.TempDir()

			var path string
			if tt.name == "non-existent file" {
				path = "/nonexistent/file.csv"
			} else if tt.name == "empty file" {
				path = tmp + "/empty.csv"
				os.WriteFile(path, []byte(""), 0644)
			} else {
				path = tmp + "/test.csv"
				os.WriteFile(path, []byte(tt.data), 0644)
			}

			a := New(path, tt.opts...)
			rows, err := a.Read(context.Background())

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Read failed: %v", err)
			}

			if len(rows) != tt.expectLen {
				t.Errorf("expected %d rows, got %d", tt.expectLen, len(rows))
			}

			if tt.checkRow != nil && len(rows) > 0 {
				tt.checkRow(t, rows)
			}
		})
	}
}

func TestAdapter_Read_CancelledContext(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/test.csv"
	data := "id,name\n1,Alice\n2,Bob\n3,Charlie\n4,Diana\n5,Eve"
	os.WriteFile(path, []byte(data), 0644)

	a := New(path)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := a.Read(ctx)
	if err == nil {
		t.Error("expected error with cancelled context, got nil")
	}
}

func TestAdapter_Read_Timeout(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/test.csv"
	data := "id,name\n1,Alice"
	os.WriteFile(path, []byte(data), 0644)

	a := New(path)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond)

	_, err := a.Read(ctx)
	if err == nil {
		t.Error("expected error with timeout context, got nil")
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

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	if rows[0]["name"] != "Alice" {
		t.Errorf("expected name Alice, got %v", rows[0]["name"])
	}
}

func TestAdapter_EmptyFile(t *testing.T) {
	tmp := t.TempDir()
	path := tmp + "/empty.csv"
	os.WriteFile(path, []byte(""), 0644)

	a := New(path)
	rows, err := a.Read(context.Background())
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	if len(rows) != 0 {
		t.Errorf("expected 0 rows for empty file, got %d", len(rows))
	}
}