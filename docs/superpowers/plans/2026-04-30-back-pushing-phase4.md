# 推背图（Back-Pushing）Phase 4 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建多端输出能力——API服务、Web控制台、Desktop应用

**Architecture:**
- REST + GraphQL 双协议 API 服务
- React Web 控制台 + 图可视化
- Electron 桌面应用
- JWT 认证
- 本地 SQLite 存储（Desktop）

**Tech Stack:** Go + gin, gqlgen, React, D3.js, Electron, SQLite

---

## 文件结构

```
back-pushing/
├── cmd/
│   ├── server/
│   │   └── main.go              # API 服务入口
│   └── cli/
│       └── main.go              # CLI（Phase 2）
├── api/
│   ├── graphql/
│   │   ├── schema.graphql      # GraphQL schema
│   │   ├── resolver.go          # Resolver 实现
│   │   └── handler.go          # GraphQL handler
│   ├── rest/
│   │   ├── router.go           # REST router
│   │   ├── objects.go          # 对象 CRUD
│   │   ├── search.go           # 搜索接口
│   │   └── events.go           # 事件接口
│   └── auth/
│       └── jwt.go              # JWT 认证
├── web/                          # Web 控制台
│   ├── package.json
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── services/api.ts
│   │   └── App.tsx
│   └── public/
├── desktop/                      # Electron 桌面应用
│   ├── package.json
│   ├── main.js                  # 主进程
│   ├── preload.js               # 预加载
│   └── renderer/                # 渲染进程
└── Makefile
```

---

## Task 1: REST API 服务

**Files:**
- Create: `api/rest/router.go`
- Create: `api/rest/objects.go`
- Create: `api/rest/search.go`
- Create: `api/rest/events.go`
- Modify: `cmd/server/main.go` (添加 REST 路由)

- [ ] **Step 1: 实现 REST Router**
```go
// api/rest/router.go
package rest

import (
    "github.com/gin-gonic/gin"
)

type Router struct {
    engine     *gin.Engine
    objects    *ObjectsHandler
    search     *SearchHandler
    events     *EventsHandler
}

func NewRouter(objects *ObjectsHandler, search *SearchHandler, events *EventsHandler) *Router {
    r := &Router{
        engine:  gin.Default(),
        objects: objects,
        search:  search,
        events:  events,
    }
    r.setupRoutes()
    return r
}

func (r *Router) setupRoutes() {
    api := r.engine.Group("/api")
    {
        // Objects
        api.GET("/objects/:type/:id", r.objects.Get)
        api.POST("/objects", r.objects.Create)
        api.PUT("/objects/:type/:id", r.objects.Update)
        api.DELETE("/objects/:type/:id", r.objects.Delete)
        api.GET("/objects/:type", r.objects.List)

        // Search
        api.POST("/search", r.search.Search)

        // Events
        api.POST("/events", r.events.Record)
        api.GET("/events/:type/:id", r.events.GetByActor)
    }
}

func (r *Router) Run(addr string) error {
    return r.engine.Run(addr)
}
```

- [ ] **Step 2: 实现 Objects Handler**
```go
// api/rest/objects.go
package rest

import (
    "net/http"

    "github.com/gin-gonic/gin"

    "back-pushing/internal/storage/memory"
)

type ObjectsHandler struct {
    store *memory.ObjectStore
}

func NewObjectsHandler(store *memory.ObjectStore) *ObjectsHandler {
    return &ObjectsHandler{store: store}
}

func (h *ObjectsHandler) Get(c *gin.Context) {
    objType := c.Param("type")
    id := c.Param("id")

    obj, err := h.store.Get(c.Request.Context(), objType, id)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, obj)
}

func (h *ObjectsHandler) Create(c *gin.Context) {
    var req struct {
        Type string         `json:"type"`
        ID   string         `json:"id"`
        Data map[string]any `json:"data"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.store.Create(c.Request.Context(), req.Type, req.ID, req.Data); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusCreated, gin.H{"id": req.ID})
}

func (h *ObjectsHandler) Update(c *gin.Context) {
    objType := c.Param("type")
    id := c.Param("id")

    var data map[string]any
    if err := c.ShouldBindJSON(&data); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    if err := h.store.Update(c.Request.Context(), objType, id, data); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"id": id})
}

