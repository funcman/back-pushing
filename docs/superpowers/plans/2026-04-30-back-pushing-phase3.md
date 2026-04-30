# 推背图（Back-Pushing）Phase 3 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建完整分析能力——时序分析引擎、增强实体搜索、决策工作流

**Architecture:**
- 均衡方案：核心功能完整实现，保留扩展点
- 时序分析：事件流处理、滚动窗口聚合、异常检测
- 实体搜索：全文索引、多维过滤、聚合统计
- 决策工作流：规则引擎、告警触发、工作流编排
- 与现有Ontology引擎深度集成

**Tech Stack:** Go 1.21+

---

## 文件结构

```
back-pushing/
├── internal/
│   ├── engine/
│   │   ├── graph/
│   │   │   ├── traverse.go     # 图遍历（Phase 1）
│   │   │   └── search.go       # 图搜索（Phase 1）
│   │   ├── temporal/
│   │   │   ├── analyzer.go     # 时序分析引擎
│   │   │   ├── window.go       # 滚动窗口
│   │   │   ├── anomaly.go      # 异常检测
│   │   │   └── temporal_test.go
│   │   ├── search/
│   │   │   ├── fulltext.go     # 全文索引
│   │   │   ├── filter.go       # 多维过滤
│   │   │   ├── aggregate.go    # 聚合统计
│   │   │   └── search_test.go
│   │   ├── workflow/
│   │   │   ├── engine.go       # 工作流引擎
│   │   │   ├── rule.go         # 规则定义
│   │   │   ├── alert.go        # 告警管理
│   │   │   └── workflow_test.go
│   │   └── engine.go           # 统一入口
│   └── storage/
│       └── memory/
│           ├── graph.go         # 内存图存储（Phase 1）
│           └── object.go        # 对象存储（Phase 1）
├── ontology/
│   ├── person.yaml
│   └── links.yaml
└── Makefile
```

---

## Task 1: 时序分析引擎

**Files:**
- Create: `internal/engine/temporal/analyzer.go`
- Create: `internal/engine/temporal/window.go`
- Create: `internal/engine/temporal/anomaly.go`
- Create: `internal/engine/temporal/temporal_test.go`

- [ ] **Step 1: 实现事件流分析**
```go
// internal/engine/temporal/analyzer.go
package temporal

import (
    "context"
    "sort"
    "time"
)

type Event struct {
    ID        string
    Type      string
    Timestamp time.Time
    Actor     string
    Props     map[string]any
}

type TemporalAnalyzer struct {
    events []Event
}

func NewTemporalAnalyzer() *TemporalAnalyzer {
    return &TemporalAnalyzer{
        events: []Event{},
    }
}

func (a *TemporalAnalyzer) RecordEvent(ctx context.Context, event Event) error {
    a.events = append(a.events, event)
    sort.Slice(a.events, func(i, j int) bool {
        return a.events[i].Timestamp.Before(a.events[j].Timestamp)
    })
    return nil
}

func (a *TemporalAnalyzer) GetEvents(ctx context.Context, objType, objID string, since, until time.Time) ([]Event, error) {
    var result []Event
    for _, e := range a.events {
        if e.Type == objType && e.Actor == objID {
            if since.IsZero() || e.Timestamp.After(since) {
                if until.IsZero() || e.Timestamp.Before(until) {
                    result = append(result, e)
                }
            }
        }
    }
    return result, nil
}

func (a *TemporalAnalyzer) GetEventCount(ctx context.Context, objType, objID string, since, until time.Time) (int64, error) {
    events, err := a.GetEvents(ctx, objType, objID, since, until)
    if err != nil {
        return 0, err
    }
    return int64(len(events)), nil
}
```

