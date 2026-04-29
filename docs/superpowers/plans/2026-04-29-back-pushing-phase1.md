# 推背图（Back-Pushing）Phase 1 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建最小可用核心——Ontology YAML解析、内存对象存储、基础图遍历引擎

**Architecture:**
- Go语言实现，模块化设计
- YAML驱动Ontology定义，运行时解析注册
- 内存存储层，图结构用邻接表实现
- API层提供GraphQL查询能力和Action调用

**Tech Stack:** Go 1.21+, yaml.v3, gqlgen (GraphQL), go-redis/leveldb (可选)

---

## 文件结构

```
back-pushing/
├── cmd/
│   └── server/
│       └── main.go              # 服务入口
├── internal/
│   ├── ontology/
│   │   ├── parser.go            # YAML解析器
│   │   ├── registry.go         # 对象类型注册表
│   │   ├── link.go             # Link关系管理
│   │   ├── path.go             # Path路径解析
│   │   ├── action.go           # Action调度器
│   │   ├── classification.go   # 分类标签引擎
│   │   └── types.go            # 类型定义
│   ├── storage/
│   │   ├── memory/
│   │   │   ├── graph.go        # 内存图存储（邻接表）
│   │   │   ├── object.go       # 对象存储
│   │   │   └── index.go        # 索引
│   │   └── interfaces.go       # 存储接口定义
│   ├── engine/
│   │   └── graph/
│   │       ├── traverse.go     # 图遍历
│   │       └── search.go       # 图搜索
│   ├── api/
│   │   └── graphql/
│   │       ├── schema.graphql  # GraphQL schema
│   │       └── resolver.go     # GraphQL resolver
│   └── action/
│       └── context.go         # Action上下文
├── ontology/                    # 用户定义的YAML文件
│   ├── person.yaml
│   └── links.yaml
├── go.mod
├── go.sum
└── Makefile
```

---

## Task 1: 项目初始化

**Files:**
- Create: `go.mod`
- Create: `Makefile`
- Create: `cmd/server/main.go`

- [ ] **Step 1: 初始化Go模块**
```bash
go mod init github.com/funcman/back-pushing
```

- [ ] **Step 2: 创建Makefile**
```makefile
.PHONY: build run test clean

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

clean:
	rm -rf bin/
```

- [ ] **Step 3: 创建main.go骨架**
```go
package main

import (
	"fmt"
	"log"
)

func main() {
	fmt.Println("Back-Pushing Server Starting...")
	log.Println("Initializing Ontology...")
}
```

- [ ] **Step 4: 运行验证**
```bash
go build -o bin/server ./cmd/server && ./bin/server
```
Expected: 输出 "Back-Pushing Server Starting..."

- [ ] **Step 5: Commit**
```bash
git add -A && git commit -m "feat: project scaffolding"
```

---

## Task 2: Ontology解析器

**Files:**
- Create: `internal/ontology/types.go`
- Create: `internal/ontology/parser.go`
- Create: `internal/ontology/parser_test.go`

- [ ] **Step 1: 定义Ontology类型**
```go
// internal/ontology/types.go
package ontology

type ObjectType struct {
	Name        string
	Description string
	Type        string // "entity", "event", "concept"
	Properties  map[string]Property
	Links       map[string]LinkDef
	Actions     []ActionDef
	Paths       map[string]PathDef
}

type Property struct {
	Type     string
	Primary  bool
	Indexed  bool
	Unique   bool
	Computed bool
	Source   string
}

type LinkDef struct {
	Target  string
	Through string
	Reverse string
}

type ActionDef struct {
	Name        string
	Description string
	Handler     string
	Args        map[string]string
}

type PathDef struct {
	Description string
	Steps       []string
}

type Ontology struct {
	ObjectTypes map[string]ObjectType
	Links       map[string]Link
	Paths       map[string]PathDef
	Classification *Classification
}

type Link struct {
	Description string
	Source      string
	Target      string
	Type        string
	Properties  map[string]Property
	Actions     []ActionDef
}

type Classification struct {
	Levels     []string
	DataHandling map[string]DataHandling
	ObjectTags  map[string]ObjectTag
}

type DataHandling struct {
	Description string
	Actions     []string
}

type ObjectTag struct {
	Sensitivity string
	Handling    string
}
```