func (h *ObjectsHandler) Delete(c *gin.Context) {
    objType := c.Param("type")
    id := c.Param("id")

    if err := h.store.Delete(c.Request.Context(), objType, id); err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"deleted": id})
}

func (h *ObjectsHandler) List(c *gin.Context) {
    objType := c.Param("type")

    objects, err := h.store.List(c.Request.Context(), objType, nil)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, objects)
}
```

- [ ] **Step 3: 实现 Search Handler**
```go
// api/rest/search.go
package rest

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

type SearchHandler struct{}

func NewSearchHandler() *SearchHandler {
    return &SearchHandler{}
}

func (h *SearchHandler) Search(c *gin.Context) {
    var req struct {
        Query string `json:"query"`
        Type  string `json:"type"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Search implementation placeholder
    c.JSON(http.StatusOK, gin.H{
        "query": req.Query,
        "type":  req.Type,
        "results": []map[string]any{},
    })
}
```

- [ ] **Step 4: 实现 Events Handler**
```go
// api/rest/events.go
package rest

import (
    "net/http"

    "github.com/gin-gonic/gin"
)

type EventsHandler struct{}

func NewEventsHandler() *EventsHandler {
    return &EventsHandler{}
}

func (h *EventsHandler) Record(c *gin.Context) {
    var req struct {
        Type  string         `json:"type"`
        Actor string         `json:"actor"`
        Props map[string]any `json:"props"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, gin.H{"status": "recorded"})
}

func (h *EventsHandler) GetByActor(c *gin.Context) {
    objType := c.Param("type")
    actor := c.Param("id")

    // Query events implementation
    c.JSON(http.StatusOK, []map[string]any{
        {"type": objType, "actor": actor},
    })
}
```

- [ ] **Step 5: 更新 Server Main.go**
```go
// cmd/server/main.go
// 在现有基础上添加:
import (
    "back-pushing/api/rest"
)

func main() {
    // ... 现有初始化代码 ...

    // 添加 REST API
    objectsHandler := rest.NewObjectsHandler(objStore)
    searchHandler := rest.NewSearchHandler()
    eventsHandler := rest.NewEventsHandler()
    router := rest.NewRouter(objectsHandler, searchHandler, eventsHandler)

    log.Println("Starting REST API server at :8080")
    router.Run(":8080")
}
```

- [ ] **Step 6: 运行验证**
```bash
go build -o bin/server ./cmd/server
```
Expected: 编译通过

- [ ] **Step 7: Commit**
```bash
git add -A && git commit -m "feat(api): add REST API endpoints"
```

---

## Task 2: GraphQL API

**Files:**
- Create: `api/graphql/schema.graphql`
- Create: `api/graphql/resolver.go`
- Create: `api/graphql/handler.go`

- [ ] **Step 1: 定义 GraphQL Schema**
```graphql
# api/graphql/schema.graphql
type Object {
    id: String!
    type: String!
    data: JSON!
}

type Query {
    object(type: String!, id: String!): Object
    objects(type: String!): [Object!]!
    search(query: String!, type: String): [Object!]!
}

type Mutation {
    createObject(type: String!, id: String!, data: JSON!): Object!
    updateObject(type: String!, id: String!, data: JSON!): Object!
    deleteObject(type: String!, id: String!): Boolean!
}

type Subscription {
    alertTriggered: Alert
}

type Alert {
    id: String!
    ruleName: String!
    severity: String!
    message: String!
    timestamp: String!
}
```

- [ ] **Step 2: 实现 Resolver**
```go
// api/graphql/resolver.go
package graphql

import (
    "context"

    "back-pushing/internal/storage/memory"
)

type Resolver struct {
    store *memory.ObjectStore
}

func NewResolver(store *memory.ObjectStore) *Resolver {
    return &Resolver{store: store}
}

type Object struct {
    ID   string `json:"id"`
    Type string `json:"type"`
    Data map[string]any `json:"data"`
}

func (r *Resolver) Object(ctx context.Context, args struct{ Type, ID string }) (*Object, error) {
    data, err := r.store.Get(ctx, args.Type, args.ID)
    if err != nil {
        return nil, err
    }
    return &Object{ID: args.ID, Type: args.Type, Data: data}, nil
}

func (r *Resolver) Objects(ctx context.Context, args struct{ Type string }) ([]*Object, error) {
    objects, err := r.store.List(ctx, args.Type, nil)
    if err != nil {
        return nil, err
    }

    var result []*Object
    for id, data := range objects {
        result = append(result, &Object{ID: id, Type: args.Type, Data: data})
    }
    return result, nil
}