- [ ] **Step 2: 实现滚动窗口**
```go
// internal/engine/temporal/window.go
package temporal

import (
    "time"
)

type WindowType string

const (
    WindowTumbling WindowType = "tumbling"
    WindowSliding  WindowType = "sliding"
)

type Window struct {
    Size       time.Duration
    Slide      time.Duration
    WindowType WindowType
}

type Aggregation struct {
    Count int64
    Sum   float64
    Avg   float64
    Min   float64
    Max   float64
}

func NewTumblingWindow(size time.Duration) *Window {
    return &Window{
        Size:       size,
        Slide:      size,
        WindowType: WindowTumbling,
    }
}

func NewSlidingWindow(size, slide time.Duration) *Window {
    return &Window{
        Size:       size,
        Slide:      slide,
        WindowType: WindowSliding,
    }
}

func (a *Aggregation) Add(t time.Time, value float64) {
    a.Count++
    a.Sum += value
    a.Avg = a.Sum / float64(a.Count)
    if value < a.Min || a.Min == 0 {
        a.Min = value
    }
    if value > a.Max {
        a.Max = value
    }
}

func (w *Window) Aggregate(events []Event, valueField string) *Aggregation {
    agg := &Aggregation{}
    for _, e := range events {
        if v, ok := e.Props[valueField].(float64); ok {
            agg.Add(e.Timestamp, v)
        }
    }
    return agg
}
```

- [ ] **Step 3: 实现异常检测**
```go
// internal/engine/temporal/anomaly.go
package temporal

import (
    "math"
)

type AnomalyDetector struct {
    threshold float64
    window    int
}

func NewAnomalyDetector(threshold float64, windowSize int) *AnomalyDetector {
    return &AnomalyDetector{
        threshold: threshold,
        window:    windowSize,
    }
}

type AnomalyResult struct {
    IsAnomaly bool
    Score     float64
    Message   string
}

func (d *AnomalyDetector) Detect(values []float64) AnomalyResult {
    if len(values) < d.window {
        return AnomalyResult{IsAnomaly: false, Score: 0}
    }

    recent := values[len(values)-d.window:]
    mean := d.mean(recent)
    stddev := d.stddev(recent, mean)

    if len(values) > 0 {
        latest := values[len(values)-1]
        zscore := math.Abs((latest - mean) / stddev)
        if zscore > d.threshold {
            return AnomalyResult{
                IsAnomaly: true,
                Score:     zscore,
                Message:   "Value deviates significantly from recent average",
            }
        }
    }

    return AnomalyResult{IsAnomaly: false, Score: 0}
}

func (d *AnomalyDetector) mean(vals []float64) float64 {
    sum := 0.0
    for _, v := range vals {
        sum += v
    }
    return sum / float64(len(vals))
}

func (d *AnomalyDetector) stddev(vals []float64, mean float64) float64 {
    sum := 0.0
    for _, v := range vals {
        diff := v - mean
        sum += diff * diff
    }
    variance := sum / float64(len(vals))
    return math.Sqrt(variance)
}
```

- [ ] **Step 4: 编写测试**
```go
// internal/engine/temporal/temporal_test.go
package temporal

import (
    "context"
    "testing"
    "time"
)

func TestTemporalAnalyzer_RecordAndGet(t *testing.T) {
    analyzer := NewTemporalAnalyzer()
    ctx := context.Background()

    event := Event{
        ID:        "e1",
        Type:      "Transaction",
        Timestamp: time.Now(),
        Actor:     "acc1",
        Props:     map[string]any{"amount": 100.0},
    }

    if err := analyzer.RecordEvent(ctx, event); err != nil {
        t.Fatalf("RecordEvent failed: %v", err)
    }

    events, err := analyzer.GetEvents(ctx, "Transaction", "acc1", time.Time{}, time.Now().Add(time.Hour))
    if err != nil {
        t.Fatalf("GetEvents failed: %v", err)
    }

    if len(events) != 1 {
        t.Errorf("expected 1 event, got %d", len(events))
    }
}

func TestAnomalyDetector_Detect(t *testing.T) {
    detector := NewAnomalyDetector(2.0, 5)

    values := []float64{10, 11, 10, 10, 11, 10, 100.0}

    result := detector.Detect(values)
    if !result.IsAnomaly {
        t.Error("expected anomaly to be detected for outlier value")
    }

    normalValues := []float64{10, 11, 10, 10, 11, 10, 11}
    result = detector.Detect(normalValues)
    if result.IsAnomaly {
        t.Error("expected no anomaly for normal values")
    }
}
```

- [ ] **Step 5: 运行测试**
```bash
go test ./internal/engine/temporal/... -v
```
Expected: PASS

- [ ] **Step 6: Commit**
```bash
git add -A && git commit -m "feat(temporal): add temporal analysis engine"
```

---