- [ ] **Step 2: 实现YAML解析器**
```go
// internal/ontology/parser.go
package ontology

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func ParseOntology(dir string) (*Ontology, error) {
	ont := &Ontology{
		ObjectTypes: make(map[string]ObjectType),
		Links:       make(map[string]Link),
		Paths:       make(map[string]PathDef),
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read ontology dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", entry.Name(), err)
		}

		var doc map[string]any
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}

		if objectTypes, ok := doc["object_types"].(map[string]any); ok {
			for name, v := range objectTypes {
				obj := parseObjectType(name, v.(map[string]any))
				ont.ObjectTypes[name] = obj
			}
		}

		if links, ok := doc["links"].(map[string]any); ok {
			for name, v := range links {
				link := parseLink(name, v.(map[string]any))
				ont.Links[name] = link
			}
		}

		if paths, ok := doc["paths"].(map[string]any); ok {
			for name, v := range paths {
				path := parsePath(name, v.(map[string]any))
				ont.Paths[name] = path
			}
		}
	}

	return ont, nil
}

func parseObjectType(name string, data map[string]any) ObjectType {
	obj := ObjectType{
		Name:        name,
		Description: getString(data, "description"),
		Type:        getString(data, "type", "entity"),
		Properties:  make(map[string]Property),
		Links:       make(map[string]LinkDef),
		Actions:     []ActionDef{},
		Paths:       make(map[string]PathDef),
	}

	if props, ok := data["properties"].(map[string]any); ok {
		for pname, pv := range props {
			obj.Properties[pname] = parseProperty(pv.(map[string]any))
		}
	}

	if links, ok := data["links"].(map[string]any); ok {
		for lname, lv := range links {
			obj.Links[lname] = parseLinkDef(lv.(map[string]any))
		}
	}

	return obj
}

func parseProperty(data map[string]any) Property {
	return Property{
		Type:     getString(data, "type"),
		Primary:  getBool(data, "primary"),
		Indexed:  getBool(data, "indexed"),
		Unique:   getBool(data, "unique"),
		Computed: getBool(data, "computed"),
		Source:   getString(data, "source"),
	}
}

func parseLinkDef(data map[string]any) LinkDef {
	return LinkDef{
		Target:  getString(data, "target"),
		Through: getString(data, "through"),
		Reverse: getString(data, "reverse"),
	}
}

func parseLink(name string, data map[string]any) Link {
	return Link{
		Description: getString(data, "description"),
		Source:      getString(data, "source"),
		Target:      getString(data, "target"),
		Type:        getString(data, "type", "many-to-many"),
	}
}

func parsePath(name string, data map[string]any) PathDef {
	return PathDef{
		Description: getString(data, "description"),
		Steps:       getStringSlice(data, "steps"),
	}
}

func getString(m map[string]any, key string, defaults ...string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	if len(defaults) > 0 {
		return defaults[0]
	}
	return ""
}

func getBool(m map[string]any, key string) bool {
	if v, ok := m[key].(bool); ok {
		return v
	}
	return false
}

func getStringSlice(m map[string]any, key string) []string {
	if v, ok := m[key].([]any); ok {
		result := make([]string, len(v))
		for i, e := range v {
			result[i] = e.(string)
		}
		return result
	}
	return nil
}
```

- [ ] **Step 3: 编写测试**
```go
// internal/ontology/parser_test.go
package ontology

import (
	"os"
	"testing"
)

func TestParseOntology(t *testing.T) {
	dir := t.TempDir()

	personYAML := `
object_types:
  Person:
    description: Test person
    type: entity
    properties:
      id:
        type: string
        primary: true
      name:
        type: string
        indexed: true
`
	os.WriteFile(dir+"/person.yaml", []byte(personYAML), 0644)

	ont, err := ParseOntology(dir)
	if err != nil {
		t.Fatalf("ParseOntology failed: %v", err)
	}

	if len(ont.ObjectTypes) != 1 {
		t.Errorf("expected 1 object type, got %d", len(ont.ObjectTypes))
	}

	person, ok := ont.ObjectTypes["Person"]
	if !ok {
		t.Fatal("Person type not found")
	}

	if person.Description != "Test person" {
		t.Errorf("expected description 'Test person', got '%s'", person.Description)
	}

	if !person.Properties["id"].Primary {
		t.Error("id should be primary")
	}

	if !person.Properties["name"].Indexed {
		t.Error("name should be indexed")
	}
}
```

- [ ] **Step 4: 运行测试**
```bash
go test ./internal/ontology/ -v
```
Expected: PASS

- [ ] **Step 5: Commit**
```bash
git add -A && git commit -m "feat(ontology): add YAML parser and types"
```

---

## Task 3: 对象注册表

**Files:**
- Create: `internal/ontology/registry.go`
- Create: `internal/ontology/registry_test.go`

- [ ] **Step 1: 实现注册表**
```go
// internal/ontology/registry.go
package ontology

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu          sync.RWMutex
	objectTypes map[string]*ObjectType
	links       map[string]*Link
	paths       map[string]*PathDef
}

func NewRegistry() *Registry {
	return &Registry{
		objectTypes: make(map[string]*ObjectType),
		links:       make(map[string]*Link),
		paths:       make(map[string]*PathDef),
	}
}

func (r *Registry) RegisterOntology(ont *Ontology) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for name, obj := range ont.ObjectTypes {
		if _, exists := r.objectTypes[name]; exists {
			return fmt.Errorf("object type %s already registered", name)
		}
		objCopy := obj
		r.objectTypes[name] = &objCopy
	}

	for name, link := range ont.Links {
		if _, exists := r.links[name]; exists {
			return fmt.Errorf("link %s already registered", name)
		}
		linkCopy := link
		r.links[name] = &linkCopy
	}

	for name, path := range ont.Paths {
		pathCopy := path
		r.paths[name] = &pathCopy
	}

	return nil
}

func (r *Registry) GetObjectType(name string) (*ObjectType, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	obj, ok := r.objectTypes[name]
	if !ok {
		return nil, fmt.Errorf("object type %s not found", name)
	}
	return obj, nil
}

func (r *Registry) GetLink(name string) (*Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	link, ok := r.links[name]
	if !ok {
		return nil, fmt.Errorf("link %s not found", name)
	}
	return link, nil
}

func (r *Registry) GetPath(name string) (*PathDef, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	path, ok := r.paths[name]
	if !ok {
		return nil, fmt.Errorf("path %s not found", name)
	}
	return path, nil
}

func (r *Registry) ListObjectTypes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.objectTypes))
	for name := range r.objectTypes {
		names = append(names, name)
	}
	return names
}
```

