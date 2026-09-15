# AbingBlog-Backend

个人博客后端 —— Go 语言练手项目，目标是串起 **本地开发 → 分层架构 → Docker → CI → CD → 公网访问** 的完整闭环。

技术栈横跨 Gin + GORM + MySQL + Redis + JWT，已覆盖文章/分类/标签/作品墙 CRUD、JWT 单管理员鉴权、Docker 多阶段构建、GitHub Actions CI/CD，配套一个 Vue 3 + TypeScript 前端（含后台管理）。

## 技术栈

| 层 | 选型 | 用途 |
|---|---|---|
| HTTP 框架 | Gin | 路由、中间件、请求解析 |
| ORM | GORM | 数据库操作、自动迁移、软删除 |
| 配置 | YAML + 环境变量覆盖 | `configs/config.yaml` 兜底，部署时靠环境变量注入 |
| 鉴权 | JWT（golang-jwt/v5）+ bcrypt | 单管理员登录，后台接口鉴权 |
| 数据库 | MySQL 8.0 | 持久化存储 |
| 缓存 | Redis（客户端已接入，缓存策略待实现） | 文章列表/详情缓存，Cache Aside |
| 消息队列 | 待定（计划中） | 浏览量异步落库 |
| 容器 | Docker + Docker Compose | 本地开发环境 + 生产部署 |
| 反向代理 | Nginx（`deploy/gateway/`） | 统一入口，`/` → 前端，`/api` → 后端 |
| CI/CD | GitHub Actions | CI：格式 + vet + test + build；CD：推 GHCR 镜像 + SSH 部署 |

## 项目结构

```
Blog-Backend/
├── cmd/
│   ├── server/
│   │   └── main.go              # 程序入口：装配依赖，启动服务（启动时自动 seed 管理员）
│   └── seed/
│       └── main.go              # 独立的管理员初始化命令（可重复执行，已存在则跳过）
├── configs/
│   ├── config.yaml              # 配置文件（端口、数据库、Redis、JWT、CORS）
│   └── config_test.go           # 配置解析测试
├── internal/                    # 业务代码（Go 惯例：internal 对外不可 import）
│   ├── config/
│   │   └── config.go            # 读取 config.yaml → Config 结构体，再用环境变量覆盖
│   ├── database/
│   │   └── database.go          # 连接 MySQL，自动建库建表（AutoMigrate）
│   ├── cache/
│   │   └── cache.go             # 连接 Redis（go-redis/v9）并启动探活
│   ├── auth/
│   │   ├── jwt.go               # GenerateToken / ParseToken（HS256，校验签名算法）
│   │   └── password.go          # bcrypt 哈希与校验
│   ├── model/
│   │   └── model.go             # 数据库表 → Go struct（User/Category/Tag/Article/Project/SiteMetric）
│   ├── repository/              # 数据访问层 —— 只写 SQL，不写业务逻辑
│   │   ├── article.go           # ArticleRepo：文章 CRUD + 分页查询 + 标签关联
│   │   ├── category.go          # CategoryRepo：分类 CRUD + 分组统计文章数
│   │   ├── tag.go               # TagRepo：标签 CRUD + 分组统计文章数
│   │   ├── user.go              # UserRepo：按用户名查管理员
│   │   ├── project.go           # ProjectRepo：作品墙 CRUD + 分页
│   │   └── site_metric.go       # SiteMetricRepo：启动次数原子自增
│   ├── service/                 # 业务逻辑层 —— 校验、状态规则、数据组装
│   │   ├── article.go           # ArticleService：创建/更新/发布/撤回/删除
│   │   ├── category.go          # CategoryService：唯一性校验、删除保护
│   │   ├── tag.go               # TagService：唯一性校验、删除时清理关联
│   │   ├── user.go              # UserService：SeedAdmin + Login 校验密码签发 token
│   │   ├── project.go           # ProjectService：slug 唯一性、状态规则
│   │   └── site_metric.go       # SiteMetricService：累计启动次数
│   ├── handler/                 # HTTP 接口层 —— 解析请求 → 调 service → 封装响应
│   │   ├── handler.go           # Handler 聚合结构体（所有 handler 的集合）
│   │   ├── article.go           # ArticleHandler：文章列表/详情/创建/更新/删除/发布
│   │   ├── category.go          # CategoryHandler：分类 CRUD
│   │   ├── tag.go               # TagHandler：标签 CRUD
│   │   ├── project.go           # ProjectHandler：项目 CRUD
│   │   ├── user.go              # UserHandler：登录
│   │   ├── site_metric.go       # SiteMetricHandler：/visits 读写
│   │   ├── health.go            # HealthHandler：健康检查（DB 连通性）
│   │   └── dto.go               # 返回给前端的 JSON 结构（与数据库 model 解耦）
│   ├── middleware/
│   │   ├── auth.go              # AuthRequired：解析 Bearer token，失败 401
│   │   └── cors.go              # CORS 白名单，放行前端 dev server
│   ├── router/
│   │   └── router.go            # 路由注册 + 权限分组（公共 / 后台）
│   ├── response/
│   │   └── response.go          # 统一响应格式 {code, message, data}
│   └── errcode/
│       └── errcode.go           # 业务错误码定义 + BizError 类型
├── deploy/
│   └── gateway/                 # 统一入口 Nginx：/ → 前端，/api → 后端
│       ├── docker-compose.yml
│       └── nginx.conf
├── docs/
│   ├── api-v1.md                # API 文档（前后端契约）
│   ├── deployment.md            # 部署与调试手册（ECS + Nginx + HTTPS + 回滚）
│   └── TODO.md                  # 待办与阶段规划
├── docker-compose.yml           # 本地 MySQL + Redis 开发环境
├── Dockerfile                   # 生产镜像，多阶段构建
├── docker-compose.prod.yml      # 生产服务编排（mysql + redis + server）
├── .github/workflows/
│   ├── ci.yml                   # 格式 + vet + test(-race) + build
│   └── deploy.yml               # 发布 GHCR 镜像并通过 SSH 部署
├── go.mod
└── go.sum
```