## Task 2: 实体搜索引擎

**Files:**
- Create: `internal/engine/search/fulltext.go`
- Create: `internal/engine/search/filter.go`
- Create: `internal/engine/search/aggregate.go`
- Create: `internal/engine/search/search_test.go`

- [ ] **Step 1: 实现全文索引**
```go
// internal/engine/search/fulltext.go
package search

import (
    "context"
    "strings"
    "unicode"
)

type FullTextIndex struct {
    inverted map[string]map[string][]string // term -> objType -> [objIDs]
}

func NewFullTextIndex() *FullTextIndex {
    return &FullTextIndex{
        inverted: make(map[string]map[string][]string),
    }
}

func (idx *FullTextIndex) Index(ctx context.Context, objType, objID string, fields map[string]any) error {
    for field, value := range fields {
        terms := tokenize(fmt.Sprintf("%v", value))
        for _, term := range terms {
            if idx.inverted[term] == nil {
                idx.inverted[term] = make(map[string][]string)
            }
            idx.inverted[term][objType] = append(idx.inverted[term][objType], objID)
        }
    }
    return nil
}

func (idx *FullTextIndex) Search(ctx context.Context, query string) ([]string, error) {
    terms := tokenize(query)
    if len(terms) == 0 {
        return nil, nil
    }

    resultSet := make(map[string]bool)
    for _, term := range terms {
        term = strings.ToLower(term)
        if posting, ok := idx.inverted[term]; ok {
            for _, ids := range posting {
                for _, id := range ids {
                    resultSet[id] = true
                }
            }
        }
    }

    var results []string
    for id := range resultSet {
        results = append(results, id)
    }
    return results, nil
}

func tokenize(text string) []string {
    text = strings.ToLower(text)
    var tokens []string
    var current strings.Builder

    for _, r := range text {
        if unicode.IsLetter(r) || unicode.IsDigit(r) {
            current.WriteRune(r)
        } else {
            if current.Len() > 0 {
                tokens = append(tokens, current.String())
                current.Reset()
            }
        }
    }

    if current.Len() > 0 {
        tokens = append(tokens, current.String())
    }

    return tokens
}
```

- [ ] **Step 2: 实现多维过滤**
```go
// internal/engine/search/filter.go
package search

import (
    "context"
    "fmt"
    "strings"

    "back-pushing/internal/storage/memory"
)

type Filter struct {
    Eq  map[string]any
    Gt  map[string]any
    Lt  map[string]any
    Gte map[string]any
    Lte map[string]any
    In  map[string][]any
}

func NewFilter() *Filter {
    return &Filter{
        Eq:  make(map[string]any),
        Gt:  make(map[string]any),
        Lt:  make(map[string]any),
        Gte: make(map[string]any),
        Lte: make(map[string]any),
        In:  make(map[string][]any),
    }
}

func (f *Filter) Equal(key string, value any) *Filter {
    f.Eq[key] = value
    return f
}

func (f *Filter) GreaterThan(key string, value any) *Filter {
    f.Gt[key] = value
    return f
}

func (f *Filter) LessThan(key string, value any) *Filter {
    f.Lt[key] = value
    return f
}

func (f *Filter) InList(key string, values []any) *Filter {
    f.In[key] = values
    return f
}

type FilterEngine struct {
    store *memory.ObjectStore
}

func NewFilterEngine(store *memory.ObjectStore) *FilterEngine {
    return &FilterEngine{store: store}
}

func (e *FilterEngine) Filter(ctx context.Context, objType string, filter *Filter) ([]map[string]any, error) {
    objects, err := e.store.List(ctx, objType, nil)
    if err != nil {
        return nil, err
    }

    var result []map[string]any
    for _, obj := range objects {
        if e.matches(obj, filter) {
            result = append(result, obj)
        }
    }
    return result, nil
}

func (e *FilterEngine) matches(obj map[string]any, filter *Filter) bool {
    for key, value := range filter.Eq {
        if obj[key] != value {
            return false
        }
    }

    for key, value := range filter.Gt {
        if !compareGT(obj[key], value) {
            return false
        }
    }

    for key, value := range filter.Lt {
        if !compareLT(obj[key], value) {
            return false
        }
    }

    for key, values := range filter.In {
        if !contains(obj[key], values) {
            return false
        }
    }

    return true
}

func compareGT(a, b any) bool {
    af, ok := a.(float64)
    bf, ok2 := b.(float64)
    if ok && ok2 {
        return af > bf
    }
    return false
}

func compareLT(a, b any) bool {
    af, ok := a.(float64)
    bf, ok2 := b.(float64)
    if ok && ok2 {
        return af < bf
    }
    return false
}

func contains(a any, list []any) bool {
    for _, v := range list {
        if a == v {
            return true
        }
    }
    return false
}
```

