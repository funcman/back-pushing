# 推背图（Back-Pushing）Phase 2 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建数据接入层，支持从 JSON/CSV/SQL 数据源导入数据并映射到 Ontology

**Architecture:**
- 分离架构：adapter 包负责读取数据源，mapper 只负责映射转换
- 显式配置：所有字段映射在 YAML 中声明
- 环境变量分离：数据库凭据通过环境变量注入
- CLI 命令统一入口

**Tech Stack:** Go 1.21+, gopkg.in/yaml.v3, github.com/gocarina/gocsv, github.com/go-sql-driver/mysql, github.com/joho/godotenv

---

## 文件结构

```
back-pushing/
├── cmd/
│   └── server/
│       └── main.go              # 服务入口（Phase 1）
├── internal/
│   ├── adapter/
│   │   ├── adapter.go          # 数据源接口定义
│   │   ├── json/
│   │   │   └── adapter.go      # JSON 适配器
│   │   ├── csv/
│   │   │   └── adapter.go      # CSV 适配器
│   │   └── sql/
│   │       └── adapter.go      # SQL 适配器
│   ├── mapper/
│   │   ├── config.go           # 映射配置结构
│   │   └── mapper.go           # 映射引擎
│   └── cli/
│       └── import.go           # import 命令
├── mapping/                      # 映射配置文件示例
│   ├── person导入.yaml
│   ├── transaction导入.yaml
│   └── user导入.yaml
├── go.mod
└── Makefile
```

---

## Task 1: Adapter 接口定义

**Files:**
- Create: `internal/adapter/adapter.go`

- [ ] **Step 1: 定义数据源接口**
```go
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
```

- [ ] **Step 2: 运行验证**
```bash
go build ./internal/adapter/
```
Expected: 编译通过

- [ ] **Step 3: Commit**
```bash
git add -A && git commit -m "feat(adapter): define DataSource interface"
```

---

## Task 2: JSON 适配器

**Files:**
- Create: `internal/adapter/json/adapter.go`

- [ ] **Step 1: 实现 JSON 适配器**
```go
// internal/adapter/json/adapter.go
package json

import (
    "context"
    "encoding/json"
    "os"

    "back-pushing/internal/adapter"
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
```

- [ ] **Step 2: 编写测试**
```go
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
```

- [ ] **Step 3: 运行测试**
```bash
go test ./internal/adapter/json/... -v
```
Expected: PASS

- [ ] **Step 4: Commit**
```bash
git add -A && git commit -m "feat(adapter): add JSON adapter"
```

---

## Task 3: CSV 适配器

**Files:**
- Create: `internal/adapter/csv/adapter.go`
- Create: `internal/adapter/csv/adapter_test.go`

- [ ] **Step 1: 实现 CSV 适配器**
```go
// internal/adapter/csv/adapter.go
package csv

import (
    "context"
    "encoding/csv"
    "os"

    "back-pushing/internal/adapter"
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
        return nil, err
    }
    defer f.Close()

    reader := csv.NewReader(f)
    reader.Comma = a.delimiter
    reader.TrimLeadingSpace = true

    records, err := reader.ReadAll()
    if err != nil {
        return nil, err
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
```

- [ ] **Step 2: 编写测试**
```go
// internal/adapter/csv/adapter_test.go
package csv

import (
    "context"
    "os"
    "testing"
)

func TestAdapter_Read(t *testing.T) {
    tmp := t.TempDir()
    path := tmp + "/test.csv"

    data := "id,name,email\n1,Alice,alice@example.com\n2,Bob,bob@example.com"
    os.WriteFile(path, []byte(data), 0644)

    a := New(path)
    rows, err := a.Read(context.Background())
    if err != nil {
        t.Fatalf("Read failed: %v", err)
    }

    if len(rows) != 2 {
        t.Errorf("expected 2 rows, got %d", len(rows))
    }

    if rows[0]["name"] != "Alice" {
        t.Errorf("expected name Alice, got %v", rows[0]["name"])
    }
}
```

- [ ] **Step 3: 运行测试**
```bash
go test ./internal/adapter/csv/... -v
```
Expected: PASS

- [ ] **Step 4: Commit**
```bash
git add -A && git commit -m "feat(adapter): add CSV adapter with delimiter support"
```

---

## Task 4: SQL 适配器

**Files:**
- Create: `internal/adapter/sql/adapter.go`
- Create: `internal/adapter/sql/adapter_test.go`