## 快速开始

完整的阿里云 ECS、Docker、Nginx、HTTPS、CI/CD、回滚和故障排查说明见 [`docs/deployment.md`](docs/deployment.md)。

### 1. 前置条件

- Go 1.27+
- Docker Desktop（提供本地 MySQL + Redis）

### 2. 启动 MySQL 与 Redis

```bash
# 启动依赖容器（后台运行）
docker compose up -d

# 确认已就绪
docker compose ps
# 看到 blog-mysql / blog-redis 状态为 healthy 即可

# 停止（数据不丢失，存在 Docker 命名卷里）
docker compose down
```

### 3. 启动后端服务

```bash
# 安装依赖
go mod tidy

# 运行（首次启动会自动创建初始管理员 admin/admin123456）
go run ./cmd/server/

# 看到以下输出表示启动成功：
# Redis connected at 127.0.0.1:6379 (db=0)
# server running at http://localhost:8080
```

需要手动重置/新建管理员时用独立命令（已存在则跳过，不会覆盖密码）：

```bash
ABINGBLOG_ADMIN_USERNAME=admin ABINGBLOG_ADMIN_PASSWORD=你的密码 go run ./cmd/seed/
```

### 4. 验证

```bash
# 健康检查
curl http://localhost:8080/api/v1/healthz

# 文章列表（此时为空）
curl http://localhost:8080/api/v1/articles

# 登录拿 token（返回值里的 token 给下面所有 admin 请求用）
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123456"}'

TOKEN=<上一步返回的 token>

# 创建分类
curl -X POST http://localhost:8080/api/v1/admin/categories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Go","slug":"go"}'

# 创建标签
curl -X POST http://localhost:8080/api/v1/admin/tags \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"后端"}'

# 创建文章
curl -X POST http://localhost:8080/api/v1/admin/articles \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"title":"第一篇文章","content":"Hello World","status":"published","category_id":1,"tag_ids":[1]}'

# 查看文章列表 / 详情
curl http://localhost:8080/api/v1/articles
curl http://localhost:8080/api/v1/articles/1

# 项目列表（作品墙）
curl http://localhost:8080/api/v1/projects

# 累计启动次数（只读 / 自增一次）
curl http://localhost:8080/api/v1/visits
curl -X POST http://localhost:8080/api/v1/visits/start
```

