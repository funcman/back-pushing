# 推背图（Back-Pushing）— Palantir原型架构设计

> 日期：2026-04-29
> 状态：已批准

## 一、整体架构

```
┌─────────────────────────────────────────────────────────┐
│                      前端层（多端）                       │
│   Web控制台  │  CLI工具  │  Desktop  │  API客户端        │
└──────────┬──────────────────────────────────────────────┘
           │ HTTP/gRPC
┌──────────▼──────────────────────────────────────────────┐
│                    API 服务层                            │
│              GraphQL/REST + Action Invocation            │
│              认证 │ 鉴权 │ 审计日志                       │
└──────────┬──────────────────────────────────────────────┘
           │
┌──────────▼──────────────────────────────────────────────┐
│                   Ontology 运行时                         │
│  Object Registry │ Link Manager │ Path Resolver          │
│  Action Dispatcher │ Classification Engine              │
│  Computed Property Evaluator                             │
└──────────┬──────────────────────────────────────────────┘
           │
┌──────────▼──────────────────────────────────────────────┐
│                   混合存储层                             │
│  图存储（实体关系）│ 列存储（事件时序）│ KV（热点/缓存）    │
│  SQLite/LevelDB（原型）                                   │
└─────────────────────────────────────────────────────────┘
```

### 核心设计原则

- **API驱动**：所有前端共享同一数据引擎
- **Ontology作为核心骨架**：YAML声明式定义
- **适配器模式**：统一异构数据源接入
- **混合存储**：按查询场景分配最优存储

---

## 二、Ontology 核心设计

### 2.1 三种核心建模对象

| 类型 | 说明 | 示例 |
|------|------|------|
| Entity | 静态实体，有相对稳定的属性 | Person, Organization, Device |
| Event | 时间相关的动态记录 | Transaction, LoginEvent, Alert |
| Concept | 抽象类别，用于分类和聚合 | RiskLevel, Status, Category |

### 2.2 对象定义示例

```yaml
# ontology/person.yaml
object_types:
  Person:
    description: 自然人实体
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
      risk_score:
        type: float
        computed: true
        source: "RISK_ASSESSMENT.person_risk(person_id=id)"
      created_at:
        type: datetime
      tags:
        type: list[string]
        indexed: true

    links:
      works_at:
        target: Organization
        through: Employment
        reverse: employees
      knows:
        target: Person
        reverse: known_by

    actions:
      - name: "Escalate to Review"
        description: "提交给风控部门审查"
        handler: action.escalate_review
        args:
          reason: string
          priority: enum[low, medium, high]
      - name: "Flag as Suspicious"
        description: "标记为可疑人员"
        handler: action.flag_suspicious
        args:
          reason: string
          evidence: list[string]

    paths:
      reporting_chain:
        description: "员工汇报链"
        steps:
          - Person-[reports_to]->Person
      organizational_affiliation:
        description: "组织归属链"
        steps:
          - Person-[works_at]->Organization
          - Organization-[located_in]->Country
```

### 2.3 事件对象

```yaml
object_types:
  Transaction:
    description: 交易事件
    type: event
    properties:
      id:
        type: string
        primary: true
      timestamp:
        type: datetime
        indexed: true
      amount:
        type: decimal
      currency:
        type: string
        default: "CNY"
      parties:
        type: list[Person | Organization]
      risk_indicators:
        type: list[string]
        computed: true
        source: "RISK_SCORING.transaction_risk(tx_id=id)"

    temporal:
      window: 90d
      retention: 7y

    actions:
      - name: "Block Transaction"
        handler: action.block_transaction
      - name: "Trigger Enhanced Due Diligence"
        handler: action.edd_workflow
```

### 2.4 关系（Link）定义

```yaml
links:
  Employment:
    description: 雇佣关系
    source: Person
    target: Organization
    type: many-to-many
    properties:
      role: string
      department: string
      since: date
      until: date | null
      active: boolean
        computed: "until == null || until > now()"

    actions:
      - name: "Terminate"
        handler: action.terminate_employment
      - name: "Transfer"
        handler: action.transfer_employee

  Person_knows_Person:
    source: Person
    target: Person
    type: many-to-many
    properties:
      weight: float
      source: string
      verified: boolean
      since: datetime

    actions:
      - name: "Remove Connection"
        handler: action.remove_connection
```

### 2.5 路径（Path）预定义

