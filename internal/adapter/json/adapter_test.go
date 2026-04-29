// internal/adapter/json/adapter_test.go
package json

import (
    "context"
    "errors"
    "os"
    "testing"
)

func TestAdapter_Read(t *testing.T) {
    tests := []struct {
        name      string
        path      string
        wantLen   int
        checkErr  func(err error) bool
    }{
        {
            name:    "valid JSON file",
            path:    writeTestFile(t, `[{"id": "1", "name": "Alice"}, {"id": "2", "name": "Bob"}]`),
            wantLen: 2,
            checkErr: func(err error) bool { return err == nil },
        },
        {
            name:    "non-existent file",
            path:    "/non/existent/file.json",
            wantLen: 0,
            checkErr: func(err error) bool { return err != nil && errors.Is(err, os.ErrNotExist) },
        },
        {
            name:    "invalid JSON",
            path:    writeTestFile(t, `{not valid json}`),
            wantLen: 0,
            checkErr: func(err error) bool {
                return err != nil && (errors.Is(err, os.ErrInvalid) || err.Error() != "")
            },
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            a := New(tt.path)
            rows, err := a.Read(context.Background())

            if !tt.checkErr(err) {
                if err == nil {
                    t.Errorf("expected error, got nil")
                } else {
                    t.Errorf("unexpected error: %v", err)
                }
                return
            }

            if err == nil && len(rows) != tt.wantLen {
                t.Errorf("expected %d rows, got %d", tt.wantLen, len(rows))
            }
        })
    }
}

func TestAdapter_Read_ContextCanceled(t *testing.T) {
    tmp := t.TempDir()
    path := tmp + "/test.json"
    writeTestFile(t, `[{"id": "1"}]`)

    a := New(path)
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately

    _, err := a.Read(ctx)
    if err == nil {
        t.Error("expected error when context is canceled")
    }
    if !errors.Is(err, context.Canceled) {
        t.Errorf("expected context.Canceled, got %v", err)
    }
}

func writeTestFile(t *testing.T, data string) string {
    t.Helper()
    tmp := t.TempDir()
    path := tmp + "/test.json"
    if err := os.WriteFile(path, []byte(data), 0644); err != nil {
        t.Fatalf("failed to write test file: %v", err)
    }
    return path
}