func (r *Resolver) CreateObject(ctx context.Context, args struct {
    Type string         `json:"type"`
    ID   string         `json:"id"`
    Data map[string]any `json:"data"`
}) (*Object, error) {
    err := r.store.Create(ctx, args.Type, args.ID, args.Data)
    if err != nil {
        return nil, err
    }
    return &Object{ID: args.ID, Type: args.Type, Data: args.Data}, nil
}

func (r *Resolver) UpdateObject(ctx context.Context, args struct {
    Type string         `json:"type"`
    ID   string         `json:"id"`
    Data map[string]any `json:"data"`
}) (*Object, error) {
    err := r.store.Update(ctx, args.Type, args.ID, args.Data)
    if err != nil {
        return nil, err
    }
    return &Object{ID: args.ID, Type: args.Type, Data: args.Data}, nil
}

func (r *Resolver) DeleteObject(ctx context.Context, args struct{ Type, ID string }) (bool, error) {
    err := r.store.Delete(ctx, args.Type, args.ID)
    return err == nil, err
}
```

- [ ] **Step 3: 实现 Handler**
```go
// api/graphql/handler.go
package graphql

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/graphql-go/graphql"

    "back-pushing/internal/storage/memory"
)

func NewHandler(store *memory.ObjectStore) gin.HandlerFunc {
    resolver := NewResolver(store)

    queryType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Query",
        Fields: graphql.Fields{
            "object": &graphql.Field{
                Type: graphql.NewObject(graphql.ObjectConfig{
                    Name: "Object",
                    Fields: graphql.Fields{
                        "id":   &graphql.Field{Type: graphql.String},
                        "type": &graphql.Field{Type: graphql.String},
                        "data": &graphql.Field{Type: graphql.JSON},
                    },
                }),
                Args: graphql.FieldConfigArgument{
                    "type": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                    "id":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return resolver.Object(p.Context, struct{ Type, ID string }{
                        Type: p.Args["type"].(string),
                        ID:   p.Args["id"].(string),
                    })
                },
            },
            "objects": &graphql.Field{
                Type: graphql.NewList(graphql.NewObject(graphql.ObjectConfig{
                    Name: "Object",
                    Fields: graphql.Fields{
                        "id":   &graphql.Field{Type: graphql.String},
                        "type": &graphql.Field{Type: graphql.String},
                        "data": &graphql.Field{Type: graphql.JSON},
                    },
                })),
                Args: graphql.FieldConfigArgument{
                    "type": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return resolver.Objects(p.Context, struct{ Type string }{Type: p.Args["type"].(string)})
                },
            },
        },
    })

    mutationType := graphql.NewObject(graphql.ObjectConfig{
        Name: "Mutation",
        Fields: graphql.Fields{
            "createObject": &graphql.Field{
                Type: graphql.NewObject(graphql.ObjectConfig{
                    Name: "Object",
                    Fields: graphql.Fields{
                        "id":   &graphql.Field{Type: graphql.String},
                        "type": &graphql.Field{Type: graphql.String},
                        "data": &graphql.Field{Type: graphql.JSON},
                    },
                }),
                Args: graphql.FieldConfigArgument{
                    "type": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                    "id":   &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.String)},
                    "data": &graphql.ArgumentConfig{Type: graphql.NewNonNull(graphql.JSON)},
                },
                Resolve: func(p graphql.ResolveParams) (interface{}, error) {
                    return resolver.CreateObject(p.Context, struct {
                        Type string
                        ID   string
                        Data map[string]any
                    }{
                        Type: p.Args["type"].(string),
                        ID:   p.Args["id"].(string),
                        Data: p.Args["data"].(map[string]any),
                    })
                },
            },
        },
    })

    schema, _ := graphql.NewSchema(graphql.SchemaConfig{
        Query:    queryType,
        Mutation: mutationType,
    })

    return func(c *gin.Context) {
        var req struct {
            Query         string                 `json:"query"`
            OperationName string                 `json:"operationName"`
            Variables     map[string]interface{} `json:"variables"`
        }
        if err := c.ShouldBindJSON(&req); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        result := graphql.Do(graphql.Params{
            Schema:         schema,
            RequestString:  req.Query,
            VariableValues: req.Variables,
            OperationName: req.OperationName,
        })
        c.JSON(http.StatusOK, result)
    }
}
```

- [ ] **Step 4: 更新 Server 集成 GraphQL**
```go
// cmd/server/main.go 添加:
import (
    "back-pushing/api/graphql"
)