- [ ] **Step 2: 编写测试**
```go
// internal/ontology/registry_test.go
package ontology

import (
	"testing"
)

func TestRegistry(t *testing.T) {
	ont := &Ontology{
		ObjectTypes: map[string]ObjectType{
			"Person": {
				Name: "Person",
				Properties: map[string]Property{
					"id": {Type: "string", Primary: true},
				},
			},
		},
		Links: map[string]Link{
			"knows": {
				Source: "Person",
				Target: "Person",
			},
		},
	}

	r := NewRegistry()

	if err := r.RegisterOntology(ont); err != nil {
		t.Fatalf("RegisterOntology failed: %v", err)
	}

	person, err := r.GetObjectType("Person")
	if err != nil {
		t.Fatalf("GetObjectType failed: %v", err)
	}
	if person.Name != "Person" {
		t.Errorf("expected Person, got %s", person.Name)
	}

	link, err := r.GetLink("knows")
	if err != nil {
		t.Fatalf("GetLink failed: %v", err)
	}
	if link.Source != "Person" {
		t.Errorf("expected source Person, got %s", link.Source)
	}

	_, err = r.GetObjectType("NonExistent")
	if err == nil {
		t.Error("expected error for non-existent type")
	}
}
```

- [ ] **Step 3: 运行测试**
```bash
go test ./internal/ontology/ -v -run TestRegistry
```
Expected: PASS

- [ ] **Step 4: Commit**
```bash
git add -A && git commit -m "feat(ontology): add registry for object types and links"
```

---

## Task 4: 内存图存储

**Files:**
- Create: `internal/storage/interfaces.go`
- Create: `internal/storage/memory/object.go`
- Create: `internal/storage/memory/graph.go`
- Create: `internal/storage/memory/object_test.go`
- Create: `internal/storage/memory/graph_test.go`

- [ ] **Step 1: 定义存储接口**
```go
// internal/storage/interfaces.go
package storage

import "context"

type ObjectStore interface {
	Create(ctx context.Context, objType string, id string, data map[string]any) error
	Get(ctx context.Context, objType string, id string) (map[string]any, error)
	Update(ctx context.Context, objType string, id string, data map[string]any) error
	Delete(ctx context.Context, objType string, id string) error
	List(ctx context.Context, objType string, filter map[string]any) ([]map[string]any, error)
}

type GraphStore interface {
	AddEdge(ctx context.Context, linkType string, fromID, toID string, props map[string]any) error
	RemoveEdge(ctx context.Context, linkType string, fromID, toID string) error
	GetEdges(ctx context.Context, linkType string, nodeID string) ([]Edge, error)
	Traverse(ctx context.Context, startID string, linkTypes []string, depth int) ([]Path, error)
}

type Edge struct {
	From  string
	To    string
	Props map[string]any
}

type Path struct {
	Nodes []string
	Edges []Edge
}

var (
	ErrNotFound      = &StorageError{msg: "object not found"}
	ErrAlreadyExists = &StorageError{msg: "object already exists"}
)

type StorageError struct {
	msg string
}

func (e *StorageError) Error() string { return e.msg }
```

- [ ] **Step 2: 实现对象存储**
```go
// internal/storage/memory/object.go
package memory

import (
	"context"
	"sync"
)

type ObjectStore struct {
	mu    sync.RWMutex
	store map[string]map[string]map[string]any // objType -> id -> data
}

func NewObjectStore() *ObjectStore {
	return &ObjectStore{
		store: make(map[string]map[string]map[string]any),
	}
}

func (s *ObjectStore) ensureType(objType string) map[string]map[string]any {
	if _, ok := s.store[objType]; !ok {
		s.store[objType] = make(map[string]map[string]any)
	}
	return s.store[objType]
}

func (s *ObjectStore) Create(ctx context.Context, objType string, id string, data map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	objs := s.ensureType(objType)
	if _, exists := objs[id]; exists {
		return ErrAlreadyExists
	}

	objs[id] = data
	return nil
}

func (s *ObjectStore) Get(ctx context.Context, objType string, id string) (map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objs, ok := s.store[objType]
	if !ok {
		return nil, ErrNotFound
	}

	obj, exists := objs[id]
	if !exists {
		return nil, ErrNotFound
	}

	result := make(map[string]any)
	for k, v := range obj {
		result[k] = v
	}
	return result, nil
}

func (s *ObjectStore) Update(ctx context.Context, objType string, id string, data map[string]any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	objs, ok := s.store[objType]
	if !ok {
		return ErrNotFound
	}

	if _, exists := objs[id]; !exists {
		return ErrNotFound
	}

	for k, v := range data {
		objs[id][k] = v
	}
	return nil
}

func (s *ObjectStore) Delete(ctx context.Context, objType string, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	objs, ok := s.store[objType]
	if !ok {
		return ErrNotFound
	}

	if _, exists := objs[id]; !exists {
		return ErrNotFound
	}

	delete(objs, id)
	return nil
}

func (s *ObjectStore) List(ctx context.Context, objType string, filter map[string]any) ([]map[string]any, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	objs, ok := s.store[objType]
	if !ok {
		return nil, nil
	}

	var result []map[string]any
	for _, data := range objs {
		if len(filter) == 0 {
			result = append(result, data)
			continue
		}
		if matchFilter(data, filter) {
			result = append(result, data)
		}
	}

	return result, nil
}

func matchFilter(data, filter map[string]any) bool {
	for k, v := range filter {
		if data[k] != v {
			return false
		}
	}
	return true
}
```

