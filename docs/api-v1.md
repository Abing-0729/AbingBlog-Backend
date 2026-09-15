# Blog API v1 设计

> 状态：草稿，待评审。评审通过后作为前后端开发的共同契约。

## 1. 通用约定

- Base URL：`/api/v1`
- 数据格式：JSON（`application/json`）
- 认证方式：JWT，请求头 `Authorization: Bearer <token>`
- 通用响应包装：

```json
{ "code": 0, "message": "success", "data": { } }
```

| HTTP 状态码 | 含义 |
|---|---|
| 200 | 成功（code=0）或业务失败（code≠0） |
| 400 | 请求参数错误 |
| 401 | 未认证 / token 失效 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

- 分页响应格式：

```json
{ "list": [], "total": 100, "page": 1, "page_size": 10 }
```

- 时间格式：ISO 8601（`2026-09-08T10:30:00+08:00`）

### 权限模型（只有一个管理员）

| 角色 | 能看到/能做的 |
|---|---|
| 游客（无需登录） | 浏览所有**已发布**文章、文章详情、分类、标签；看不到草稿，不能做任何写操作 |
| 管理员（登录后） | 游客的一切 + 发布/修改/删除文章、管理分类标签；只有站长持有账号，账号不对外公开 |

> 实现上：公共接口只查 `status=published`；`/admin/*` 路由组统一挂 JWT 中间件，无有效 token 一律 401。

## 2. 数据模型（初版）

| 表 | 关键字段 |
|---|---|
| users | id, username, password(bcrypt), created_at |
| site_metrics | id, start_count |
| categories | id, name, slug, sort, created_at |
| tags | id, name, created_at |
| articles | id, title, content(Markdown), summary, cover, category_id, status(draft/published), view_count, published_at, created_at, updated_at, deleted_at(软删) |
| article_tags | article_id, tag_id（多对多） |

## 3. 公共接口（无需登录）

### GET /articles —— 文章列表（首页列表）
- Query：`page`(默认1) `page_size`(默认10, max 50) `category_slug?` `tag?` `keyword?`（标题模糊搜索）
- 只返回 `status=published`
- 不返回 content 全文
- 响应 `data`：分页格式，list 元素：
```json
{ "id": 1, "title": "标题", "summary": "摘要", "cover": "https://...",
  "category": { "id": 1, "name": "Go", "slug": "go" },
  "tags": [ { "id": 1, "name": "后端" } ],
  "view_count": 123, "published_at": "...", "created_at": "..." }
```

### GET /articles/:id —— 文章详情
- 响应：列表字段 + `content`（Markdown 原文）
- 副作用：浏览量 +1（**v1 后期改为 MQ 异步，详见 §5**）
- 文章不存在或未发布 → 404（后台自己看草稿走 admin 接口）

### GET /categories —— 分类列表
- 响应：`[ { "id": 1, "name": "Go", "slug": "go", "article_count": 10 } ]`

### GET /tags —— 标签列表
- 响应：`[ { "id": 1, "name": "后端", "article_count": 3 } ]`

### GET /projects —— 项目列表（作品墙）
- 只返回已发布项目，按 `sort` 升序、`id` 降序排列，分页参数同文章列表（`page` / `page_size`）。
- 响应：`data: { "list": [ { "id": 1, "slug": "abingblog", "name": "ABINGBLOG", "detail": "...", "stack": "GO · VUE · MYSQL", "github_url": "", "demo_url": "", "sort": 0, "status": "published", "created_at": "...", "updated_at": "..." } ], "total": 3, "page": 1, "page_size": 10 }`
- 前端 `ProjectSummary` 的 `githubUrl` / `demoUrl` 对应 `github_url` / `demo_url`，由前端映射层转换。

### GET /healthz —— 健康检查
- 供 Docker / CD 部署做健康检查，检查 DB 连通性
- 响应：`{ "status": "ok", "db": "ok" }`

> 说明：不单独做「首页聚合接口」。首页 = 文章列表 + 分类 + 标签三个接口，前端并发请求自行组合，保持后端简单。

