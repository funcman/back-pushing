// internal/adapter/sql/adapter.go
package sql

import (
    "context"
    "database/sql"
    "fmt"

    _ "github.com/go-sql-driver/mysql"
    "github.com/funcman/back-pushing/internal/adapter"
)

type Adapter struct {
    query string
    db    *sql.DB
}

func New(dbURL string, query string) (*Adapter, error) {
    db, err := sql.Open("mysql", dbURL)
    if err != nil {
        return nil, fmt.Errorf("sql adapter open: %w", err)
    }
    return &Adapter{query: query, db: db}, nil
}

func (a *Adapter) Read(ctx context.Context) ([]map[string]any, error) {
    rows, err := a.db.QueryContext(ctx, a.query)
    if err != nil {
        return nil, fmt.Errorf("sql adapter query: %w", err)
    }
    defer rows.Close()

    cols, err := rows.Columns()
    if err != nil {
        return nil, fmt.Errorf("sql adapter columns: %w", err)
    }

    var result []map[string]any
    for rows.Next() {
        values := make([]any, len(cols))
        ptrs := make([]any, len(cols))
        for i := range values {
            ptrs[i] = &values[i]
        }

        if err := rows.Scan(ptrs...); err != nil {
            return nil, fmt.Errorf("sql adapter scan: %w", err)
        }

        record := make(map[string]any)
        for i, col := range cols {
            val := values[i]
            if b, ok := val.([]byte); ok {
                record[col] = string(b)
            } else {
                record[col] = val
            }
        }
        result = append(result, record)
    }

    return result, rows.Err()
}

func (a *Adapter) Close() error {
    return a.db.Close()
}

func NewDataSource(dbURL string, query string) (adapter.DataSource, error) {
    return New(dbURL, query)
}