- [ ] **Step 3: 实现图存储（邻接表）**
```go
// internal/storage/memory/graph.go
package memory

import (
	"context"
	"sync"
)

type GraphStore struct {
	mu     sync.RWMutex
	edges  map[string]map[string][]Edge // linkType -> fromID -> [edges]
	revIdx map[string]map[string][]Edge // linkType -> toID -> [edges] (反向索引)
}

func NewGraphStore() *GraphStore {
	return &GraphStore{
		edges:  make(map[string]map[string][]Edge),
		revIdx: make(map[string]map[string][]Edge),
	}
}

func (g *GraphStore) AddEdge(ctx context.Context, linkType string, fromID, toID string, props map[string]any) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.edges[linkType] == nil {
		g.edges[linkType] = make(map[string][]Edge)
	}
	if g.revIdx[linkType] == nil {
		g.revIdx[linkType] = make(map[string][]Edge)
	}

	edge := Edge{From: fromID, To: toID, Props: props}
	g.edges[linkType][fromID] = append(g.edges[linkType][fromID], edge)
	g.revIdx[linkType][toID] = append(g.revIdx[linkType][toID], edge)

	return nil
}

func (g *GraphStore) RemoveEdge(ctx context.Context, linkType string, fromID, toID string) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	if edges, ok := g.edges[linkType]; ok {
		g.edges[linkType][fromID] = removeEdgeFromSlice(edges[fromID], toID)
	}
	if revEdges, ok := g.revIdx[linkType]; ok {
		g.revIdx[linkType][toID] = removeEdgeFromSlice(revEdges[toID], fromID)
	}

	return nil
}

func removeEdgeFromSlice(edges []Edge, toID string) []Edge {
	result := make([]Edge, 0, len(edges))
	for _, e := range edges {
		if e.To != toID {
			result = append(result, e)
		}
	}
	return result
}

func (g *GraphStore) GetEdges(ctx context.Context, linkType string, nodeID string) ([]Edge, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var result []Edge

	if edges, ok := g.edges[linkType]; ok {
		result = append(result, edges[nodeID]...)
	}

	if revEdges, ok := g.revIdx[linkType]; ok {
		result = append(result, revEdges[nodeID]...)
	}

	return result, nil
}

func (g *GraphStore) Traverse(ctx context.Context, startID string, linkTypes []string, depth int) ([]Path, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var paths []Path
	visited := make(map[string]bool)

	g.traverseDFS(startID, linkTypes, depth, visited, Path{
		Nodes: []string{startID},
		Edges: []Edge{},
	}, &paths)

	return paths, nil
}

func (g *GraphStore) traverseDFS(currentID string, linkTypes []string, remainingDepth int, visited map[string]bool, currentPath Path, result *[]Path) {
	if remainingDepth == 0 {
		pathCopy := currentPath
		*result = append(*result, pathCopy)
		return
	}

	visited[currentID] = true

	for _, linkType := range linkTypes {
		if edges, ok := g.edges[linkType]; ok {
			for _, edge := range edges[currentID] {
				if !visited[edge.To] {
					newPath := Path{
						Nodes: append(append([]string{}, currentPath.Nodes...), edge.To),
						Edges: append(append([]Edge{}, currentPath.Edges...), edge),
					}
					g.traverseDFS(edge.To, linkTypes, remainingDepth-1, visited, newPath, result)
				}
			}
		}
	}

	delete(visited, currentID)
}
```

- [ ] **Step 4: 编写测试**
```go
// internal/storage/memory/object_test.go
package memory

import (
	"context"
	"testing"
)

func TestObjectStore(t *testing.T) {
	store := NewObjectStore()
	ctx := context.Background()

	err := store.Create(ctx, "Person", "p1", map[string]any{"name": "Alice"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	obj, err := store.Get(ctx, "Person", "p1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if obj["name"] != "Alice" {
		t.Errorf("expected name Alice, got %v", obj["name"])
	}

	err = store.Update(ctx, "Person", "p1", map[string]any{"name": "Bob"})
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	obj, _ = store.Get(ctx, "Person", "p1")
	if obj["name"] != "Bob" {
		t.Errorf("expected name Bob, got %v", obj["name"])
	}

	all, err := store.List(ctx, "Person", nil)
	if len(all) != 1 {
		t.Errorf("expected 1 object, got %d", len(all))
	}

	err = store.Delete(ctx, "Person", "p1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = store.Get(ctx, "Person", "p1")
	if err == nil {
		t.Error("expected error after delete")
	}
}
```