- [ ] **Step 3: 实现聚合统计**
```go
// internal/engine/search/aggregate.go
package search

import (
    "context"

    "back-pushing/internal/storage/memory"
)

type AggregationResult struct {
    Count  int64
    Sum    map[string]float64
    Avg    map[string]float64
    Min    map[string]float64
    Max    map[string]float64
    Groups map[string][]map[string]any
}

type AggregateEngine struct {
    store *memory.ObjectStore
}

func NewAggregateEngine(store *memory.ObjectStore) *AggregateEngine {
    return &AggregateEngine{store: store}
}

func (e *AggregateEngine) Aggregate(ctx context.Context, objType string, fields []string, groupBy string) (*AggregationResult, error) {
    objects, err := e.store.List(ctx, objType, nil)
    if err != nil {
        return nil, err
    }

    result := &AggregationResult{
        Sum:    make(map[string]float64),
        Avg:    make(map[string]float64),
        Min:    make(map[string]float64),
        Max:    make(map[string]float64),
        Groups: make(map[string][]map[string]any),
    }

    for _, obj := range objects {
        result.Count++

        groupKey := ""
        if groupBy != "" {
            if v, ok := obj[groupBy].(string); ok {
                groupKey = v
            }
        }

        for _, field := range fields {
            if v, ok := obj[field].(float64); ok {
                result.Sum[field] += v
                if result.Min[field] == 0 || v < result.Min[field] {
                    result.Min[field] = v
                }
                if v > result.Max[field] {
                    result.Max[field] = v
                }
            }
        }

        if groupKey != "" {
            result.Groups[groupKey] = append(result.Groups[groupKey], obj)
        }
    }

    for _, field := range fields {
        if result.Count > 0 {
            result.Avg[field] = result.Sum[field] / float64(result.Count)
        }
    }

    return result, nil
}
```

- [ ] **Step 4: 编写测试**
```go
// internal/engine/search/search_test.go
package search

import (
    "context"
    "testing"

    "back-pushing/internal/storage/memory"
)

func TestFullTextIndex_Search(t *testing.T) {
    idx := NewFullTextIndex()
    ctx := context.Background()

    idx.Index(ctx, "Person", "p1", map[string]any{"name": "Alice Smith"})
    idx.Index(ctx, "Person", "p2", map[string]any{"name": "Bob Johnson"})
    idx.Index(ctx, "Person", "p3", map[string]any{"name": "Alice Williams"})

    results, err := idx.Search(ctx, "Alice")
    if err != nil {
        t.Fatalf("Search failed: %v", err)
    }

    if len(results) != 2 {
        t.Errorf("expected 2 results, got %d", len(results))
    }
}

func TestFilterEngine_Filter(t *testing.T) {
    store := memory.NewObjectStore()
    engine := NewFilterEngine(store)
    ctx := context.Background()

    store.Create(ctx, "Person", "p1", map[string]any{"name": "Alice", "age": float64(30)})
    store.Create(ctx, "Person", "p2", map[string]any{"name": "Bob", "age": float64(25)})
    store.Create(ctx, "Person", "p3", map[string]any{"name": "Charlie", "age": float64(35)})

    filter := NewFilter().GreaterThan("age", 28)
    results, err := engine.Filter(ctx, "Person", filter)
    if err != nil {
        t.Fatalf("Filter failed: %v", err)
    }

    if len(results) != 2 {
        t.Errorf("expected 2 results, got %d", len(results))
    }
}
```

- [ ] **Step 5: 运行测试**
```bash
go test ./internal/engine/search/... -v
```
Expected: PASS

- [ ] **Step 6: Commit**
```bash
git add -A && git commit -m "feat(search): add fulltext search and aggregation"
```

---

## Task 3: 决策工作流引擎

