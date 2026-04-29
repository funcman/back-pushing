// internal/adapter/json/adapter_test.go
package json

import (
    "context"
    "os"
    "testing"
)

func TestAdapter_Read(t *testing.T) {
    tmp := t.TempDir()
    path := tmp + "/test.json"

    data := `[{"id": "1", "name": "Alice"}, {"id": "2", "name": "Bob"}]`
    os.WriteFile(path, []byte(data), 0644)

    a := New(path)
    rows, err := a.Read(context.Background())
    if err != nil {
        t.Fatalf("Read failed: %v", err)
    }

    if len(rows) != 2 {
        t.Errorf("expected 2 rows, got %d", len(rows))
    }
}