```go
// internal/storage/memory/graph_test.go
package memory

import (
	"context"
	"testing"
)

func TestGraphStore(t *testing.T) {
	store := NewGraphStore()
	ctx := context.Background()

	err := store.AddEdge(ctx, "knows", "p1", "p2", map[string]any{"weight": 0.9})
	if err != nil {
		t.Fatalf("AddEdge failed: %v", err)
	}

	edges, err := store.GetEdges(ctx, "knows", "p1")
	if err != nil {
		t.Fatalf("GetEdges failed: %v", err)
	}
	if len(edges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(edges))
	}

	paths, err := store.Traverse(ctx, "p1", []string{"knows"}, 2)
	if err != nil {
		t.Fatalf("Traverse failed: %v", err)
	}
	if len(paths) != 1 {
		t.Errorf("expected 1 path, got %d", len(paths))
	}
}
```

- [ ] **Step 5: 运行测试**
```bash
go test ./internal/storage/... -v
```
Expected: PASS

- [ ] **Step 6: Commit**
```bash
git add -A && git commit -m "feat(storage): add in-memory object and graph stores"
```

---

## Task 5: 图遍历引擎

**Files:**
- Create: `internal/engine/graph/traverse.go`
- Create: `internal/engine/graph/search.go`
- Create: `internal/engine/graph/traverse_test.go`

- [ ] **Step 1: 实现图遍历引擎**
```go
// internal/engine/graph/traverse.go
package graph

import (
	"context"

	"back-pushing/internal/storage/memory"
)

type TraversalEngine struct {
	graph   *memory.GraphStore
	linkMap map[string]string // link name -> actual link type
}

func NewTraversalEngine(g *memory.GraphStore) *TraversalEngine {
	return &TraversalEngine{
		graph:   g,
		linkMap: make(map[string]string),
	}
}

func (e *TraversalEngine) RegisterLink(name, linkType string) {
	e.linkMap[name] = linkType
}

func (e *TraversalEngine) BFS(ctx context.Context, startID string, linkName string, maxDepth int) ([]string, error) {
	linkType := e.linkMap[linkName]
	if linkType == "" {
		linkType = linkName
	}

	var result []string
	visited := make(map[string]bool)
	queue := []string{startID}
	currentDepth := 0

	for len(queue) > 0 && currentDepth < maxDepth {
		var nextQueue []string
		for _, id := range queue {
			if visited[id] {
				continue
			}
			visited[id] = true
			result = append(result, id)

			edges, err := e.graph.GetEdges(ctx, linkType, id)
			if err != nil {
				return nil, err
			}
			for _, edge := range edges {
				if !visited[edge.To] {
					nextQueue = append(nextQueue, edge.To)
				}
			}
		}
		queue = nextQueue
		currentDepth++
	}

	return result, nil
}

func (e *TraversalEngine) FindPaths(ctx context.Context, startID string, linkNames []string, maxDepth int) ([]Path, error) {
	linkTypes := make([]string, len(linkNames))
	for i, name := range linkNames {
		if lt, ok := e.linkMap[name]; ok {
			linkTypes[i] = lt
		} else {
			linkTypes[i] = name
		}
	}

	return e.graph.Traverse(ctx, startID, linkTypes, maxDepth)
}
```

- [ ] **Step 2: 实现图搜索**
```go
// internal/engine/graph/search.go
package graph

import (
	"context"
	"strings"
)

type SearchResult struct {
	NodeID string
	Type   string
	Data   map[string]any
}

type SearchEngine struct {
	traversal *TraversalEngine
	objStore  *memory.ObjectStore
}

func NewSearchEngine(t *TraversalEngine, o *memory.ObjectStore) *SearchEngine {
	return &SearchEngine{
		traversal: t,
		objStore:  o,
	}
}

func (e *SearchEngine) FindConnected(ctx context.Context, nodeID string, linkName string, maxDepth int) ([]SearchResult, error) {
	connected, err := e.traversal.BFS(ctx, nodeID, linkName, maxDepth)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	for _, id := range connected {
		if id == nodeID {
			continue
		}
		results = append(results, SearchResult{
			NodeID: id,
			Type:   linkName,
		})
	}

	return results, nil
}

func (e *SearchEngine) FullTextSearch(ctx context.Context, objType string, query string) ([]SearchResult, error) {
	all, err := e.objStore.List(ctx, objType, nil)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []SearchResult

	for _, obj := range all {
		for _, v := range obj {
			if str, ok := v.(string); ok {
				if strings.Contains(strings.ToLower(str), query) {
					results = append(results, SearchResult{
						NodeID: getID(obj),
						Type:   objType,
						Data:   obj,
					})
					break
				}
			}
		}
	}

	return results, nil
}

func getID(data map[string]any) string {
	if id, ok := data["id"].(string); ok {
		return id
	}
	return ""
}
```

- [ ] **Step 3: 编写测试**
```go
// internal/engine/graph/traverse_test.go
package graph

import (
	"context"
	"testing"

	"back-pushing/internal/storage/memory"
)

func TestTraversalEngine(t *testing.T) {
	graph := memory.NewGraphStore()
	ctx := context.Background()

	graph.AddEdge(ctx, "knows", "p1", "p2", nil)
	graph.AddEdge(ctx, "knows", "p2", "p3", nil)
	graph.AddEdge(ctx, "knows", "p1", "p3", nil)

	engine := NewTraversalEngine(graph)
	engine.RegisterLink("knows", "knows")

	connected, err := engine.BFS(ctx, "p1", "knows", 2)
	if err != nil {
		t.Fatalf("BFS failed: %v", err)
	}

	if len(connected) < 2 {
		t.Errorf("expected at least 2 connected nodes, got %d", len(connected))
	}
}

func TestSearchEngine(t *testing.T) {
	graph := memory.NewGraphStore()
	objStore := memory.NewObjectStore()
	ctx := context.Background()

	objStore.Create(ctx, "Person", "p1", map[string]any{"name": "Alice", "id": "p1"})
	objStore.Create(ctx, "Person", "p2", map[string]any{"name": "Bob", "id": "p2"})
	graph.AddEdge(ctx, "knows", "p1", "p2", nil)

	traversal := NewTraversalEngine(graph)
	search := NewSearchEngine(traversal, objStore)

	results, err := search.FullTextSearch(ctx, "Person", "alice")
	if err != nil {
		t.Fatalf("FullTextSearch failed: %v", err)
	}

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}
```