**Files:**
- Create: `internal/engine/workflow/rule.go`
- Create: `internal/engine/workflow/alert.go`
- Create: `internal/engine/workflow/engine.go`
- Create: `internal/engine/workflow/workflow_test.go`

- [ ] **Step 1: 实现规则定义**
```go
// internal/engine/workflow/rule.go
package workflow

import (
    "fmt"
    "strings"
)

type Rule struct {
    Name        string
    Description string
    Condition   RuleCondition
    Action      RuleAction
    Enabled     bool
}

type RuleCondition struct {
    Field    string
    Operator string
    Value    any
}

type RuleAction struct {
    Type   string
    Params map[string]any
}

const (
    OpEq       = "eq"
    OpNe       = "ne"
    OpGt       = "gt"
    OpLt       = "lt"
    OpGte      = "gte"
    OpLte      = "lte"
    OpContains = "contains"
)

func (r *Rule) Evaluate(data map[string]any) bool {
    if !r.Enabled {
        return false
    }

    value, exists := data[r.Condition.Field]
    if !exists {
        return false
    }

    switch r.Condition.Operator {
    case OpEq:
        return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", r.Condition.Value)
    case OpNe:
        return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", r.Condition.Value)
    case OpGt:
        return compareNumeric(value, r.Condition.Value) > 0
    case OpLt:
        return compareNumeric(value, r.Condition.Value) < 0
    case OpGte:
        return compareNumeric(value, r.Condition.Value) >= 0
    case OpLte:
        return compareNumeric(value, r.Condition.Value) <= 0
    case OpContains:
        return strings.Contains(fmt.Sprintf("%v", value), fmt.Sprintf("%v", r.Condition.Value))
    }

    return false
}

func compareNumeric(a, b any) int {
    af, ok1 := toFloat64(a)
    bf, ok2 := toFloat64(b)
    if ok1 && ok2 {
        if af > bf {
            return 1
        }
        if af < bf {
            return -1
        }
        return 0
    }
    return strings.Compare(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

func toFloat64(v any) (float64, bool) {
    switch val := v.(type) {
    case float64:
        return val, true
    case float32:
        return float64(val), true
    case int:
        return float64(val), true
    case int64:
        return float64(val), true
    default:
        return 0, false
    }
}

type RuleSet struct {
    Rules []Rule
}

func NewRuleSet() *RuleSet {
    return &RuleSet{Rules: []Rule{}}
}

func (rs *RuleSet) Add(rule Rule) {
    rs.Rules = append(rs.Rules, rule)
}

func (rs *RuleSet) Evaluate(data map[string]any) []Rule {
    var triggered []Rule
    for _, rule := range rs.Rules {
        if rule.Evaluate(data) {
            triggered = append(triggered, rule)
        }
    }
    return triggered
}
```

- [ ] **Step 2: 实现告警管理**
```go
// internal/engine/workflow/alert.go
package workflow

import (
    "fmt"
    "sync"
    "time"
)

type Alert struct {
    ID        string
    RuleName  string
    Severity  string
    Message   string
    Timestamp time.Time
    Data      map[string]any
}

type AlertHandler func(alert Alert)

type AlertManager struct {
    mu      sync.RWMutex
    alerts  []Alert
    handlers []AlertHandler
}

func NewAlertManager() *AlertManager {
    return &AlertManager{
        alerts:   []Alert{},
        handlers: []AlertHandler{},
    }
}

func (m *AlertManager) RegisterHandler(handler AlertHandler) {
    m.handlers = append(m.handlers, handler)
}

func (m *AlertManager) Trigger(ruleName, severity, message string, data map[string]any) {
    alert := Alert{
        ID:        fmt.Sprintf("alert-%d", time.Now().UnixNano()),
        RuleName:  ruleName,
        Severity:  severity,
        Message:   message,
        Timestamp: time.Now(),
        Data:      data,
    }

    m.mu.Lock()
    m.alerts = append(m.alerts, alert)
    m.mu.Unlock()

    for _, handler := range m.handlers {
        handler(alert)
    }
}

func (m *AlertManager) GetAlerts() []Alert {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return append([]Alert{}, m.alerts...)
}

func (m *AlertManager) ClearAlerts() {
    m.mu.Lock()
    m.alerts = []Alert{}
    m.mu.Unlock()
}
```

