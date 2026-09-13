package middle

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const JWTSubjectKey = "jwt_subject"

// JWTAuth 校验 Authorization: Bearer <token>；无效 token 直接中止请求。
func JWTAuth(secret string) gin.HandlerFunc {
	key := []byte(secret)
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			unauthorized(c, "未提供有效的认证 token")
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
		if tokenStr == "" {
			unauthorized(c, "未提供有效的认证 token")
			return
		}
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if t.Method != jwt.SigningMethodHS256 {
				return nil, jwt.ErrSignatureInvalid
			}
			return key, nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil || !token.Valid {
			unauthorized(c, "token 无效或已过期")
			return
		}
		subject, err := token.Claims.GetSubject()
		if err != nil || subject == "" {
			unauthorized(c, "token 缺少用户标识")
			return
		}
		c.Set(JWTSubjectKey, subject)
		c.Next()
	}
}

func unauthorized(c *gin.Context, message string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    1401,
		"message": message,
		"data":    nil,
	})
}
