# AbingBlog-Backend

个人博客后端 —— Go 语言练手项目，目标是串起 **本地开发 → 分层架构 → Docker → CI → CD → 公网访问** 的完整闭环。

## 技术栈

| 层 | 选型 | 用途 |
|---|---|---|
| HTTP 框架 | Gin | 路由、中间件、请求解析 |
| ORM | GORM | 数据库操作、自动迁移、软删除 |
| 配置 | Viper | 读取 YAML 配置，支持环境变量覆盖 |
| 数据库 | MySQL 8.0 | 持久化存储 |
| 缓存 | Redis（计划中） | 文章列表/详情缓存，Cache Aside |
| 消息队列 | 待定（计划中） | 浏览量异步落库 |
| 容器 | Docker + Docker Compose | 本地开发环境 + 生产部署 |
| CI/CD | GitHub Actions | 自动检查、构建、测试、部署 |

## 项目结构

```
Blog-Backend/
├── cmd/
│   └── server/
│       └── main.go              # 程序入口：装配依赖，启动服务
├── configs/
│   └── config.yaml              # 配置文件（端口、数据库连接信息）
├── internal/                    # 业务代码（Go 惯例：internal 对外不可 import）
│   ├── config/
│   │   └── config.go            # 读取 config.yaml → Config 结构体
│   ├── database/
│   │   └── database.go          # 连接 MySQL，自动建库建表（AutoMigrate）
│   ├── model/
│   │   └── model.go             # 数据库表 → Go struct（User/Category/Tag/Article）
│   ├── repository/              # 数据访问层 —— 只写 SQL，不写业务逻辑
│   │   ├── article.go           # ArticleRepo：文章 CRUD + 分页查询 + 标签关联
│   │   ├── category.go          # CategoryRepo：分类 CRUD + 分组统计文章数
│   │   └── tag.go               # TagRepo：标签 CRUD + 分组统计文章数
│   ├── service/                 # 业务逻辑层 —— 校验、状态规则、数据组装
│   │   ├── article.go           # ArticleService：创建/更新/发布/撤回/删除
│   │   ├── category.go          # CategoryService：唯一性校验、删除保护
│   │   └── tag.go               # TagService：唯一性校验、删除时清理关联
│   ├── handler/                 # HTTP 接口层 —— 解析请求 → 调 service → 封装响应
│   │   ├── handler.go           # Handler 聚合结构体（所有 handler 的集合）
│   │   ├── article.go           # ArticleHandler：文章列表/详情/创建/更新/删除/发布
│   │   ├── category.go          # CategoryHandler：分类 CRUD
│   │   ├── tag.go               # TagHandler：标签 CRUD
│   │   ├── health.go            # HealthHandler：健康检查（DB 连通性）
│   │   └── dto.go               # 返回给前端的 JSON 结构（与数据库 model 解耦）
│   ├── router/
│   │   └── router.go            # 路由注册 + 权限分组（公共 / 后台）
│   ├── response/
│   │   └── response.go          # 统一响应格式 {code, message, data}
│   └── errcode/
│       └── errcode.go           # 业务错误码定义 + BizError 类型
├── docs/
│   └── api-v1.md                # API 文档（前后端契约）
├── docker-compose.yml           # 本地 MySQL 开发环境
├── .github/workflows/
│   └── ci.yml                   # GitHub Actions：vet + build + test
├── go.mod
└── go.sum
```

## 快速开始

### 1. 前置条件

- Go 1.27+
- Docker Desktop（提供本地 MySQL）

### 2. 启动 MySQL

```bash
# 启动 MySQL 容器（后台运行）
docker compose up -d

# 确认 MySQL 已就绪
docker compose ps
# 看到 blog-mysql 状态为 healthy 即可

# 停止（数据不丢失，存在 Docker 命名卷里）
docker compose down
```

### 3. 启动后端服务

```bash
# 安装依赖
go mod tidy

# 运行
go run ./cmd/server/

# 看到以下输出表示启动成功：
# server running at http://localhost:8080
```

### 4. 验证

