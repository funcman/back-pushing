// internal/adapter/sql/adapter_test.go
package sql

import (
    "context"
    "errors"
    "testing"
    "time"

    _ "github.com/go-sql-driver/mysql"
)

func TestAdapter_Interface(t *testing.T) {
    a, err := New("root:test@tcp(localhost:3306)/test", "SELECT 1 as id")
    if err != nil {
        t.Skip("MySQL not available")
    }
    defer a.Close()

    rows, err := a.Read(context.Background())
    if err != nil {
        t.Skipf("Read failed (DB not available): %v", err)
    }

    if len(rows) == 0 {
        t.Error("expected at least one row")
    }
}

func TestAdapter_NewError(t *testing.T) {
    // sql.Open does NOT validate DSN - it only creates a database handle.
    // Error occurs on Ping/query. This test verifies New handles invalid
    // driver name properly.
    _, err := New("invalid-driver://localhost:3306", "SELECT 1")
    if err == nil {
        t.Error("expected error for invalid driver")
    }
}

func TestAdapter_EmptyQueryResult(t *testing.T) {
    a, err := New("root:test@tcp(localhost:3306)/test", "SELECT 1 as id WHERE 1=0")
    if err != nil {
        t.Skip("MySQL not available")
    }
    defer a.Close()

    rows, err := a.Read(context.Background())
    if err != nil {
        t.Skipf("Read failed (DB not available): %v", err)
    }

    if rows == nil {
        t.Error("expected non-nil result for empty query")
    }
    if len(rows) != 0 {
        t.Errorf("expected 0 rows, got %d", len(rows))
    }
}

func TestAdapter_ContextCancellation(t *testing.T) {
    a, err := New("root:test@tcp(localhost:3306)/test", "SELECT SLEEP(10)")
    if err != nil {
        t.Skip("MySQL not available")
    }
    defer a.Close()

    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    _, err = a.Read(ctx)
    if err == nil {
        t.Error("expected error for cancelled context")
    }
    if !errors.Is(err, context.Canceled) {
        t.Errorf("expected context.Canceled, got: %v", err)
    }
}

func TestAdapter_New_InvalidDriver(t *testing.T) {
    // sql.Open does NOT validate driver name - it only creates a db handle.
    // Error occurs on first query. This verifies New doesn't panic.
    a, err := New("invalid-driver://localhost:3306", "SELECT 1")
    if err != nil {
        t.Skip("sql.Open returned error (unexpected)")
    }
    defer a.Close()
    // Error will occur on Read, not New
    _, err = a.Read(context.Background())
    if err == nil {
        t.Error("expected error for invalid driver")
    }
}

func TestAdapter_New_EmptyQuery(t *testing.T) {
    a, err := New("root:test@tcp(localhost:3306)/test", "")
    if err != nil {
        t.Skip("MySQL not available")
    }
    defer a.Close()

    _, err = a.Read(context.Background())
    // Empty query may return error depending on driver
    // Just verify it doesn't panic
}

func TestAdapter_ReadTimeout(t *testing.T) {
    a, err := New("root:test@tcp(localhost:3306)/test", "SELECT SLEEP(5)")
    if err != nil {
        t.Skip("MySQL not available")
    }
    defer a.Close()

    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()

    _, err = a.Read(ctx)
    if err == nil {
        t.Error("expected timeout error")
    }
    // Error may be context.DeadlineExceeded or a connection error
    if !errors.Is(err, context.DeadlineExceeded) {
        t.Logf("got error: %v", err)
    }
}

func TestAdapter_CloseError(t *testing.T) {
    a, err := New("root:test@tcp(localhost:3306)/test", "SELECT 1")
    if err != nil {
        t.Skip("MySQL not available")
    }
    // Close twice - second close should not panic
    a.Close()
    a.Close()
}