> 未带 token 访问 `/api/v1/admin/*` 会返回 401，这是 JWT 中间件在起作用——可以顺手验证一下鉴权链路。

## 架构设计

### 分层架构（三层）

```
HTTP 请求
  │
  ▼
┌──────────┐
│  router   │  路由层：URL → handler 映射，分组控制权限边界
└────┬─────┘
     │
     ▼
┌──────────┐
│  handler  │  接口层：解析请求参数、调 service、封装 JSON 响应
└────┬─────┘  不写业务逻辑，不直接操作数据库
     │
     ▼
┌──────────┐
│  service  │  业务逻辑层：参数校验、状态规则、数据组装、抛出业务错误
└────┬─────┘  不碰 HTTP 的东西（gin.Context），不拼 SQL
     │
     ▼
┌──────────┐
│ repository│  数据访问层：只写 GORM/SQL 查询，不写业务逻辑
└────┬─────┘
     │
     ▼
┌──────────┐
│  model    │  数据模型：Go struct = 数据库表结构
└────┬─────┘
     │
     ▼
  MySQL
```

### 为什么要分层？

每一层只做一件事，改一个地方不影响其他地方：

- **改数据库表结构** → 只改 `model/` 和 `repository/`，handler 和 service 不用动
- **改业务规则**（比如"草稿不能超过 10 篇"）→ 只改 `service/`，handler 和 repo 不用动
- **改接口返回格式** → 只改 `handler/dto.go`，service 和 repo 不用动
- **换数据库**（MySQL → PostgreSQL）→ 只改 `repository/` 和 `database/`，上层完全无感

### 依赖注入（手写，不用框架）

```go
// cmd/server/main.go — 依赖装配顺序：底层 → 上层
cfg := config.Load()
cache.Init(cfg)          // Redis 探活，连不上直接 Fatal，避免带病启动
db := database.Init(cfg)

// 1. repo 依赖 db
articleRepo  := repository.NewArticleRepo(db)
categoryRepo := repository.NewCategoryRepo(db)
tagRepo      := repository.NewTagRepo(db)
userRepo     := repository.NewUserRepo(db)
metricRepo   := repository.NewSiteMetricRepo(db)
projectRepo  := repository.NewProjectRepo(db)

// 2. service 依赖 repo（UserService 额外依赖 cfg，用来取 JWT 密钥和初始管理员）
userSvc := service.NewUserService(userRepo, cfg)
userSvc.SeedAdmin()      // 启动时确保初始管理员存在，已存在则跳过

// 3. handler 依赖 service
h := &handler.Handler{
    Article:  handler.NewArticleHandler(service.NewArticleService(articleRepo, categoryRepo, tagRepo)),
    Category: handler.NewCategoryHandler(service.NewCategoryService(categoryRepo)),
    Tag:      handler.NewTagHandler(service.NewTagService(tagRepo)),
    User:     handler.NewUserHandler(userSvc),
    Health:   handler.NewHealthHandler(db),
    Metric:   handler.NewSiteMetricHandler(service.NewSiteMetricService(metricRepo)),
    Project:  handler.NewProjectHandler(service.NewProjectService(projectRepo)),
}

// 4. router 依赖 handler + JWT 密钥 + CORS 白名单
r := router.Setup(h, cfg.JWT.Secret, cfg.CORS.AllowOrigins)
```

这种手写方式的好处：**显式、可读、不需要学 wire 之类的工具**。依赖关系一目了然。

### 配置加载：YAML 兜底 + 环境变量覆盖

`config.Load()` 分两步（`internal/config/config.go`）：

1. 读 `configs/config.yaml` 得到本地开发默认值
2. 逐个字段检查对应的大写环境变量是否设置，设置了就覆盖（空字符串视为未设置，避免容器里漏传的变量把默认值冲成空）

