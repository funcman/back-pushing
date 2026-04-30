# 推背图 (Back-Pushing)

> Palantir原型架构实现 - Ontology驱动的数据映射与分析平台

## 项目概述

推背图是一个基于Ontology的数据映射与分析平台，支持从多种数据源（JSON/CSV/SQL）导入数据，通过YAML配置的映射规则转换为统一的Object模型，并提供完整的分析能力。

## 技术架构

```
┌─────────────────────────────────────────────────────────┐
│                      前端层（多端）                       │
│   Web控制台  │  CLI工具  │  Desktop  │  API客户端        │
└──────────┬──────────────────────────────────────────────┘
           │
┌──────────▼──────────────────────────────────────────────┐
│                    API 服务层                            │
│              REST API + GraphQL + JWT认证                │
└──────────┬──────────────────────────────────────────────┘
           │
┌──────────▼──────────────────────────────────────────────┐
│                   Ontology 运行时                         │
│  Object Registry │ Link Manager │ Path Resolver          │
│  Action Dispatcher │ Classification Engine              │
└──────────┬──────────────────────────────────────────────┘
           │
┌──────────▼──────────────────────────────────────────────┐
│                   混合存储层                              │
│     Object Store  │  Graph Store  │  Temporal Engine    │
└─────────────────────────────────────────────────────────┘
```

## 项目阶段

| Phase | 内容 | 状态 |
|-------|------|------|
| Phase 1 | 最小可行核心 - Ontology解析、内存存储、图遍历 | ✅ |
| Phase 2 | 数据接入层 - JSON/CSV/SQL适配器、Mapper、CLI | ✅ |
| Phase 3 | 完整分析能力 - 时序分析、搜索、工作流引擎 | ✅ |
| Phase 4 | 多端输出 - REST/GraphQL API、Web控制台、Desktop | ✅ |

## 技术栈

| 层级 | 技术 |
|------|------|
| 核心语言 | Go 1.21+ |
| API | REST + GraphQL + JWT |
| Web | React + TypeScript + D3.js |
| Desktop | Electron |
| 数据格式 | YAML |

## 快速开始

### 前置要求

- Go 1.21+
- Node.js 18+ (用于Web和Desktop)
- MySQL (可选，用于SQL数据源)

### 构建

```bash
# 构建所有组件
make build-all

# 仅构建后端
make build-server

# 仅构建CLI
make build-cli

# 仅构建Web
make build-web

# 仅构建Desktop
make build-desktop
```

### 运行

```bash
# 运行API服务器
make run

# 运行CLI
make run-cli

# 运行Web开发服务器
cd web && npm run dev
```

### 测试

```bash
make test
```

## 目录结构

```
back-pushing/
├── cmd/
│   ├── server/           # API服务器入口
│   └── cli/              # CLI工具入口
├── api/
│   ├── graphql/          # GraphQL API
│   ├── rest/             # REST API
│   └── auth/             # JWT认证
├── internal/
│   ├── ontology/          # Ontology解析与注册
│   ├── storage/
│   │   └── memory/       # 内存存储
│   ├── adapter/          # 数据源适配器
│   │   ├── json/         # JSON适配器
│   │   ├── csv/          # CSV适配器
│   │   └── sql/          # SQL适配器
│   ├── mapper/           # 数据映射引擎
│   ├── engine/
│   │   ├── graph/        # 图遍历引擎
│   │   ├── temporal/     # 时序分析引擎
│   │   ├── search/       # 全文搜索引擎
│   │   └── workflow/     # 工作流引擎
│   └── action/           # Action调度
├── web/                   # React Web控制台
├── desktop/              # Electron桌面应用
├── mapping/              # 映射配置文件
├── testdata/             # 测试数据
└── docs/                 # 设计文档和计划
```

## 使用示例

### 导入数据

```bash
# 使用CLI导入CSV数据
./bin/cli --mapping ./mapping/person导入.yaml

# 使用CLI导入JSON数据
./bin/cli --mapping ./mapping/transaction导入.yaml
```

### 映射配置示例

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

## API端点

### REST API

| Method | Endpoint | 描述 |
|--------|----------|------|
| GET | /api/objects/:type/:id | 获取对象 |
| POST | /api/objects | 创建对象 |
| PUT | /api/objects/:type/:id | 更新对象 |
| DELETE | /api/objects/:type/:id | 删除对象 |
| GET | /api/objects/:type | 列出对象 |
| POST | /api/search | 搜索 |
| POST | /api/events | 记录事件 |
| GET | /api/events/:type/:id | 查询事件 |

### GraphQL

```graphql
query {
  objects(type: "Person") {
    id
    type
    data
  }
}

mutation {
  createObject(type: "Person", id: "p1", data: {name: "Alice"}) {
    id
    type
  }
}
```

## 开发指南

### 添加新的数据适配器

1. 在 `internal/adapter/` 下创建新目录
2. 实现 `DataSource` 接口
3. 在 `internal/cli/import.go` 中添加新适配器支持

### 添加新的Action

1. 在 `internal/action/` 中实现Handler
2. 在 `cmd/server/main.go` 中注册Action

## 文档

- [设计文档](docs/superpowers/specs/2026-04-29-back-pushing-design.md)
- [Phase 1 计划](docs/superpowers/plans/2026-04-29-back-pushing-phase1.md)
- [Phase 2 计划](docs/superpowers/plans/2026-04-29-back-pushing-phase2.md)
- [Phase 3 计划](docs/superpowers/plans/2026-04-30-back-pushing-phase3.md)
- [Phase 4 计划](docs/superpowers/plans/2026-04-30-back-pushing-phase4.md)

## 许可证

MIT
