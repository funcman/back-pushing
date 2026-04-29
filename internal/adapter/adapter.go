// internal/adapter/adapter.go
package adapter

import "context"

// DataSource 数据源接口
type DataSource interface {
    Read(ctx context.Context) ([]map[string]any, error)
    Close() error
}

// Type 数据源类型
type Type string

const (
    TypeJSON Type = "json"
    TypeCSV  Type = "csv"
    TypeSQL  Type = "sql"
)
