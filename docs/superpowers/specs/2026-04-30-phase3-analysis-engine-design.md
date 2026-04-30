# 推背图（Back-Pushing）Phase 3 分析引擎设计

> 日期：2026-04-30
> 状态：已批准

## 一、整体架构

```
internal/engine/
├── temporal/          # 时序分析引擎
│   ├── analyzer.go   # 事件流分析
│   ├── window.go     # 滚动窗口聚合
│   └── anomaly.go    # 异常检测
├── search/            # 实体搜索引擎
│   ├── fulltext.go   # 全文索引（倒排索引）
│   ├── filter.go     # 多维过滤
│   └── aggregate.go  # 聚合统计
├── workflow/         # 决策工作流引擎
│   ├── engine.go     # 工作流引擎核心
│   ├── rule.go       # 规则定义与评估
│   └── alert.go      # 告警管理
└── engine.go         # 统一入口，集成各引擎
```

### 设计原则

- **均衡方案**：核心功能完整实现，保留扩展点
- **与现有系统集成**：与 Phase 1 的存储和 Action 系统深度集成
- **YAML 配置兼容**：规则和工作流支持 YAML 配置定义

---

## 二、时序分析引擎

### 2.1 事件流分析 (analyzer.go)

**功能：**
- 记录事件（类型、时间戳、参与者、属性）
- 按对象类型和 ID 查询时间范围内的历史事件
- 事件计数统计

**接口：**
```go
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

func NewTemporalAnalyzer() *TemporalAnalyzer
func (a *TemporalAnalyzer) RecordEvent(ctx context.Context, event Event) error
func (a *TemporalAnalyzer) GetEvents(ctx context.Context, objType, objID string, since, until time.Time) ([]Event, error)
func (a *TemporalAnalyzer) GetEventCount(ctx context.Context, objType, objID string, since, until time.Time) (int64, error)
```

### 2.2 滚动窗口聚合 (window.go)

**功能：**
- 支持 Tumbling Window（不重叠）和 Sliding Window（重叠）
- 时间窗口内的聚合统计（Count / Sum / Avg / Min / Max）

**接口：**
```go
type Window struct {
    Size       time.Duration
    Slide      time.Duration
    WindowType WindowType  // tumbling / sliding
}

type Aggregation struct {
    Count int64
    Sum   float64
    Avg   float64
    Min   float64
    Max   float64
}

func (w *Window) Aggregate(events []Event, valueField string) *Aggregation
```

### 2.3 异常检测 (anomaly.go)

**功能：**
- 基于滑动窗口的统计异常检测
- Z-score 算法识别偏离正常范围的值
- 可配置阈值

**接口：**
```go
type AnomalyDetector struct {
    threshold float64  // Z-score 阈值
    window    int      // 参考窗口大小
}

type AnomalyResult struct {
    IsAnomaly bool
    Score     float64
    Message   string
}

func (d *AnomalyDetector) Detect(values []float64) AnomalyResult
```

---

## 三、实体搜索引擎

### 3.1 全文索引 (fulltext.go)

**功能：**
- 对文本字段建立倒排索引
- 简单分词（按字母/数字边界）
- 关键词搜索返回匹配对象 ID 列表

**接口：**
```go
type FullTextIndex struct {
    inverted map[string]map[string][]string  // term -> objType -> [objIDs]
}

func NewFullTextIndex() *FullTextIndex
func (idx *FullTextIndex) Index(ctx context.Context, objType, objID string, fields map[string]any) error
func (idx *FullTextIndex) Search(ctx context.Context, query string) ([]string, error)
```

### 3.2 多维过滤 (filter.go)

**功能：**
- 支持比较操作符：Eq / Gt / Lt / Gte / Lte / In / Contains
- 链式调用构建复杂过滤条件
- 对内存中的对象列表进行过滤

**接口：**
```go
type Filter struct {
    Eq  map[string]any
    Gt  map[string]any
    Lt  map[string]any
    Gte map[string]any
    Lte map[string]any
    In  map[string][]any
}

type FilterEngine struct {
    store *memory.ObjectStore
}

func (e *FilterEngine) Filter(ctx context.Context, objType string, filter *Filter) ([]map[string]any, error)
```