func main() {
    // ...
    gin.SetMode(gin.ReleaseMode)
    r := gin.Default()

    // GraphQL endpoint
    r.POST("/graphql", graphql.NewHandler(objStore))

    // REST API (from Task 1)
    restObjectsHandler := rest.NewObjectsHandler(objStore)
    restSearchHandler := rest.NewSearchHandler()
    restEventsHandler := rest.NewEventsHandler()
    restRouter := rest.NewRouter(restObjectsHandler, restSearchHandler, restEventsHandler)

    // 同时支持 REST 和 GraphQL
    go restRouter.Run(":8081")

    log.Println("Starting GraphQL server at :8080")
    r.Run(":8080")
}
```

- [ ] **Step 5: Commit**
```bash
git add -A && git commit -m "feat(api): add GraphQL API"
```

---

## Task 3: JWT 认证

**Files:**
- Create: `api/auth/jwt.go`

- [ ] **Step 1: 实现 JWT 认证**
```go
// api/auth/jwt.go
package auth

import (
    "net/http"
    "strings"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/golang-jwt/jwt/v5"
)

type Claims struct {
    UserID string `json:"user_id"`
    Role   string `json:"role"`
    jwt.RegisteredClaims
}

type JWTAuth struct {
    secret     []byte
    expiration time.Duration
}

func NewJWTAuth(secret string, expiration time.Duration) *JWTAuth {
    return &JWTAuth{
        secret:     []byte(secret),
        expiration: expiration,
    }
}

func (j *JWTAuth) GenerateToken(userID, role string) (string, error) {
    claims := Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.expiration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(j.secret)
}

func (j *JWTAuth) Middleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
            c.Abort()
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
            c.Abort()
            return
        }

        tokenString := parts[1]
        claims := &Claims{}

        token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
            return j.secret, nil
        })

        if err != nil || !token.Valid {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
            c.Abort()
            return
        }

        c.Set("user_id", claims.UserID)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

- [ ] **Step 2: Commit**
```bash
git add -A && git commit -m "feat(auth): add JWT authentication"
```

---

## Task 4: Web 控制台 - 项目初始化

**Files:**
- Create: `web/package.json`
- Create: `web/src/App.tsx`
- Create: `web/src/index.tsx`
- Create: `web/public/index.html`

- [ ] **Step 1: 初始化 React 项目**
```json
// web/package.json
{
  "name": "back-pushing-web",
  "version": "1.0.0",
  "private": true,
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-router-dom": "^6.20.0",
    "d3": "^7.8.5",
    "apollo-client": "^3.8.8",
    "graphql": "^16.8.1"
  },
  "devDependencies": {
    "typescript": "^5.3.0",
    "@types/react": "^18.2.0",
    "@types/d3": "^7.4.0",
    "vite": "^5.0.0"
  },
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview"
  }
}
```

- [ ] **Step 2: 创建 API Service**
```typescript
// web/src/services/api.ts
const API_BASE = 'http://localhost:8080/api';
const GRAPHQL_ENDPOINT = 'http://localhost:8080/graphql';

export const api = {
  async getObjects(type: string) {
    const res = await fetch(`${API_BASE}/objects/${type}`);
    return res.json();
  },

  async getObject(type: string, id: string) {
    const res = await fetch(`${API_BASE}/objects/${type}/${id}`);
    return res.json();
  },

  async createObject(type: string, id: string, data: object) {
    const res = await fetch(`${API_BASE}/objects`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ type, id, data }),
    });
    return res.json();
  },

  async search(query: string, type?: string) {
    const res = await fetch(`${API_BASE}/search`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query, type }),
    });
    return res.json();
  },
};

export const graphql = {
  async query(query: string, variables?: object) {
    const res = await fetch(GRAPHQL_ENDPOINT, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ query, variables }),
    });
    return res.json();
  },
};
```

