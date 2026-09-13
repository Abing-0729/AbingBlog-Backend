package middleware

import (
	"abingblog-backend/internal/auth"
	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"
	"github.com/gin-gonic/gin"
	"strings"
)

// AuthRequired 返回一个Gin中间件，校验JWT token
func AuthRequired(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 取 Authorization 头，格式必须是 "Bearer <token>"
		authHeader := c.GetHeader("Authorization")
		// 2. 空 或 不以 "Bearer " 开头 → 401 + return（Abort 后必须 return，否则继续往下走）
		if authHeader == "" {
			response.Fail(c, 401, errcode.CodeLoginFail, "令牌有误")
			return
		}
		// 3. 切出 token 字符串（strings.TrimPrefix 去掉 "Bearer "）
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == authHeader {
			response.Fail(c, 401, errcode.CodeLoginFail, "登入失败")
			return
		}
		// 4. auth.ParseToken(token, secret)，出错 → 401 + return
		UserID, err := auth.ParseToken(token, secret)
		if err != nil {
			response.Fail(c, 401, errcode.CodeLoginFail, "登入失败")
			return
		}
		// 5. 成功：c.Set("user_id", userID) 存进上下文供后续 handler 用
		c.Set("user_id", UserID)
		c.Next()
		// 6. c.Next() 放行
	}
}
