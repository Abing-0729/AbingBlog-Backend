package service

import (
	"time"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepo
	secret    []byte
	expiresIn int
}

func NewAuthService(userRepo *repository.UserRepo, secret string, expiresIn int) *AuthService {
	if expiresIn <= 0 {
		expiresIn = 86400
	}
	return &AuthService{
		userRepo:  userRepo,
		secret:    []byte(secret),
		expiresIn: expiresIn,
	}
}

// Login 验证用户名/密码，成功返回JWTtoken有效期
func (s *AuthService) Login(username, password string) (string, int, error) {
	u, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return "", 0, err
	}
	if u == nil {
		return "", 0, errcode.New(errcode.CodeLoginFail, "用户名或密码错误")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password)); err != nil {
		return "", 0, errcode.New(errcode.CodeLoginFail, "用户名或密码错误")
	}
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   u.Username,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.expiresIn) * time.Second)),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", 0, err
	}
	return token, s.expiresIn, nil
}