- [ ] **Step 1: 实现 SQL 适配器**
```go
// internal/adapter/sql/adapter.go
package sql

import (
    "context"
    "database/sql"
    "fmt"

    _ "github.com/go-sql-driver/mysql"
    "back-pushing/internal/adapter"
)

type Adapter struct {
    query string
    db    *sql.DB
}

func New(dbURL string, query string) (*Adapter, error) {
    db, err := sql.Open("mysql", dbURL)
    if err != nil {
        return nil, err
    }
    return &Adapter{query: query, db: db}, nil
}

func (a *Adapter) Read(ctx context.Context) ([]map[string]any, error) {
    rows, err := a.db.QueryContext(ctx, a.query)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    cols, err := rows.Columns()
    if err != nil {
        return nil, err
    }

    var result []map[string]any
    for rows.Next() {
        values := make([]any, len(cols))
        ptrs := make([]any, len(cols))
        for i := range values {
            ptrs[i] = &values[i]
        }

        if err := rows.Scan(ptrs...); err != nil {
            return nil, err
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
```

- [ ] **Step 2: 编写测试**
```go
// internal/adapter/sql/adapter_test.go
package sql

import (
    "context"
    "testing"

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
        t.Fatalf("Read failed: %v", err)
    }

    if len(rows) == 0 {
        t.Error("expected at least one row")
    }
}
```

- [ ] **Step 3: 运行测试**
```bash
go test ./internal/adapter/sql/... -v
```
Expected: PASS (或 SKIP 如果 MySQL 不可用)

- [ ] **Step 4: Commit**
```bash
git add -A && git commit -m "feat(adapter): add SQL adapter"
```

---

## Task 5: Mapper 配置解析

**Files:**
- Create: `internal/mapper/config.go`
- Create: `internal/mapper/config_test.go`

- [ ] **Step 1: 定义配置结构**
```go
// internal/mapper/config.go
package mapper

// MappingConfig 映射配置
type MappingConfig struct {
    Source struct {
        Type  string `yaml:"type"`  // json | csv | sql
        Path  string `yaml:"path"`
        Query string `yaml:"query"`
    } `yaml:"source"`

    Env map[string]string `yaml:"env"`

    Target struct {
        ObjectType string `yaml:"object_type"`
    } `yaml:"target"`

    Fields []FieldMapping `yaml:"fields"`
}

// FieldMapping 字段映射
type FieldMapping struct {
    Source string `yaml:"source"`
    Target string `yaml:"target"`
    Type   string `yaml:"type"`  // string | int | float | bool | datetime
    Link   string `yaml:"link,omitempty"`
}
```

- [ ] **Step 2: 实现配置加载**
```go
// internal/mapper/config.go

func LoadConfig(path string) (*MappingConfig, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }

    var cfg MappingConfig
    if err := yaml.Unmarshal(data, &cfg); err != nil {
        return nil, err
    }

    return &cfg, nil
}
```

- [ ] **Step 3: 编写测试**
```go
// internal/mapper/config_test.go
package mapper

import (
    "os"
    "testing"
)

func TestLoadConfig(t *testing.T) {
    tmp := t.TempDir()
    path := tmp + "/mapping.yaml"

    yaml := `
source:
  type: csv
  path: ./data/persons.csv

target:
  object_type: Person

fields:
  - source: id
    target: id
    type: string
  - source: name
    target: name
    type: string
`
    os.WriteFile(path, []byte(yaml), 0644)

    cfg, err := LoadConfig(path)
    if err != nil {
        t.Fatalf("LoadConfig failed: %v", err)
    }

    if cfg.Source.Type != "csv" {
        t.Errorf("expected type csv, got %s", cfg.Source.Type)
    }

    if cfg.Target.ObjectType != "Person" {
        t.Errorf("expected object_type Person, got %s", cfg.Target.ObjectType)
    }

    if len(cfg.Fields) != 2 {
        t.Errorf("expected 2 fields, got %d", len(cfg.Fields))
    }
}
```

- [ ] **Step 4: 运行测试**
```bash
go test ./internal/mapper/... -v
```
Expected: PASS

- [ ] **Step 5: Commit**
```bash
git add -A && git commit -m "feat(mapper): add config parsing"
```

---

## Task 6: Mapper 映射引擎

**Files:**
- Modify: `internal/mapper/config.go`
- Create: `internal/mapper/mapper.go`
- Create: `internal/mapper/mapper_test.go`

