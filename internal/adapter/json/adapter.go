// internal/adapter/json/adapter.go
package json

import (
    "context"
    "encoding/json"
    "fmt"
    "os"

    "github.com/funcman/back-pushing/internal/adapter"
)

type Adapter struct {
    path string
}

func New(path string) *Adapter {
    return &Adapter{path: path}
}

func (a *Adapter) Read(ctx context.Context) ([]map[string]any, error) {
    select {
    case <-ctx.Done():
        return nil, fmt.Errorf("json adapter read: %w", ctx.Err())
    default:
    }

    data, err := os.ReadFile(a.path)
    if err != nil {
        return nil, fmt.Errorf("json adapter read: %w", err)
    }

    var result []map[string]any
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, fmt.Errorf("json adapter read: %w", err)
    }

    return result, nil
}

func (a *Adapter) Close() error {
    return nil
}

func NewDataSource(path string) adapter.DataSource {
    return New(path)
}
