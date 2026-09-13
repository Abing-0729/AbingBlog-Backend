package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

// GenerateToken 登录成功后签发 JWT，返回 token 字符串。
// 参数：userID 用户主键，secret 签名密钥，expireHours 有效期（小时）。
func GenerateToken(userID uint, secret string, expireHours int) (string, error) {
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken 校验并解析 token，成功返回 userID。
// 失败（过期/签名不对/格式错）返回 error，中间件据此返回 401。
// ParseToken 的思路（验签是关键，别漏）：
//   - jwt.ParseWithClaims(tokenStr, &Claims{}, keyFunc)
//   - keyFunc 里务必校验 token.Method 是不是 *jwt.SigningMethodHMAC，
//     否则会被 alg=none / 换算法攻击绕过验签
//   - 返回 []byte(secret) 作为验签密钥
//   - 检查 token.Valid，取出 *Claims 返回
func ParseToken(tokenStr, secret string) (uint, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}
	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}
	return token.Claims.(*Claims).UserID, nil
}