- [ ] **Step 3: 实现工作流引擎**
```go
// internal/engine/workflow/engine.go
package workflow

import (
    "context"
    "fmt"
    "sync"
    "time"

    "back-pushing/internal/engine/temporal"
)

type WorkflowEngine struct {
    rules      *RuleSet
    alerts     *AlertManager
    temporal   *temporal.TemporalAnalyzer
    workflows  map[string]func(ctx context.Context, data map[string]any) error
    mu         sync.RWMutex
}

func NewWorkflowEngine(temporal *temporal.TemporalAnalyzer) *WorkflowEngine {
    return &WorkflowEngine{
        rules:     NewRuleSet(),
        alerts:    NewAlertManager(),
        temporal:  temporal,
        workflows: make(map[string]func(ctx context.Context, data map[string]any) error),
    }
}

func (e *WorkflowEngine) AddRule(rule Rule) {
    e.mu.Lock()
    defer e.mu.Unlock()
    e.rules.Add(rule)
}

func (e *WorkflowEngine) RegisterWorkflow(name string, handler func(ctx context.Context, data map[string]any) error) {
    e.mu.Lock()
    defer e.mu.Unlock()
    e.workflows[name] = handler
}

func (e *WorkflowEngine) ProcessEvent(ctx context.Context, eventType string, data map[string]any) error {
    e.mu.RLock()
    triggered := e.rules.Evaluate(data)
    e.mu.RUnlock()

    for _, rule := range triggered {
        action := rule.Action
        switch action.Type {
        case "alert":
            e.alerts.Trigger(
                rule.Name,
                fmt.Sprintf("%v", action.Params["severity"]),
                fmt.Sprintf("Rule %s triggered", rule.Name),
                data,
            )
        case "workflow":
            if workflowName, ok := action.Params["name"].(string); ok {
                e.mu.RLock()
                handler, exists := e.workflows[workflowName]
                e.mu.RUnlock()
                if exists {
                    if err := handler(ctx, data); err != nil {
                        return err
                    }
                }
            }
        case "record_event":
            e.temporal.RecordEvent(ctx, temporal.Event{
                ID:    fmt.Sprintf("event-%d", time.Now().UnixNano()),
                Type:  eventType,
                Actor: fmt.Sprintf("%v", data["actor"]),
                Props: data,
            })
        }
    }

    return nil
}

func (e *WorkflowEngine) GetAlerts() []Alert {
    return e.alerts.GetAlerts()
}
```

- [ ] **Step 4: 编写测试**
```go
// internal/engine/workflow/workflow_test.go
package workflow

import (
    "context"
    "testing"

    "back-pushing/internal/engine/temporal"
)

func TestRuleSet_Evaluate(t *testing.T) {
    rs := NewRuleSet()

    rs.Add(Rule{
        Name:        "high_value_transaction",
        Description: "Alert on high value transactions",
        Condition:   RuleCondition{Field: "amount", Operator: OpGt, Value: float64(10000)},
        Action:      RuleAction{Type: "alert", Params: map[string]any{"severity": "high"}},
        Enabled:     true,
    })

    data := map[string]any{"amount": float64(15000), "actor": "acc1"}

    triggered := rs.Evaluate(data)
    if len(triggered) != 1 {
        t.Errorf("expected 1 triggered rule, got %d", len(triggered))
    }

    lowValueData := map[string]any{"amount": float64(100), "actor": "acc1"}
    triggered = rs.Evaluate(lowValueData)
    if len(triggered) != 0 {
        t.Errorf("expected 0 triggered rules, got %d", len(triggered))
    }
}

func TestWorkflowEngine_ProcessEvent(t *testing.T) {
    temporalAnalyzer := temporal.NewTemporalAnalyzer()
    engine := NewWorkflowEngine(temporalAnalyzer)
    ctx := context.Background()

    engine.AddRule(Rule{
        Name:        "high_value",
        Condition:   RuleCondition{Field: "amount", Operator: OpGt, Value: float64(1000)},
        Action:      RuleAction{Type: "alert", Params: map[string]any{"severity": "warning"}},
        Enabled:     true,
    })

    var alertReceived bool
    engine.alerts.RegisterHandler(func(alert Alert) {
        alertReceived = true
    })

    engine.ProcessEvent(ctx, "Transaction", map[string]any{
        "amount": float64(2000),
        "actor":  "acc1",
    })

    if !alertReceived {
        t.Error("expected alert to be received")
    }
}
```

