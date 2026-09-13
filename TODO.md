# 项目 TODO

> 记录当前进行中的任务和后续规划。带 🔧 的是"核心链路，自己写"，带 📦 的是"工程化样板，可让 AI 代劳"。
> 目标：后端上线完整闭环（本地开发 → Docker → CI → CD → Linux 服务器 → 公网）。

---

## 阶段一：认证收尾（进行中）

认证整条链已搭好脚手架，样板部分（config / User repo / bcrypt / seed / handler 绑定）已实现完整。
**剩下 3 处 `TODO(you)` 是核心链路，留给你自己写。** 全部填完认证就通了。

数据流回顾：

```
POST /login → 查 User → bcrypt 比对 → 签发 JWT → 前端存 token
           → 请求 admin 接口带 Authorization: Bearer <token>
           → 中间件 ParseToken 验签 → 通过放行 / 失败 401
```

### 🔧 TODO 1 — `internal/auth/jwt.go`：实现 JWT 签发与验签

这是整个认证的核心。JWT = 一段防篡改的字符串，服务端不存 session，token 自带 `user_id` 和过期时间，用密钥签名保证没被改过。

**要实现两个函数**（文件里已有函数签名和详细注释，把 `panic` 换成实现）：

`GenerateToken(userID, secret, expireHours)` — 登录成功时调，签发 token：
- 定义 Claims 结构（放 `user_id` + 标准过期字段）：
  ```go
  type Claims struct {
      UserID uint `json:"user_id"`
      jwt.RegisteredClaims
  }
  ```
- `jwt.NewWithClaims(jwt.SigningMethodHS256, claims)` 构造
- 过期时间：`ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour))`
- `token.SignedString([]byte(secret))` 得到字符串

`ParseToken(tokenStr, secret)` — 中间件里调，验签并取出 userID：
- `jwt.ParseWithClaims(tokenStr, &Claims{}, keyFunc)`
- **⚠️ 安全关键**：keyFunc 里必须校验签名算法，否则会被 `alg=none` 攻击绕过验签：
  ```go
  func(token *jwt.Token) (interface{}, error) {
      if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
          return nil, fmt.Errorf("非法签名算法")
      }
      return []byte(secret), nil
  }
  ```
- 检查 `token.Valid`，从 `token.Claims.(*Claims)` 取出 `UserID` 返回

Go 小抄：`import "github.com/golang-jwt/jwt/v5"`（go.mod 已有，直接用）。

---

### 🔧 TODO 2 — `internal/service/user.go` 的 `Login()`：登录业务

注释里写了 4 步，照着填：
1. `s.repo.GetByUsername(req.Username)` — 返回 `(nil, nil)` 表示用户不存在
2. 用户不存在 **或** `auth.CheckPassword(user.Password, req.Password)` 为 false → 都返回**同一个**错误：
   `return "", errcode.New(errcode.CodeLoginFail, "用户名或密码错误")`
   （不要分别提示"用户不存在""密码错"，否则暴露了哪个用户名存在）
3. `auth.GenerateToken(user.ID, s.cfg.JWT.Secret, s.cfg.JWT.ExpireHours)` 签发
4. 返回 token

填完把注释里那行占位的 `_ = errcode.CodeLoginFail` 删掉。

---

### 🔧 TODO 3 — 路由挂载 + 重写 JWT 中间件

**a) `internal/router/router.go`**：
- 公开分组加登录路由：`v1.POST("/login", h.User.Login)`
- admin 分组挂中间件（`Setup` 需要多接收一个 secret 参数，或直接从传入的 cfg 拿）：
  ```go
  admin := v1.Group("/admin")
  admin.Use(middleware.AuthRequired(secret))
  { ... }
  ```

**b) `internal/middleware/auth.go`**：你现在写的这版逻辑是断的（token 为空时 Abort 后直接 return，且没校验、没 `c.Next()`），需要重写：
```go
func AuthRequired(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. 取 Authorization 头，格式必须是 "Bearer <token>"
        authHeader := c.GetHeader("Authorization")
        // 2. 空 或 不以 "Bearer " 开头 → 401 + return（Abort 后必须 return，否则继续往下走）
        // 3. 切出 token 字符串（strings.TrimPrefix 去掉 "Bearer "）
        // 4. auth.ParseToken(token, secret)，出错 → 401 + return
        // 5. 成功：c.Set("user_id", userID) 存进上下文供后续 handler 用
        // 6. c.Next() 放行
    }
}
```
关键点：**每个 `AbortWithStatusJSON` 后面都要跟 `return`**，否则请求会继续走到真正的 handler。这正是你现在那版的 bug。