这样同一份二进制在本地和容器里跑同一套代码路径：本地靠 YAML，生产靠 `MYSQL_HOST`、`REDIS_HOST`、`JWT_SECRET` 等环境变量（见 `docker-compose.prod.yml`）。

> ⚠️ `configs/config.yaml` 里的 `jwt.secret` 只是占位值，`JWTConfig.Validate()` 要求密钥至少 32 字节。生产**必须**用 `JWT_SECRET` 环境变量注入强密钥。

### 统一响应格式

所有接口返回同一个 JSON 结构：

```json
{
  "code": 0,
  "message": "success",
  "data": { ... }
}
```

- `code=0` 成功，`code≠0` 失败（具体含义见 `docs/api-v1.md` 的错误码表）
- handler 层统一用 `response.OK()` / `response.Fail()` / `response.Error()` 三个函数
- service 层抛出 `errcode.BizError`，handler 层自动映射为 HTTP 状态码

### Model vs DTO

数据库的 Go struct（`model.Article`）和返回给前端的 JSON 结构（`articleDTO`）是两套：

- 数据库加了字段，不自动暴露给前端
- 列表接口不返回 `content`（Markdown 正文很大），详情接口才返回
- `handler/dto.go` 里的 `toArticleDTO()` 负责转换，控制哪些字段暴露

### 路由权限模型

```
/api/v1/*          → 公共接口（游客可访问，只返回 status=published 的内容）
/api/v1/admin/*    → 后台接口（必须携带 Authorization: Bearer <token>）
```

具体路由（`internal/router/router.go`）：

| 分组 | 路由 |
|---|---|
| 公共 | `GET /articles`、`GET /articles/:id`、`GET /categories`、`GET /tags`、`GET /projects`、`GET /healthz`、`GET /visits`、`POST /visits/start`、`POST /login` |
| 后台（JWT） | `articles` / `projects` 的增删改查 + `PUT /articles/:id/status`，`categories` / `tags` 的增改删 |

鉴权链路：`POST /login` → `UserService.Login` 查用户 → `bcrypt` 比对密码 → `auth.GenerateToken` 签发 JWT → 前端存 token → 请求 admin 接口带 `Authorization` 头 → `middleware.AuthRequired` 调 `auth.ParseToken` 验签 → 通过则 `c.Set("user_id", ...)` 并放行，失败 401。

> `ParseToken` 的 keyFunc 里**必须**校验签名算法是 HMAC，否则会被 `alg=none` 攻击绕过验签。

## 数据库表关系

```
users (管理员，单用户)          site_metrics (站点指标，单行 id=1)

categories (分类)
  │
  ├── articles (文章) ──多对多── tags (标签)
  │                        │
  │                   article_tags (中间表，GORM 自动生成)

projects (作品墙，独立表)
```

| 表 | 说明 |
|---|---|
| `users` | 管理员账号，密码 bcrypt 入库，绝不存明文 |
| `categories` | 文章分类，`slug` 供前端筛选，`sort` 控制展示顺序 |
| `tags` | 文章标签，删除时清理 `article_tags` 关联 |
| `articles` | 文章，正文为 Markdown 原文（前端负责渲染） |
| `projects` | 作品墙条目，字段对齐前端 `ProjectSummary` |
| `site_metrics` | 站点指标，固定 `id=1`，用 `UPDATE ... + 1` 原子自增 |

- 一篇文章属于一个分类（`category_id`，0 表示未分类），可以有多个标签（多对多，通过 `article_tags` 中间表）
- 文章和项目都支持软删除（`deleted_at`）：DELETE 只打标记，数据可恢复
- 分类下有文章时不允许删除分类（防止文章变成"无分类"状态）
- 文章和项目共用 `status` 约定：`draft` 草稿 / `published` 已发布，公共接口只放行 `published`

## 如何添加新功能

以添加「评论」功能为例，标准流程：

