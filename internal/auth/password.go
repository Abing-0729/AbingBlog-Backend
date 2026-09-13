package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword 把明文密码做 bcrypt 哈希，入库只存哈希值，绝不存明文。
// bcrypt 自带盐，同一密码每次哈希结果都不同，无需自己拼盐。
func HashPassword(plain string) (string, error) {
	// DefaultCost=10，够用；成本越高越慢越安全，登录接口能接受
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// CheckPassword 校验明文与库里哈希是否匹配。匹配返回 true。
// 注意：不要自己比较字符串，bcrypt 每次哈希不同，必须用这个函数比对。
func CheckPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
