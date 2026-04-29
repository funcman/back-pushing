// internal/adapter/json/adapter.go
package json

import (
    "context"
    "encoding/json"
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
    data, err := os.ReadFile(a.path)
    if err != nil {
        return nil, err
    }

    var result []map[string]any
    if err := json.Unmarshal(data, &result); err != nil {
        return nil, err
    }

    return result, nil
}

func (a *Adapter) Close() error {
    return nil
}

func NewDataSource(path string) adapter.DataSource {
    return New(path)
}