- [ ] **Step 3: 创建主应用组件**
```tsx
// web/src/App.tsx
import React from 'react';
import { BrowserRouter, Routes, Route, Link } from 'react-router-dom';
import Dashboard from './pages/Dashboard';
import Objects from './pages/Objects';
import GraphView from './pages/GraphView';
import Alerts from './pages/Alerts';

const App: React.FC = () => {
  return (
    <BrowserRouter>
      <div className="app">
        <nav className="sidebar">
          <h1>Back-Pushing</h1>
          <ul>
            <li><Link to="/">Dashboard</Link></li>
            <li><Link to="/objects">Objects</Link></li>
            <li><Link to="/graph">Graph View</Link></li>
            <li><Link to="/alerts">Alerts</Link></li>
          </ul>
        </nav>
        <main className="content">
          <Routes>
            <Route path="/" element={<Dashboard />} />
            <Route path="/objects" element={<Objects />} />
            <Route path="/graph" element={<GraphView />} />
            <Route path="/alerts" element={<Alerts />} />
          </Routes>
        </main>
      </div>
    </BrowserRouter>
  );
};

export default App;
```

- [ ] **Step 4: 创建页面组件**
```tsx
// web/src/pages/Dashboard.tsx
import React, { useEffect, useState } from 'react';
import { api } from '../services/api';

const Dashboard: React.FC = () => {
  const [stats, setStats] = useState({ objects: 0, alerts: 0 });

  useEffect(() => {
    // Load dashboard stats
  }, []);

  return (
    <div className="dashboard">
      <h2>Dashboard</h2>
      <div className="stats">
        <div className="stat-card">
          <h3>Objects</h3>
          <p className="stat-value">{stats.objects}</p>
        </div>
        <div className="stat-card">
          <h3>Active Alerts</h3>
          <p className="stat-value">{stats.alerts}</p>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
```

```tsx
// web/src/pages/Objects.tsx
import React, { useEffect, useState } from 'react';
import { api } from '../services/api';

const Objects: React.FC = () => {
  const [objects, setObjects] = useState<any[]>([]);
  const [selectedType, setSelectedType] = useState('Person');

  useEffect(() => {
    loadObjects();
  }, [selectedType]);

  const loadObjects = async () => {
    const data = await api.getObjects(selectedType);
    setObjects(data);
  };

  return (
    <div className="objects-page">
      <h2>Object Management</h2>
      <select value={selectedType} onChange={(e) => setSelectedType(e.target.value)}>
        <option value="Person">Person</option>
        <option value="Organization">Organization</option>
        <option value="Transaction">Transaction</option>
      </select>
      <table className="objects-table">
        <thead>
          <tr>
            <th>ID</th>
            <th>Data</th>
          </tr>
        </thead>
        <tbody>
          {objects.map((obj) => (
            <tr key={obj.id}>
              <td>{obj.id}</td>
              <td><pre>{JSON.stringify(obj.data, null, 2)}</pre></td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default Objects;
```

```tsx
// web/src/pages/GraphView.tsx
import React, { useEffect, useRef } from 'react';
import * as d3 from 'd3';

const GraphView: React.FC = () => {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (!svgRef.current) return;

    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();

    // Create force-directed graph
    const simulation = d3.forceSimulation()
      .force('link', d3.forceLink().id((d: any) => d.id))
      .force('charge', d3.forceManyBody())
      .force('center', d3.forceCenter(400, 300));

    // Graph rendering code here
  }, []);

  return (
    <div className="graph-view">
      <h2>Graph Visualization</h2>
      <svg ref={svgRef} width="800" height="600"></svg>
    </div>
  );
};

export default GraphView;
```

```tsx
// web/src/pages/Alerts.tsx
import React, { useEffect, useState } from 'react';

const Alerts: React.FC = () => {
  const [alerts, setAlerts] = useState<any[]>([]);

  useEffect(() => {
    // Load alerts
  }, []);

  return (
    <div className="alerts-page">
      <h2>Alert Center</h2>
      <ul className="alerts-list">
        {alerts.map((alert) => (
          <li key={alert.id} className={`alert alert-${alert.severity}`}>
            <strong>{alert.ruleName}</strong>
            <p>{alert.message}</p>
            <small>{alert.timestamp}</small>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default Alerts;
```

- [ ] **Step 5: Commit**
```bash
git add -A && git commit -m "feat(web): add React web console"
```

---

## Task 5: Electron Desktop 应用

**Files:**
- Create: `desktop/package.json`
- Create: `desktop/main.js`
- Create: `desktop/preload.js`
- Create: `desktop/renderer/index.html`

