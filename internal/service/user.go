package service

import (
	"abingblog-backend/internal/errcode"
	"log"
	"strings"

	"abingblog-backend/internal/auth"
	"abingblog-backend/internal/config"
	"abingblog-backend/internal/model"
	"abingblog-backend/internal/repository"
)

// UserService 登录 / 账号相关业务
type UserService struct {
	repo *repository.UserRepo
	cfg  *config.Config
}

func NewUserService(repo *repository.UserRepo, cfg *config.Config) *UserService {
	return &UserService{repo: repo, cfg: cfg}
}

// LoginReq 登录请求体
type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// SeedAdmin 启动时确保初始管理员存在：账号已存在则跳过，否则用 config.Admin 建号（密码 bcrypt 入库）。
// 这段是样板，已实现完整。
func (s *UserService) SeedAdmin() error {
	username := strings.TrimSpace(s.cfg.Admin.Username)
	if username == "" || s.cfg.Admin.Password == "" {
		log.Println("未配置 admin.username/password，跳过初始管理员创建")
		return nil
	}

	exist, err := s.repo.GetByUsername(username)
	if err != nil {
		return err
	}
	if exist != nil {
		return nil // 已存在，不重复建号，也不覆盖密码
	}

	hashed, err := auth.HashPassword(s.cfg.Admin.Password)
	if err != nil {
		return err
	}
	if err := s.repo.Create(&model.User{Username: username, Password: hashed}); err != nil {
		return err
	}
	log.Printf("已创建初始管理员账号: %s", username)
	return nil
}

// Login 校验用户名密码，成功返回签发的 JWT token。

func (s *UserService) Login(req LoginReq) (string, error) {
	user, err := s.repo.GetByUsername(req.Username)

	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errcode.New(errcode.CodeLoginFail, "用户名或密码错误")
	}
	if !auth.CheckPassword(user.Password, req.Password) {
		return "", errcode.New(errcode.CodeLoginFail, "用户名或密码错误")
	}

	token, err := auth.GenerateToken(user.ID, s.cfg.JWT.Secret, s.cfg.JWT.ExpireHours)
	if err != nil {
		return "", err
	}
	return token, nil
}