建议 401 响应体复用统一格式：`response.Fail(c, http.StatusUnauthorized, errcode.CodeLoginFail, "...")`（但要先 `c.Abort()`，因为 Fail 只写响应不中断）。

---

## 阶段二：认证之后（按顺序做）

### 📦 2.1 CORS 跨域（联调前端前必做）
Vue dev server（如 `localhost:5173`）和后端（`localhost:8080`）不同源，浏览器会拦。
需要一个 CORS 中间件放行前端来源、允许 `Authorization` 头。样板，可让我代劳。

### 🔧 2.2 从 token 里拿当前用户（可选加固）
中间件已把 `user_id` 存进上下文（`c.Set`）。如果后续要记录"谁改了文章"之类，
在 handler 里 `c.GetUint("user_id")` 取。现在单用户博客可以先不做。

### 📦📖 2.3 接入 Redis 缓存（V1 要求）
- 先在 `docker-compose.yml` 加 redis 服务
- 缓存热点读接口：文章详情、文章列表
- 关键要自己搞懂：缓存读写时机、缓存与 DB 一致性（更新文章时怎么让缓存失效）、缓存穿透/雪崩
- 📦 我可代劳：redis 客户端封装、config 加 redis 段、compose 配置
- 📖 你要懂：缓存策略本身（这是面试高频，别让 AI 全代）

### 📦📖 2.4 一个真实的 MQ 异步业务（V1 要求）
挑一个真的适合异步的场景，别为了用而用。候选：
- 文章浏览量 `view_count` 累加：每次访问发消息，消费端批量落库（避免高频写 DB）
- 或"发布文章后异步生成/推送"之类
- 📖 你要懂：为什么用 MQ（削峰/解耦）、消息可靠性、消费幂等
- 📦 我可代劳：MQ 客户端封装、compose 加 MQ 服务

---

## 阶段三：Docker 化 → CI/CD → 上线（你最想补的 CD）

### 🔧📖 3.1 后端 Dockerfile（当前缺失，闭环断点）
现在 compose 只有 MySQL，**应用本体没镜像**，这个不补 CD 无从谈起。
- 写多阶段构建 Dockerfile（builder 阶段 `go build`，运行阶段用 alpine 装二进制，镜像小）
- config 要能被环境变量覆盖（viper.AutomaticEnv 已开，但 key 映射要验证）
- 把 backend 服务加进 docker-compose，和 mysql/redis 组网
- 📖 你要懂：多阶段构建为什么、容器间怎么用服务名互相访问（host 不再是 127.0.0.1）

### 🔧📖 3.2 CD（GitHub Actions 自动部署）
现在 CI 只到 build/test（`.github/workflows/ci.yml`）。CD 要加：
- 构建镜像 → 推到镜像仓库（GHCR / DockerHub）
- SSH 到 Linux 服务器 → 拉新镜像 → `docker compose up -d` 重启
- 密钥（服务器 IP/SSH key/仓库密码）放 GitHub Secrets
- 📖 这块你重点练：整条 push→自动上线的链路，是你路线图的核心目标

### 📖 3.3 服务器 + 公网访问
- Linux 服务器装 Docker
- Nginx 反向代理（80/443 → 后端 8080）
- 域名解析 + HTTPS（Let's Encrypt / certbot）
- 📖 全流程你自己走一遍

---

## 阶段四：上线后加固（有余力再做）
- 结构化日志（替换 `log.Printf`，如 zap/slog）
- 请求日志中间件、panic recover 中间件
- 配置分环境（dev/prod 分文件）
- 限流、监控（Prometheus + Grafana）

---

## 备注
- `server.exe` 是编译产物，建议加进 `.gitignore`，别提交
- `jwt.secret` 现在是 `"123"`，上线前**务必**用环境变量换成强密钥
- `admin.password` 同理，生产别用默认值