- [ ] **Step 4: 运行测试**
```bash
go test ./internal/engine/... -v
```
Expected: PASS

- [ ] **Step 5: Commit**
```bash
git add -A && git commit -m "feat(engine): add graph traversal and search engine"
```

---

## Task 6: Action调度器

**Files:**
- Create: `internal/action/context.go`
- Create: `internal/action/dispatcher.go`
- Create: `internal/action/examples.go`
- Create: `internal/action/dispatcher_test.go`

- [ ] **Step 1: 实现Action上下文**
```go
// internal/action/context.go
package action

import (
	"context"
	"sync"
	"time"

	"back-pushing/internal/storage/memory"
)

type ActionContext struct {
	ObjectStore *memory.ObjectStore
	EventLogger *EventLogger
	Workflow    *WorkflowEngine
	AuditLog    *AuditLogger
	mu          sync.RWMutex
	state       map[string]map[string]any
}

func NewActionContext(objStore *memory.ObjectStore) *ActionContext {
	return &ActionContext{
		ObjectStore: objStore,
		EventLogger: NewEventLogger(),
		Workflow:    NewWorkflowEngine(),
		AuditLog:    NewAuditLogger(),
		state:       make(map[string]map[string]any),
	}
}

func (ctx *ActionContext) LogEvent(event *Event) error {
	return ctx.EventLogger.Log(event)
}

func (ctx *ActionContext) TriggerWorkflow(name string, params map[string]any) error {
	return ctx.Workflow.Trigger(name, params)
}

func (ctx *ActionContext) UpdateObject(objType, id string, data map[string]any) error {
	return ctx.ObjectStore.Update(context.Background(), objType, id, data)
}

func (ctx *ActionContext) SetState(key string, value map[string]any) {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	if ctx.state[key] == nil {
		ctx.state[key] = make(map[string]any)
	}
	for k, v := range value {
		ctx.state[key][k] = v
	}
}

func (ctx *ActionContext) GetState(key string) map[string]any {
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	if ctx.state[key] == nil {
		return nil
	}
	result := make(map[string]any)
	for k, v := range ctx.state[key] {
		result[k] = v
	}
	return result
}

type Event struct {
	Type      string
	Timestamp time.Time
	Payload   any
	Actor     string
}

type EventLogger struct {
	mu     sync.Mutex
	events []Event
}

func NewEventLogger() *EventLogger {
	return &EventLogger{
		events: make([]Event, 0),
	}
}

func (l *EventLogger) Log(e *Event) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}
	l.events = append(l.events, *e)
	return nil
}

func (l *EventLogger) GetEvents() []Event {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Event{}, l.events...)
}

type WorkflowEngine struct {
	mu       sync.Mutex
	workflows map[string]func(map[string]any) error
}

func NewWorkflowEngine() *WorkflowEngine {
	return &WorkflowEngine{
		workflows: make(map[string]func(map[string]any) error),
	}
}

func (w *WorkflowEngine) Register(name string, handler func(map[string]any) error) {
	w.workflows[name] = handler
}

func (w *WorkflowEngine) Trigger(name string, params map[string]any) error {
	w.mu.Lock()
	handler, ok := w.workflows[name]
	w.mu.Unlock()

	if !ok {
		return nil
	}
	return handler(params)
}

type AuditLogger struct {
	mu   sync.Mutex
	logs []AuditEntry
}

type AuditEntry struct {
	Timestamp  time.Time
	Action     string
	ObjectType string
	ObjectID   string
	Actor      string
	Result     string
}

func NewAuditLogger() *AuditLogger {
	return &AuditLogger{
		logs: make([]AuditEntry, 0),
	}
}

func (l *AuditLogger) Log(entry AuditEntry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	l.logs = append(l.logs, entry)
}
```

- [ ] **Step 2: 实现Action调度器**
```go
// internal/action/dispatcher.go
package action

import (
	"fmt"
)

type Handler func(ctx *ActionContext, input any) (output any, err error)

type Dispatcher struct {
	mu map[string]Handler
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		mu: make(map[string]Handler),
	}
}

func (d *Dispatcher) Register(name string, handler Handler) {
	d.mu[name] = handler
}

func (d *Dispatcher) Dispatch(ctx *ActionContext, name string, input any) (any, error) {
	handler, ok := d.mu[name]
	if !ok {
		return nil, fmt.Errorf("action %s not found", name)
	}

	output, err := handler(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("action %s failed: %w", name, err)
	}

	return output, nil
}

func (d *Dispatcher) ListActions() []string {
	actions := make([]string, 0, len(d.mu))
	for name := range d.mu {
		actions = append(actions, name)
	}
	return actions
}
```