### GET /visits —— 读取累计启动次数（只读）
- 无需登录；入口屏幕加载时调用，纯展示。
- **绝不自增**：刷新/重播随便调，数字不变。计数行还没建过时返回 0。
- 响应：`data: { "start_count": 42 }`

### POST /visits/start —— 记录一次启动
- 无需登录；访客点击街机入口的 `PRESS START` 时调用一次。
- 原子自增（`UPDATE ... start_count = start_count + 1`），并发下不丢计数。
- 响应：`data: { "start_count": 43 }`（返回自增后的新值）
- `start_count` 持久化在站点指标表中，用于入口屏幕显示累计启动次数。

## 4. 认证与后台接口（JWT）

### POST /auth/login —— 登录
- 请求：`{ "username": "admin", "password": "xxx" }`
- 成功：`data: { "token": "<jwt>", "expires_in": 86400 }`
- 失败：code=1001（用户名或密码错误）
- 无注册接口——博客只有站长一个用户，初始化时用脚本/命令建管理员

### 后台文章（`/admin/*`，均需 JWT）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /admin/articles | 列表含草稿，支持 `status` 筛选，分页同公共列表 |
| GET | /admin/articles/:id | 单条详情（后台编辑用）：不限状态含草稿，返回带 Markdown 正文 |
| POST | /admin/articles | 创建：`{ title, content, summary?, cover?, category_id, tag_ids[], status }` |
| PUT | /admin/articles/:id | 更新，字段同创建（全量更新） |
| DELETE | /admin/articles/:id | 软删除 |
| PUT | /admin/articles/:id/status | 发布/撤回：`{ "status": "published" }` |

### 后台分类 / 标签

| 方法 | 路径 | 说明 |
|---|---|---|
| POST | /admin/categories | `{ name, slug?, sort? }` |
| PUT | /admin/categories/:id | 更新 |
| DELETE | /admin/categories/:id | 删除；分类下有文章则拒绝（code=1003） |
| POST | /admin/tags | `{ name }` |
| PUT | /admin/tags/:id | 更新 |
| DELETE | /admin/tags/:id | 删除，同时清关联 |

### 后台项目（`/admin/*`，均需 JWT）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | /admin/projects | 列表含草稿，支持 `status` 筛选，分页同公共列表 |
| POST | /admin/projects | 创建：`{ slug, name, detail?, stack?, github_url?, demo_url?, sort?, status }` |
| PUT | /admin/projects/:id | 更新，字段同创建（全量更新） |
| DELETE | /admin/projects/:id | 软删除 |

### 业务错误码（初版）

| code | 含义 |
|---|---|
| 0 | 成功 |
| 1000 | 请求参数错误 |
| 1001 | 用户名或密码错误 |
| 1002 | 文章不存在 |
| 1003 | 分类下存在文章，不可删除 |
| 1004 | 分类不存在 |
| 1005 | 标签不存在 |
| 1006 | 文章标题不能为空 |
| 1007 | 文章状态非法（只能 draft/published） |
| 1008 | 名称不能为空 |
| 1009 | 名称已存在 |
| 1010 | 项目不存在 |
| 1011 | 项目 slug 不能为空 |
| 1012 | 项目 slug 已存在 |
| 5000 | 服务器内部错误 |

## 5. Redis 缓存与 MQ 异步业务（初版方案，实现阶段细化）

- **Redis（缓存）**：Cache Aside 模式
  - 缓存对象：文章列表（按 分页+筛选参数 为 key）、文章详情、分类/标签列表
  - 失效策略：后台增删改时主动删除对应缓存
- **MQ（真实异步业务，候选）**：
  1. **浏览量统计（推荐）**：详情接口 Redis `INCR` 立即返回 + 发 MQ 消息，消费者攒批异步落库 `view_count`。真实解耦：读接口不直接写库
  2. 备选：文章发布时异步生成摘要/渲染 HTML
  - MQ 选型（RabbitMQ / Kafka）待定，第二阶段决定

## 6. 后续（不阻塞 v1）

- POST /admin/uploads —— 图片上传（Markdown 插图）
- 修改密码接口
- 评论、全文搜索、按时间归档、RSS
- HTTPS、域名、日志、监控（部署阶段）
