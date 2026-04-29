// internal/adapter/csv/adapter.go
package csv

import (
    "context"
    "encoding/csv"
    "fmt"
    "os"

    "github.com/funcman/back-pushing/internal/adapter"
)

type Adapter struct {
    path      string
    delimiter rune
    hasHeader bool
}

type Option func(*Adapter)

func WithDelimiter(d rune) Option {
    return func(a *Adapter) { a.delimiter = d }
}

func WithHeader() Option {
    return func(a *Adapter) { a.hasHeader = true }
}

func New(path string, opts ...Option) *Adapter {
    a := &Adapter{
        path:      path,
        delimiter: ',',
        hasHeader: true,
    }
    for _, opt := range opts {
        opt(a)
    }
    return a
}

func (a *Adapter) Read(ctx context.Context) ([]map[string]any, error) {
    f, err := os.Open(a.path)
    if err != nil {
        return nil, fmt.Errorf("csv adapter open: %w", err)
    }
    defer f.Close()

    reader := csv.NewReader(f)
    reader.Comma = a.delimiter
    reader.TrimLeadingSpace = true

    records, err := reader.ReadAll()
    if err != nil {
        return nil, fmt.Errorf("csv adapter read: %w", err)
    }

    if len(records) == 0 {
        return nil, nil
    }

    var result []map[string]any
    var headers []string

    if a.hasHeader {
        headers = records[0]
        for _, row := range records[1:] {
            record := make(map[string]any)
            for i, val := range row {
                if i < len(headers) {
                    record[headers[i]] = val
                }
            }
            result = append(result, record)
        }
    } else {
        for _, row := range records {
            record := make(map[string]any)
            for i, val := range row {
                record[fmt.Sprintf("col%d", i)] = val
            }
            result = append(result, record)
        }
    }

    return result, nil
}

func (a *Adapter) Close() error {
    return nil
}

func NewDataSource(path string, opts ...Option) adapter.DataSource {
    return New(path, opts...)
}