package repository

import (
	"errors"

	"abingblog-backend/internal/model"

	"gorm.io/gorm"
)

// UserRepo 用户数据访问层。博客只有站长一个账号，接口保持最小
type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

// GetByUsername 按用户名查用户，不存在返回 (nil, nil)（与 TagRepo.GetByName 一致的约定）
func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Where("username = ?", username).First(&u).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// GetByID 按主键查用户（中间件校验后可用 user_id 反查当前用户），不存在返回 (nil, nil)
func (r *UserRepo) GetByID(id uint) (*model.User, error) {
	var u model.User
	err := r.db.First(&u, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &u, err
}

// Create 建号。password 字段必须是已 bcrypt 哈希后的值，绝不传明文
func (r *UserRepo) Create(u *model.User) error {
	return r.db.Create(u).Error
}
