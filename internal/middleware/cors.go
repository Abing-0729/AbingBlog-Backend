package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Cors 返回一个跨域中间件。浏览器同源策略会拦截"前端源 != 后端源"的请求，
// 需要后端在响应头里显式声明允许哪些源、哪些方法、哪些请求头。
//
// allowOrigins 是允许的前端来源白名单（如 http://localhost:5173）。
// 只回显命中白名单的 Origin，而不是无脑返回 "*"：因为一旦要携带凭证
// （Authorization 头 / Cookie），规范禁止用 "*"，必须回显具体源。
func Cors(allowOrigins []string) gin.HandlerFunc {
	// 白名单转成 set，O(1) 命中判断
	allowed := make(map[string]struct{}, len(allowOrigins))
	for _, o := range allowOrigins {
		allowed[o] = struct{}{}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		// 命中白名单才回写 CORS 头；非跨域请求（无 Origin）或不在白名单的源直接放行给后续逻辑，
		// 浏览器那边拿不到放行头自然会拦，不需要后端主动报错
		if _, ok := allowed[origin]; ok && origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			// 回显了具体源 + 允许携带凭证，前端才能带上 Authorization
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			// 告知浏览器：响应会因 Origin 不同而不同，避免缓存串源
			c.Header("Vary", "Origin")
		}

		// 预检请求（浏览器在真正的跨域带头请求前先发一个 OPTIONS 探路），
		// 直接 204 结束，不进入业务 handler
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