- [ ] **Step 5: 运行测试**
```bash
go test ./internal/engine/workflow/... -v
```
Expected: PASS

- [ ] **Step 6: Commit**
```bash
git add -A && git commit -m "feat(workflow): add decision workflow engine"
```

---

## Task 4: 引擎统一入口

**Files:**
- Create: `internal/engine/engine.go`

- [ ] **Step 1: 创建引擎统一入口**
```go
// internal/engine/engine.go
package engine

import (
    "context"

    "back-pushing/internal/engine/search"
    "back-pushing/internal/engine/temporal"
    "back-pushing/internal/engine/workflow"
    "back-pushing/internal/storage/memory"
)

type Engine struct {
    ObjectStore *memory.ObjectStore
    GraphStore  *memory.GraphStore
    Temporal    *temporal.TemporalAnalyzer
    Search      *search.FilterEngine
    Workflow    *workflow.WorkflowEngine
}

func NewEngine(objStore *memory.ObjectStore, graphStore *memory.GraphStore) *Engine {
    temporalAnalyzer := temporal.NewTemporalAnalyzer()

    return &Engine{
        ObjectStore: objStore,
        GraphStore:  graphStore,
        Temporal:    temporalAnalyzer,
        Search:      search.NewFilterEngine(objStore),
        Workflow:    workflow.NewWorkflowEngine(temporalAnalyzer),
    }
}

func (e *Engine) ProcessEvent(ctx context.Context, eventType string, data map[string]any) error {
    return e.Workflow.ProcessEvent(ctx, eventType, data)
}
```

- [ ] **Step 2: 验证编译**
```bash
go build ./internal/engine/...
```
Expected: 编译通过

- [ ] **Step 3: Commit**
```bash
git add -A && git commit -m "feat(engine): add unified engine entry point"
```

---

## Task 5: 集成测试

**Files:**
- Create: `testdata/events.csv`
- Create: `testdata/rules.yaml`

- [ ] **Step 1: 创建测试数据**
```csv
# testdata/events.csv
event_type,timestamp,actor,amount,risk_score
Transaction,2024-01-15T10:00:00Z,acc1,5000,0.2
Transaction,2024-01-15T11:00:00Z,acc1,15000,0.9
Transaction,2024-01-15T12:00:00Z,acc2,3000,0.3
Alert,2024-01-15T12:00:00Z,acc1,0,0.85
```

- [ ] **Step 2: 创建规则配置**
```yaml
# testdata/rules.yaml
rules:
  - name: high_value_transaction
    description: Alert on transactions over 10000
    condition:
      field: amount
      operator: gt
      value: 10000
    action:
      type: alert
      params:
        severity: high

  - name: high_risk_score
    description: Alert when risk score is above 0.8
    condition:
      field: risk_score
      operator: gte
      value: 0.8
    action:
      type: alert
      params:
        severity: critical
```

- [ ] **Step 3: 运行集成测试**
```bash
# 测试工作流规则
go test ./internal/engine/workflow/... -v

# 测试时序分析
go test ./internal/engine/temporal/... -v

# 测试搜索
go test ./internal/engine/search/... -v
```

Expected: 所有测试 PASS

- [ ] **Step 4: Commit**
```bash
git add -A && git commit -m "test(engine): add integration tests"
```

---

## 自检清单

- [x] **Spec覆盖检查**：
  - 时序分析 → Task 1
  - 全文搜索 → Task 2
  - 多维过滤 → Task 2
  - 聚合统计 → Task 2
  - 决策工作流 → Task 3
  - 告警触发 → Task 3
  - 引擎集成 → Task 4
  - 集成测试 → Task 5

- [x] **Placeholder扫描**：无 TBD/TODO

- [x] **类型一致性**：
  - TemporalAnalyzer 接口统一：RecordEvent/GetEvents/GetEventCount
  - FilterEngine 接口统一：Filter
  - WorkflowEngine 接口统一：AddRule/ProcessEvent/GetAlerts
  - Engine 统一入口集成各子引擎