```
1. internal/model/model.go       → 加 Comment struct 定义表结构
2. internal/repository/comment.go → NewCommentRepo，写 CRUD 方法
3. internal/service/comment.go    → NewCommentService，写业务校验
4. internal/handler/comment.go    → NewCommentHandler，写 HTTP 接口
5. internal/handler/handler.go    → 在 Handler 结构体里加 Comment 字段
6. internal/router/router.go      → 注册路由
7. cmd/server/main.go             → 装配依赖链：repo → service → handler
8. docs/api-v1.md                 → 更新 API 文档
```

每一步都是**复制已有的模式**——看看 `category.go` 的三层怎么写，照着写评论就行。

## 开发命令速查

```bash
# 运行
go run ./cmd/server/

# 构建
go build -o server.exe ./cmd/server/

# 静态检查
go vet ./...

# 测试
go test ./...

# 格式化检查（CI 会跑，格式不对直接失败）
gofmt -l .

# 安装/更新依赖
go mod tidy

# 查看依赖关系
go mod graph
```

## 路线图

| 阶段 | 内容 | 状态 |
|---|---|---|
| 1 | 项目骨架 + 分层架构 + 文章/分类/标签 CRUD | ✅ 已完成 |
| 2 | JWT 登录认证 + 后台接口鉴权 + 环境变量覆盖配置 | ✅ 已完成 |
| 3 | 作品墙（projects）+ 站点指标（累计启动次数） | ✅ 已完成 |
| 4 | Docker 化（多阶段构建）+ 生产 compose | ✅ 已完成 |
| 5 | CI（格式/vet/test/build）+ CD（GHCR + SSH 部署） | 🟡 流水线已就绪，待配置 Secrets 并首次跑通 |
| 6 | 前后端联调 + 网关编排（Nginx） | 🟡 网关配置已就绪，待部署验证 |
| 7 | Redis 缓存（Cache Aside） | 🟡 客户端已接入探活，缓存策略待实现 |
| 8 | MQ 异步（浏览量落库） | 🔲 待开发 |

已完成的能力链路：**本地开发 → 分层架构 → 容器化 → CI → CD**。剩余目标是**公网访问闭环**。

接口契约见 [`docs/api-v1.md`](docs/api-v1.md)，部署与排查见 [`docs/deployment.md`](docs/deployment.md)，待办拆解见 [`docs/TODO.md`](docs/TODO.md)。

## 常见问题

**Q: 为什么 `internal/` 下分这么多包，不能全写在一个文件里吗？**

A: 可以，但分层后每一层职责单一，改 bug 时能快速定位。面试官看到分层清晰的项目会比看到一个大文件印象好得多。这也是 Go 社区的主流做法。

**Q: 为什么不用 wire 之类的依赖注入框架？**

A: 这个项目依赖少，手动装配不到 10 行代码，引入框架反而增加学习成本。等依赖多到手动装配变痛苦时再考虑也不迟。

**Q: 如何部署生产环境？**

A: 准备一台装了 Docker 的 Linux 服务器，在 GitHub 配置 `DEPLOY_HOST`、`DEPLOY_USER`、`DEPLOY_SSH_KEY`、`DEPLOY_PATH`、`MYSQL_ROOT_PASSWORD` 和 `JWT_SECRET`（长度 ≥32 字节）。推送版本标签（如 `v1.0.0`）或手动运行 `Deploy` workflow，流水线会构建镜像推到 GHCR、SSH 到服务器拉起 `docker-compose.prod.yml` 并做健康检查。前端容器和网关（`deploy/gateway/`）另需部署。完整步骤见 [`docs/deployment.md`](docs/deployment.md)。

**Q: 为什么 Redis 接进来了却没看到缓存代码？**

A: `internal/cache/cache.go` 目前只做「连上 Redis + 启动探活」这一件事。具体的缓存策略——读时 Cache Aside、写时如何让缓存失效、key 怎么设计、TTL 多久、穿透/雪崩怎么防——属于 service 层的业务决策，还没写。这是面试高频点，故意留着自己实现而不是让 AI 代劳。

**Q: 数据库表结构改了怎么办？**

A: 开发期用 `AutoMigrate` 自动同步表结构（`database/database.go`）。上线后应换成正式的迁移工具（如 `golang-migrate`）。