### 3.3 聚合统计 (aggregate.go)

**功能：**
- 对指定字段进行聚合计算（Count / Sum / Avg / Min / Max）
- 可选的 groupBy 分组
- 基于 ObjectStore 数据进行计算

**接口：**
```go
type AggregationResult struct {
    Count  int64
    Sum    map[string]float64
    Avg    map[string]float64
    Min    map[string]float64
    Max    map[string]float64
    Groups map[string][]map[string]any
}

func (e *AggregateEngine) Aggregate(ctx context.Context, objType string, fields []string, groupBy string) (*AggregationResult, error)
```

---

## 四、决策工作流引擎

### 4.1 规则定义 (rule.go)

**功能：**
- 定义条件规则（字段 + 操作符 + 值）
- 支持操作符：eq / ne / gt / lt / gte / lte / contains
- 批量评估数据是否匹配规则

**接口：**
```go
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

func (r *Rule) Evaluate(data map[string]any) bool
func (rs *RuleSet) Add(rule Rule)
func (rs *RuleSet) Evaluate(data map[string]any) []Rule
```

### 4.2 告警管理 (alert.go)

**功能：**
- 创建和存储告警记录
- 注册告警处理函数（webhook / 日志 / 回调）
- 查询告警历史

**接口：**
```go
type Alert struct {
    ID        string
    RuleName  string
    Severity  string
    Message   string
    Timestamp time.Time
    Data      map[string]any
}

type AlertManager struct {
    alerts   []Alert
    handlers []AlertHandler
}

func (m *AlertManager) RegisterHandler(handler AlertHandler)
func (m *AlertManager) Trigger(ruleName, severity, message string, data map[string]any)
func (m *AlertManager) GetAlerts() []Alert
```

### 4.3 工作流引擎 (engine.go)

**功能：**
- 接收事件并触发规则评估
- 规则匹配后执行相应动作（alert / workflow / record_event）
- 与 TemporalAnalyzer 集成实现事件记录

**接口：**
```go
type WorkflowEngine struct {
    rules    *RuleSet
    alerts   *AlertManager
    temporal *TemporalAnalyzer
}

func NewWorkflowEngine(temporal *TemporalAnalyzer) *WorkflowEngine
func (e *WorkflowEngine) AddRule(rule Rule)
func (e *WorkflowEngine) ProcessEvent(ctx context.Context, eventType string, data map[string]any) error
func (e *WorkflowEngine) GetAlerts() []Alert
```

---

## 五、引擎统一入口

### 5.1 统一引擎 (engine.go)

**功能：**
- 整合所有子引擎（Temporal / Search / Workflow）
- 共享 ObjectStore 和 GraphStore
- 提供统一初始化接口

**接口：**
```go
type Engine struct {
    ObjectStore *memory.ObjectStore
    GraphStore  *memory.GraphStore
    Temporal    *TemporalAnalyzer
    Search      *SearchEngine
    Workflow    *WorkflowEngine
}

func NewEngine(objStore *memory.ObjectStore, graphStore *memory.GraphStore) *Engine
func (e *Engine) ProcessEvent(ctx context.Context, eventType string, data map[string]any) error
```

---

## 六、数据流

```
数据导入 → 规则评估 → [触发告警] → [触发工作流]
                    ↓
              时序记录 ← 查询/聚合 ← 搜索请求
```

---

## 七、与 Phase 1/2 的关系

| Phase | 提供 |
|-------|------|
| Phase 1 | Ontology、对象存储、图遍历、Action 系统 |
| Phase 2 | 数据接入、Mapper、CLI |
| Phase 3 | 时序分析、全文搜索、规则引擎、告警管理 |

**集成点：**
- WorkflowEngine 使用 ObjectStore 进行事件记录
- WorkflowEngine 的 Action 可调用 Phase 1 的 Action Handler
- Engine 统一入口由 cmd/server 初始化

---

## 八、测试策略

- 单元测试：每个子包独立测试核心逻辑
- 集成测试：Engine 统一入口测试
- 使用内存存储，无需外部依赖