- [ ] **Step 3: 编写示例Action**
```go
// internal/action/examples.go
package action

import (
	"fmt"
	"time"
)

type EscalateReviewInput struct {
	Reason   string `json:"reason"`
	Priority string `json:"priority"`
	PersonID string `json:"person_id"`
}

type EscalateReviewOutput struct {
	CaseID     string    `json:"case_id"`
	AssignedTo string    `json:"assigned_to"`
	Status     string    `json:"status"`
	Timestamp  time.Time `json:"timestamp"`
}

func EscalateReview(ctx *ActionContext, input any) (any, error) {
	in, ok := input.(EscalateReviewInput)
	if !ok {
		return nil, fmt.Errorf("invalid input type")
	}

	caseID := fmt.Sprintf("CASE-%d", time.Now().UnixNano()%1000000)

	event := &Event{
		Type:      "REVIEW_ESCALATION",
		Timestamp: time.Now(),
		Payload:   in,
	}
	ctx.LogEvent(event)

	ctx.Workflow.Register("review_escalation", func(params map[string]any) error {
		return nil
	})
	ctx.TriggerWorkflow("review_escalation", map[string]any{
		"case_id":  caseID,
		"person_id": in.PersonID,
		"priority":  in.Priority,
	})

	ctx.UpdateObject("Person", in.PersonID, map[string]any{
		"review_status":    "escalated",
		"last_escalation": time.Now(),
	})

	return EscalateReviewOutput{
		CaseID:     caseID,
		AssignedTo: "risk_team",
		Status:     "pending",
		Timestamp:  time.Now(),
	}, nil
}
```

- [ ] **Step 4: 编写测试**
```go
// internal/action/dispatcher_test.go
package action

import (
	"context"
	"testing"

	"back-pushing/internal/storage/memory"
)

func TestDispatcher(t *testing.T) {
	objStore := memory.NewObjectStore()
	ctx := NewActionContext(objStore)
	dispatcher := NewDispatcher()

	dispatcher.Register("escalate_review", EscalateReview)

	objStore.Create(context.Background(), "Person", "p1", map[string]any{
		"id":   "p1",
		"name": "Alice",
	})

	input := EscalateReviewInput{
		Reason:   "Suspicious activity",
		Priority: "high",
		PersonID: "p1",
	}

	output, err := dispatcher.Dispatch(ctx, "escalate_review", input)
	if err != nil {
		t.Fatalf("Dispatch failed: %v", err)
	}

	result := output.(EscalateReviewOutput)
	if result.CaseID == "" {
		t.Error("CaseID should not be empty")
	}
	if result.Status != "pending" {
		t.Errorf("expected status pending, got %s", result.Status)
	}
}
```

- [ ] **Step 5: 运行测试**
```bash
go test ./internal/action/... -v
```
Expected: PASS

- [ ] **Step 6: Commit**
```bash
git add -A && git commit -m "feat(action): add action context and dispatcher"
```

---

## Task 7: 服务入口集成

**Files:**
- Modify: `cmd/server/main.go`
- Create: `ontology/person.yaml`
- Create: `ontology/links.yaml`
- Create: `ontology/paths.yaml`