- [ ] **Step 1: 实现映射引擎**
```go
// internal/mapper/mapper.go
package mapper

import (
    "context"
    "fmt"
    "time"

    "back-pushing/internal/adapter"
)

type Object struct {
    ID    string
    Type  string
    Data  map[string]any
    Links []Link
}

type Link struct {
    FromID  string
    ToID    string
    LinkType string
    Props   map[string]any
}

type Mapper struct {
    config *MappingConfig
}

func New(cfg *MappingConfig) *Mapper {
    return &Mapper{config: cfg}
}

func (m *Mapper) Map(ctx context.Context, source adapter.DataSource) ([]Object, error) {
    rows, err := source.Read(ctx)
    if err != nil {
        return nil, fmt.Errorf("read source: %w", err)
    }

    var objects []Object
    for _, row := range rows {
        obj := Object{
            ID:   m.getField(row, "id"),
            Type: m.config.Target.ObjectType,
            Data: make(map[string]any),
        }

        for _, f := range m.config.Fields {
            val := row[f.Source]
            converted, err := m.convert(val, f.Type)
            if err != nil {
                continue
            }
            obj.Data[f.Target] = converted

            if f.Link != "" {
                obj.Links = append(obj.Links, Link{
                    FromID:   obj.ID,
                    ToID:     fmt.Sprintf("%v", val),
                    LinkType: f.Link,
                })
            }
        }

        objects = append(objects, obj)
    }

    return objects, nil
}

func (m *Mapper) getField(row map[string]any, key string) string {
    if val, ok := row[key]; ok {
        return fmt.Sprintf("%v", val)
    }
    return ""
}

func (m *Mapper) convert(val any, typ string) (any, error) {
    if val == nil {
        return nil, nil
    }

    switch typ {
    case "string":
        return fmt.Sprintf("%v", val), nil
    case "int":
        switch v := val.(type) {
        case float64:
            return int(v), nil
        case string:
            var i int
            fmt.Sscanf(v, "%d", &i)
            return i, nil
        }
    case "float":
        switch v := val.(type) {
        case float64:
            return v, nil
        case string:
            var f float64
            fmt.Sscanf(v, "%f", &f)
            return f, nil
        }
    case "datetime":
        switch v := val.(type) {
        case string:
            if t, err := time.Parse(time.RFC3339, v); err == nil {
                return t, nil
            }
            return v, nil
        }
    }

    return val, nil
}
```

- [ ] **Step 2: 编写测试**
```go
// internal/mapper/mapper_test.go
package mapper

import (
    "context"
    "testing"

    "back-pushing/internal/adapter"
)

type mockSource struct {
    rows []map[string]any
}

func (m *mockSource) Read(ctx context.Context) ([]map[string]any, error) {
    return m.rows, nil
}

func (m *mockSource) Close() error {
    return nil
}

func TestMapper_Map(t *testing.T) {
    cfg := &MappingConfig{
        Target: struct{ ObjectType string }{ObjectType: "Person"},
        Fields: []FieldMapping{
            {Source: "id", Target: "id", Type: "string"},
            {Source: "name", Target: "name", Type: "string"},
            {Source: "org_id", Target: "organization_id", Type: "string", Link: "works_at"},
        },
    }

    mapper := New(cfg)
    source := &mockSource{
        rows: []map[string]any{
            {"id": "1", "name": "Alice", "org_id": "org1"},
            {"id": "2", "name": "Bob", "org_id": "org2"},
        },
    }

    objects, err := mapper.Map(context.Background(), source)
    if err != nil {
        t.Fatalf("Map failed: %v", err)
    }

    if len(objects) != 2 {
        t.Errorf("expected 2 objects, got %d", len(objects))
    }

    if objects[0].Data["name"] != "Alice" {
        t.Errorf("expected name Alice, got %v", objects[0].Data["name"])
    }

    if len(objects[0].Links) != 1 {
        t.Errorf("expected 1 link, got %d", len(objects[0].Links))
    }
}
```

- [ ] **Step 3: 运行测试**
```bash
go test ./internal/mapper/... -v
```
Expected: PASS

- [ ] **Step 4: Commit**
```bash
git add -A && git commit -m "feat(mapper): add mapping engine"
```

---

## Task 7: CLI Import 命令

**Files:**
- Create: `internal/cli/import.go`
- Create: `cmd/cli/main.go`