```yaml
paths:
  money_flow:
    description: "追踪资金流向"
    steps:
      - Person-[transferred]->Transaction
      - Transaction-[transferred_to]->Person

  organizational_tree:
    description: "组织架构树"
    steps:
      - Organization-[parent]->Organization
      - Organization-[has_employee]->Person

  risk_exposure:
    description: "风险暴露路径"
    steps:
      - Person-[transferred]->Transaction
      - Transaction-[involves]->Organization
      - Organization-[has_sanction]->SanctionList
```

### 2.6 分类标签（Classification）

```yaml
classification:
  levels:
    - public
    - internal
    - confidential
    - restricted

  data_handling:
    PII:
      description: "个人身份信息"
      actions:
        - mask_on_export
        - audit_on_access
    FINANCIAL:
      description: "财务敏感数据"
      actions:
        - require_approval
        - log_all_access

  object_tags:
    Person:
      sensitivity: confidential
      handling: PII
    Transaction:
      sensitivity: restricted
      handling: FINANCIAL
```

---

## 三、Action Handler 框架

### 3.1 Action 定义

每个 Action 对应一个 Handler 函数，支持：
- 参数校验
- 事件记录
- 工作流触发
- 实体状态更新

### 3.2 Handler 接口

```go
type ActionContext struct {
    ObjectStore  ObjectStore
    EventLogger  EventLogger
    Workflow     WorkflowEngine
    AuditLog     AuditLogger
}

type ActionHandler func(ctx *ActionContext, input any) (output any, err error)
```

### 3.3 示例实现

```go
func EscalateReview(ctx *ActionContext, input EscalateReviewInput) (*EscalateReviewOutput, error) {
    caseID := generateCaseID()

    event := &Event{
        Type:      "REVIEW_ESCALATION",
        Timestamp: time.Now(),
        Payload:   input,
    }
    ctx.LogEvent(event)

    ctx.TriggerWorkflow("review_escalation", map[string]any{
        "case_id":  caseID,
        "person_id": input.PersonID,
        "priority":  input.Priority,
    })

    ctx.UpdateObject("Person", input.PersonID, map[string]any{
        "review_status":    "escalated",
        "last_escalation": time.Now(),
    })

    return &EscalateReviewOutput{
        CaseID:     caseID,
        AssignedTo: "risk_team",
        Status:     "pending",
    }, nil
}
```

---

## 四、分析引擎

### 4.1 图探索引擎

- 链路分析
- 路径查找
- 异常关联发现
- 社区检测

### 4.2 时序分析引擎

- 趋势分析
- 异常检测
- 滚动窗口聚合
- 多维度时序切片

### 4.3 实体搜索引擎

- 全文搜索
- 过滤条件组合
- 聚合统计
- 分页排序

### 4.4 决策引擎

- 规则匹配
- 告警触发
- 行动建议
- 工作流编排

---

## 五、技术选型

| 层级 | 技术 |
|------|------|
| 核心语言 | Go |
| 协议 | HTTP + gRPC |
| API | GraphQL + REST |
| Ontology定义 | YAML |
| 图存储 | 插件化（原型阶段用内存/LevelDB）|
| 列存储 | 插件化（原型阶段用内存）|
| KV存储 | LevelDB / 内存 |
| 前端 | Web + CLI + Desktop + API Client |

---

## 六、项目阶段规划

### Phase 1: 最小可行核心
- Ontology Core（YAML解析 + 对象注册）
- 内存存储层
- 图探索（基础遍历）

### Phase 2: 数据接入
- 适配器框架
- 至少一种数据源连接器
- 数据导入/同步

### Phase 3: 完整分析能力
- 时序分析
- 实体搜索
- 决策工作流

### Phase 4: 多端输出
- API服务
- Web控制台
- CLI工具
- Desktop应用

---

## 七、模块职责

| 模块 | 职责 |
|------|------|
| `ontology/` | YAML解析、对象注册、Link管理、Path解析 |
| `storage/` | 混合存储抽象、图/列/KV插件 |
| `adapter/` | 数据源适配器框架 |
| `engine/` | 图探索、时序、搜索、决策引擎 |
| `action/` | Action Handler实现 |
| `api/` | GraphQL/REST接口 |
| `cli/` | 命令行工具 |
| `web/` | Web控制台 |
| `desktop/` | Electron桌面应用 |
