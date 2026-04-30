# 推背图（Back-Pushing）Phase 4 多端输出设计

> 日期：2026-04-30
> 状态：已批准

## 一、整体架构

```
back-pushing/
├── cmd/
│   ├── server/          # API 服务
│   │   └── main.go
│   └── cli/            # CLI（Phase 2）
├── api/                 # API 层
│   ├── graphql/        # GraphQL 端点
│   │   ├── schema.graphql
│   │   └── resolver.go
│   ├── rest/           # REST 端点
│   │   ├── router.go
│   │   ├── objects.go
│   │   ├── search.go
│   │   └── events.go
│   └── auth/           # 认证
│       └── jwt.go
├── web/                 # Web 控制台
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   └── services/
│   └── package.json
└── desktop/            # Electron 桌面应用
    ├── main.js
    ├── preload.js
    └── renderer/
```

---

## 二、实现阶段

### Phase 4-A: API 服务（优先级最高）

**功能：**
- REST + GraphQL 双协议支持
- 对象 CRUD、搜索、图遍历接口
- JWT 认证

**REST API 端点：**
- `GET /api/objects/:type/:id` - 获取对象
- `POST /api/objects` - 创建对象
- `PUT /api/objects/:type/:id` - 更新对象
- `DELETE /api/objects/:type/:id` - 删除对象
- `GET /api/objects/:type` - 列出对象
- `POST /api/search` - 搜索
- `POST /api/events` - 记录事件
- `GET /api/events/:type/:id` - 查询事件

**GraphQL API：**
- 查询：对象、图遍历、搜索结果
- 变更：创建、更新、删除对象，触发 Action
- 订阅：告警事件（可选）

### Phase 4-B: Web 控制台（优先级次之）

**功能：**
- React 前端
- 对象管理界面（增删改查）
- 图可视化（节点和边）
- Dashboard 仪表盘
- 实时告警查看

**页面结构：**
- 首页/Dashboard：统计面板、事件时间线
- 对象管理：列表、详情、编辑
- 图探索：可视化图遍历
- 告警中心：告警列表、状态管理

### Phase 4-C: Desktop 应用（优先级最后）

**功能：**
- Electron 桌面客户端
- 本地数据存储（SQLite）
- 离线操作能力
- 与后端 API 同步

**特性：**
- 跨平台（Windows/Mac/Linux）
- 系统托盘支持
- 本地缓存

---

## 三、技术选型

| 组件 | 技术 |
|------|------|
| REST API | Go + gin 框架 |
| GraphQL | gqlgen |
| 认证 | JWT |
| Web 前端 | React + D3.js |
| 桌面应用 | Electron |
| 本地存储 | SQLite |