- [ ] **Step 1: 实现 import 命令**
```go
// internal/cli/import.go
package cli

import (
    "context"
    "fmt"
    "log"
    "os"

    "github.com/joho/godotenv"
    "back-pushing/internal/adapter"
    "back-pushing/internal/adapter/csv"
    "back-pushing/internal/adapter/json"
    sqladapter "back-pushing/internal/adapter/sql"
    "back-pushing/internal/mapper"
    "back-pushing/internal/storage/memory"
)

func Import(mappingPath string, envPath string) error {
    // 1. 加载环境变量
    if envPath != "" {
        if err := godotenv.Load(envPath); err != nil {
            log.Printf("Warning: .env file not loaded: %v", err)
        }
    }

    // 2. 加载映射配置
    cfg, err := mapper.LoadConfig(mappingPath)
    if err != nil {
        return fmt.Errorf("load config: %w", err)
    }

    // 3. 初始化数据源
    var source adapter.DataSource
    switch cfg.Source.Type {
    case "json":
        source = json.NewDataSource(cfg.Source.Path)
    case "csv":
        source = csv.NewDataSource(cfg.Source.Path)
    case "sql":
        dbURL := os.Getenv(cfg.Env["DB_URL"])
        source, err = sqladapter.NewDataSource(dbURL, cfg.Source.Query)
        if err != nil {
            return fmt.Errorf("create SQL adapter: %w", err)
        }
    default:
        return fmt.Errorf("unsupported source type: %s", cfg.Source.Type)
    }
    defer source.Close()

    // 4. 执行映射
    m := mapper.New(cfg)
    objects, err := m.Map(context.Background(), source)
    if err != nil {
        return fmt.Errorf("map objects: %w", err)
    }

    // 5. 写入存储
    store := memory.NewObjectStore()
    graph := memory.NewGraphStore()
    for _, obj := range objects {
        if err := store.Create(context.Background(), obj.Type, obj.ID, obj.Data); err != nil {
            log.Printf("Warning: create object %s/%s: %v", obj.Type, obj.ID, err)
        }
        for _, link := range obj.Links {
            graph.AddEdge(context.Background(), link.LinkType, link.FromID, link.ToID, link.Props)
        }
    }

    log.Printf("Imported %d objects", len(objects))
    return nil
}
```

- [ ] **Step 2: 创建 CLI 入口**
```go
// cmd/cli/main.go
package main

import (
    "flag"
    "log"

    "back-pushing/internal/cli"
)

func main() {
    mapping := flag.String("mapping", "", "Mapping config file path")
    env := flag.String("env", "", "Environment file path")
    flag.Parse()

    if *mapping == "" {
        log.Fatal("--mapping is required")
    }

    if err := cli.Import(*mapping, *env); err != nil {
        log.Fatalf("Import failed: %v", err)
    }
}
```

- [ ] **Step 3: 验证编译**
```bash
go build -o bin/cli ./cmd/cli
```

- [ ] **Step 4: Commit**
```bash
git add -A && git commit -m "feat(cli): add import command"
```

---

## Task 8: 集成测试

**Files:**
- Create: `mapping/person导入.yaml`
- Create: `mapping/transaction导入.yaml`

- [ ] **Step 1: 创建示例映射配置**
```yaml
# mapping/person导入.yaml
source:
  type: csv
  path: ./testdata/persons.csv

target:
  object_type: Person

fields:
  - source: id
    target: id
    type: string
  - source: name
    target: name
    type: string
  - source: email
    target: email
    type: string
```

- [ ] **Step 2: 创建测试数据**
```csv
# testdata/persons.csv
id,name,email
1,Alice,alice@example.com
2,Bob,bob@example.com
3,Charlie,charlie@example.com
```

- [ ] **Step 3: 运行集成测试**
```bash
go run ./cmd/cli/main.go --mapping ./mapping/person导入.yaml
```

Expected: 输出 "Imported 3 objects"

- [ ] **Step 4: Commit**
```bash
git add -A && git commit -m "test: add integration tests for import"
```

---

## 自检清单

- [x] **Spec覆盖检查**：
  - JSON 适配器 → Task 2
  - CSV 适配器 → Task 3
  - SQL 适配器 → Task 4
  - Mapper 配置解析 → Task 5
  - Mapper 映射引擎 → Task 6
  - CLI Import 命令 → Task 7
  - 集成测试 → Task 8

- [x] **Placeholder扫描**：无 TBD/TODO

- [x] **类型一致性**：
  - DataSource 接口统一：Read/Close
  - 各适配器实现一致
  - Mapper 配置结构清晰