```bash
# 健康检查
curl http://localhost:8080/api/v1/healthz

# 文章列表（此时为空）
curl http://localhost:8080/api/v1/articles

# 创建分类
curl -X POST http://localhost:8080/api/v1/admin/categories \
  -H "Content-Type: application/json" \
  -d '{"name":"Go","slug":"go"}'

# 创建标签
curl -X POST http://localhost:8080/api/v1/admin/tags \
  -H "Content-Type: application/json" \
  -d '{"name":"后端"}'

# 创建文章
curl -X POST http://localhost:8080/api/v1/admin/articles \
  -H "Content-Type: application/json" \
  -d '{"title":"第一篇文章","content":"Hello World","status":"published","category_id":1,"tag_ids":[1]}'

# 查看文章列表
curl http://localhost:8080/api/v1/articles
```

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
db := database.Init(cfg)

articleRepo := repository.NewArticleRepo(db)     // 1. repo 依赖 db
categoryRepo := repository.NewCategoryRepo(db)
tagRepo := repository.NewTagRepo(db)

articleSvc := service.NewArticleService(articleRepo, categoryRepo, tagRepo) // 2. service 依赖 repo
categorySvc := service.NewCategoryService(categoryRepo)
tagSvc := service.NewTagService(tagRepo)

h := &handler.Handler{                            // 3. handler 依赖 service
    Article:  handler.NewArticleHandler(articleSvc),
    Category: handler.NewCategoryHandler(categorySvc),
    Tag:      handler.NewTagHandler(tagSvc),
    Health:   handler.NewHealthHandler(db),
}

r := router.Setup(h)                              // 4. router 依赖 handler
```

这种手写方式的好处：**显式、可读、不需要学 wire 之类的工具**。依赖关系一目了然。

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
/api/v1/admin/*    → 后台接口（下一迭代挂 JWT 中间件，当前暂无鉴权）
```

## 数据库表关系

```
users (管理员)
  │
categories (分类)
  │
  ├── articles (文章) ──多对多── tags (标签)
  │                        │
  │                   article_tags (中间表，GORM 自动生成)
```

- 一篇文章属于一个分类（`category_id`）
- 一篇文章可以有多个标签（多对多，通过 `article_tags` 中间表）
- 文章支持软删除（`deleted_at`）：DELETE 只打标记，数据可恢复
- 分类下有文章时不允许删除分类（防止文章变成"无分类"状态）

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

# 安装/更新依赖
go mod tidy

# 查看依赖关系
go mod graph
```

## 路线图

| 阶段 | 内容 | 状态 |
|---|---|---|
| 1 | 项目骨架 + 分层架构 + 文章/分类/标签 CRUD | ✅ 已完成 |
| 2 | JWT 登录认证 + 后台接口鉴权 | 🔲 待开发 |
| 3 | Redis 缓存（Cache Aside） | 🔲 待开发 |
| 4 | MQ 异步（浏览量落库） | 🔲 待开发 |
| 5 | Docker 化 + 生产环境配置 | 🔲 待开发 |
| 6 | CI/CD 自动部署到服务器 | 🔲 待开发 |
| 7 | 前后端联调 + 公网访问 | 🔲 待开发 |

详见 `docs/api-v1.md`。

## 常见问题

**Q: 为什么 `internal/` 下分这么多包，不能全写在一个文件里吗？**

A: 可以，但分层后每一层职责单一，改 bug 时能快速定位。面试官看到分层清晰的项目会比看到一个大文件印象好得多。这也是 Go 社区的主流做法。

**Q: 为什么不用 wire 之类的依赖注入框架？**

A: 这个项目依赖少，手动装配不到 10 行代码，引入框架反而增加学习成本。等依赖多到手动装配变痛苦时再考虑也不迟。

**Q: 后台接口目前没有鉴权，谁都能调？**

A: 是的，当前 `/admin/*` 接口没有挂 JWT 中间件，这是下一迭代要做的事。代码里已标了 `TODO(JWT)` 注释。

**Q: 数据库表结构改了怎么办？**

A: 开发期用 `AutoMigrate` 自动同步表结构（`database/database.go`）。上线后应换成正式的迁移工具（如 `golang-migrate`）。