package repository

import (
	"errors"

	"abingblog-backend/internal/model"

	"gorm.io/gorm"
)

// FriendLinkRepo 友链数据访问层：只负责 SQL，不含业务逻辑
type FriendLinkRepo struct {
	db *gorm.DB
}

func NewFriendLinkRepo(db *gorm.DB) *FriendLinkRepo {
	return &FriendLinkRepo{db: db}
}

// ListEnabled 前台展示：只查已上架的，按 sort 升序 + 新的靠前
func (r *FriendLinkRepo) ListEnabled() ([]model.FriendLink, error) {
	list := make([]model.FriendLink, 0)
	err := r.db.Where("status = ?", model.FriendLinkApproved).
		Order("sort ASC, id DESC").
		Find(&list).Error
	return list, err
}

// ListByOwner 访客"我的友链"：不限状态，按提交时间倒序
func (r *FriendLinkRepo) ListByOwner(ownerID string) ([]model.FriendLink, error) {
	list := make([]model.FriendLink, 0)
	err := r.db.Where("owner_id = ?", ownerID).
		Order("id DESC").
		Find(&list).Error
	return list, err
}

// ListAll 后台管理：status 为负数表示不过滤（看全部），否则按状态筛选
func (r *FriendLinkRepo) ListAll(status int) ([]model.FriendLink, error) {
	query := r.db.Model(&model.FriendLink{})
	if status >= 0 {
		query = query.Where("status = ?", status)
	}
	list := make([]model.FriendLink, 0)
	err := query.Order("status ASC, id DESC").Find(&list).Error
	return list, err
}

// FindByID 按主键查；查不到返回 (nil, nil)，与"数据库故障"区分开
func (r *FriendLinkRepo) FindByID(id uint) (*model.FriendLink, error) {
	var f model.FriendLink
	err := r.db.First(&f, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &f, err
}

// Insert 新增友链
func (r *FriendLinkRepo) Insert(f *model.FriendLink) error {
	return r.db.Create(f).Error
}

// Update 全量更新业务字段（map 保证零值也能覆盖到库）
func (r *FriendLinkRepo) Update(f *model.FriendLink) error {
	return r.db.Model(f).Updates(map[string]any{
		"name":        f.Name,
		"avatar":      f.Avatar,
		"url":         f.URL,
		"description": f.Description,
		"status":      f.Status,
		"sort":        f.Sort,
	}).Error
}

// Delete 软删除（模型带 DeletedAt，Delete 自动打删除标记）
func (r *FriendLinkRepo) Delete(id uint) error {
	return r.db.Delete(&model.FriendLink{}, id).Error
}