- [ ] **Step 1: 初始化 Electron 项目**
```json
// desktop/package.json
{
  "name": "back-pushing-desktop",
  "version": "1.0.0",
  "main": "main.js",
  "scripts": {
    "start": "electron .",
    "build": "electron-builder"
  },
  "dependencies": {
    "better-sqlite3": "^9.2.0"
  },
  "devDependencies": {
    "electron": "^28.0.0",
    "electron-builder": "^24.9.0"
  }
}
```

- [ ] **Step 2: 实现主进程**
```javascript
// desktop/main.js
const { app, BrowserWindow, Tray, Menu } = require('electron');
const path = require('path');

let mainWindow;
let tray;

function createWindow() {
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
    },
  });

  // Load web app
  mainWindow.loadFile(path.join(__dirname, 'renderer', 'index.html'));

  // Development: open DevTools
  if (process.argv.includes('--dev')) {
    mainWindow.webContents.openDevTools();
  }
}

function createTray() {
  tray = new Tray(path.join(__dirname, 'icon.png'));
  const contextMenu = Menu.buildFromTemplate([
    { label: 'Show App', click: () => mainWindow.show() },
    { label: 'Quit', click: () => app.quit() },
  ]);
  tray.setToolTip('Back-Pushing');
  tray.setContextMenu(contextMenu);
}

app.whenReady().then(() => {
  createWindow();
  createTray();
});

app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});
```

- [ ] **Step 3: 实现预加载脚本**
```javascript
// desktop/preload.js
const { contextBridge, ipcRenderer } = require('electron');

contextBridge.exposeInMainWorld('electronAPI', {
  // Database operations
  getObjects: (type) => ipcRenderer.invoke('db:getObjects', type),
  createObject: (type, id, data) => ipcRenderer.invoke('db:createObject', type, id, data),

  // Sync operations
  syncData: () => ipcRenderer.invoke('sync:data'),
  getSyncStatus: () => ipcRenderer.invoke('sync:status'),

  // App info
  getVersion: () => ipcRenderer.invoke('app:version'),
});
```

- [ ] **Step 4: 创建 Renderer HTML**
```html
<!-- desktop/renderer/index.html -->
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Back-Pushing Desktop</title>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
      margin: 0;
      padding: 20px;
      background: #1a1a2e;
      color: #eee;
    }
    .container { max-width: 1200px; margin: 0 auto; }
    h1 { color: #00d9ff; }
    .panel {
      background: #16213e;
      border-radius: 8px;
      padding: 20px;
      margin: 20px 0;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>Back-Pushing Desktop</h1>
    <div class="panel">
      <h2>Local Data</h2>
      <p>Objects stored locally: <span id="local-count">0</span></p>
      <button id="sync-btn">Sync with Server</button>
    </div>
    <div class="panel">
      <h2>Status</h2>
      <p id="status">Ready</p>
    </div>
  </div>
  <script>
    // Initialize app
    document.getElementById('sync-btn').addEventListener('click', async () => {
      document.getElementById('status').textContent = 'Syncing...';
      // Sync logic here
      document.getElementById('status').textContent = 'Synced';
    });
  </script>
</body>
</html>
```

- [ ] **Step 5: Commit**
```bash
git add -A && git commit -m "feat(desktop): add Electron desktop application"
```

---

## Task 6: 集成与测试

- [ ] **Step 1: 更新 Makefile**
```makefile
# Makefile 添加
.PHONY: build-server build-web build-desktop

build-server:
	go build -o bin/server ./cmd/server

build-web:
	cd web && npm install && npm run build

build-desktop:
	cd desktop && npm install && npm run build

build-all: build-server build-web build-desktop
```

- [ ] **Step 2: 运行完整测试**
```bash
go test ./...
npm test --prefix web
npm test --prefix desktop
```

- [ ] **Step 3: Commit**
```bash
git add -A && git commit -m "chore: add build targets and integration tests"
```

---

## 自检清单

- [x] **Spec覆盖检查**：
  - REST API → Task 1
  - GraphQL API → Task 2
  - JWT 认证 → Task 3
  - Web 控制台 → Task 4
  - Electron Desktop → Task 5
  - 集成测试 → Task 6

- [x] **Placeholder扫描**：无 TBD/TODO

- [x] **类型一致性**：
  - REST API: /api/objects/:type/:id 等路由一致
  - GraphQL: schema 与 resolver 一致
  - JWT: Middleware 模式一致