- [ ] **Step 1: 实现完整main.go**
```go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/graphql-go/graphql"

	"back-pushing/internal/action"
	"back-pushing/internal/engine/graph"
	"back-pushing/internal/ontology"
	"back-pushing/internal/storage/memory"
)

func main() {
	ontologyDir := flag.String("ontology", "./ontology", "Ontology YAML directory")
	addr := flag.String("addr", ":8080", "Server address")
	flag.Parse()

	log.Println("Back-Pushing Server Starting...")

	// 1. 解析Ontology
	log.Println("Loading Ontology from:", *ontologyDir)
	ont, err := ontology.ParseOntology(*ontologyDir)
	if err != nil {
		log.Fatalf("Failed to parse ontology: %v", err)
	}
	log.Printf("Loaded %d object types, %d links", len(ont.ObjectTypes), len(ont.Links))

	// 2. 初始化存储
	objStore := memory.NewObjectStore()
	graphStore := memory.NewGraphStore()

	// 3. 初始化引擎
	traversal := graph.NewTraversalEngine(graphStore)
	search := graph.NewSearchEngine(traversal, objStore)

	// 4. 注册Action
	dispatcher := action.NewDispatcher()
	dispatcher.Register("action.escalate_review", action.EscalateReview)

	// 5. 创建GraphQL schema（简化实现）
	schema, err := createSchema(objStore, graphStore, traversal, search, dispatcher)
	if err != nil {
		log.Fatalf("Failed to create schema: %v", err)
	}

	// 6. 启动服务
	log.Println("Starting GraphQL server at", *addr)
	http.Handle("/graphql", &graphqlHandler{schema: schema})
	http.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Back-Pushing GraphQL Server\nGraphQL endpoint: /graphql")
	}))

	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

type graphqlHandler struct {
	schema *graphql.Schema
}

func (h *graphqlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	result := graphql.Do(graphql.Params{
		Schema:         *h.schema,
		RequestString:  req.Query,
		VariableValues: req.Variables,
		OperationName: req.OperationName,
	})
	json.NewEncoder(w).Encode(result)
}

func createSchema(objStore *memory.ObjectStore, graphStore *memory.GraphStore, traversal *graph.TraversalEngine, search *graph.SearchEngine, dispatcher *action.Dispatcher) (*graphql.Schema, error) {
	// 简化schema定义
	objectType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Object",
		Fields: graphql.Fields{
			"id": &graphql.Field{Type: graphql.String},
			"type": &graphql.Field{Type: graphql.String},
			"data": &graphql.Field{Type: graphql.JSON},
		},
	})

	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"object": &graphql.Field{
				Type: objectType,
				Args: graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"id":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					objType := p.Args["type"].(string)
					id := p.Args["id"].(string)
					data, err := objStore.Get(context.Background(), objType, id)
					if err != nil {
						return nil, err
					}
					return map[string]any{"id": id, "type": objType, "data": data}, nil
				},
			},
			"search": &graphql.Field{
				Type: graphql.NewList(objectType),
				Args: graphql.FieldConfigArgument{
					"query": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"type":  &graphql.ArgumentConfig{Type: graphql.String},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					query := p.Args["query"].(string)
					var objType *string
					if t, ok := p.Args["type"].(string); ok {
						objType = &t
					}
					if objType != nil {
						results, err := search.FullTextSearch(context.Background(), *objType, query)
						if err != nil {
							return nil, err
						}
						out := make([]map[string]any, len(results))
						for i, r := range results {
							out[i] = map[string]any{"id": r.NodeID, "type": r.Type, "data": r.Data}
						}
						return out, nil
					}
					return nil, nil
				},
			},
		},
	})

	mutationType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Mutation",
		Fields: graphql.Fields{
			"createObject": &graphql.Field{
				Type: objectType,
				Args: graphql.FieldConfigArgument{
					"type": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"id":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"data": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.JSON)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					objType := p.Args["type"].(string)
					id := p.Args["id"].(string)
					data := p.Args["data"].(map[string]interface{})
					err := objStore.Create(context.Background(), objType, id, data)
					if err != nil {
						return nil, err
					}
					return map[string]any{"id": id, "type": objType, "data": data}, nil
				},
			},
			"addEdge": &graphql.Field{
				Type: graphql.NewObject(graphql.ObjectConfig{
					Name: "Edge",
					Fields: graphql.Fields{
						"from":  &graphql.Field{Type: graphql.String},
						"to":    &graphql.Field{Type: graphql.String},
						"props": &graphql.Field{Type: graphql.JSON},
					},
				}),
				Args: graphql.FieldConfigArgument{
					"linkType": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"fromId":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"toId":     &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"props":    &graphql.ArgumentConfig{Type: graphql.JSON},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					linkType := p.Args["linkType"].(string)
					fromID := p.Args["fromId"].(string)
					toID := p.Args["toId"].(string)
					var props map[string]any
					if pr, ok := p.Args["props"].(map[string]any); ok {
						props = pr
					}
					err := graphStore.AddEdge(context.Background(), linkType, fromID, toID, props)
					if err != nil {
						return nil, err
					}
					return map[string]any{"from": fromID, "to": toID, "props": props}, nil
				},
			},
			"invokeAction": &graphql.Field{
				Type: graphql.JSON,
				Args: graphql.FieldConfigArgument{
					"name":  &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
					"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.JSON)},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					name := p.Args["name"].(string)
					input := p.Args["input"]
					output, err := dispatcher.Dispatch(action.NewActionContext(objStore), name, input)
					if err != nil {
						return nil, err
					}
					return output, nil
				},
			},
		},
	})

	return graphql.NewSchema(graphql.SchemaConfig{
		Query:    queryType,
		Mutation: mutationType,
	})
}
```

- [ ] **Step 2: 创建示例Ontology文件**

```yaml
# ontology/person.yaml
object_types:
  Person:
    description: 自然人实体
    type: entity
    properties:
      id:
        type: string
        primary: true
      name:
        type: string
        indexed: true
      email:
        type: string
        unique: true
      tags:
        type: list[string]

    actions:
      - name: "Escalate to Review"
        handler: action.escalate_review
```

```yaml
# ontology/links.yaml
links:
  knows:
    source: Person
    target: Person
    type: many-to-many
```

```yaml
# ontology/paths.yaml
paths:
  social_network:
    description: "社交网络路径"
    steps:
      - Person-[knows]->Person
```

- [ ] **Step 3: 验证编译和运行**
```bash
go mod tidy
go build -o bin/server ./cmd/server && ./bin/server -ontology ./ontology
```
Expected: 服务启动成功，监听 :8080

- [ ] **Step 4: Commit**
```bash
git add -A && git commit -m "feat: integrate all components in main server"
```

---

## 自检清单

- [x] **Spec覆盖检查**：
  - Ontology Core（YAML解析 + 对象注册）→ Task 2, 3
  - 内存存储层（对象 + 图）→ Task 4
  - 图探索（基础遍历）→ Task 5
  - Action框架 → Task 6
  - 服务入口集成 → Task 7

- [x] **Placeholder扫描**：无TBD/TODO

- [x] **类型一致性**：
  - `ObjectStore.Create/Get/Update/Delete` 接口一致
  - `GraphStore.AddEdge/RemoveEdge/GetEdges/Traverse` 接口一致
  - `ActionContext` 方法命名一致
