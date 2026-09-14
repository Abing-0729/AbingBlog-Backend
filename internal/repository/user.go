package repository

import (
	"abingblog-backend/internal/model"
	"errors"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// GetByUsername 按用户名查用户；不存在返回 nil, nil
func (r *UserRepo) GetByUsername(username string) (*model.User, error) {
	var u model.User
	err := r.db.Where("username = ?", username).First(&u).Error
	if errors.Is(gorm.ErrRecordNotFound, err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &u, nil
}

// Create 创建用户
func (r *UserRepo) Create(u *model.User) error {
	return r.db.Create(u).Error
}
