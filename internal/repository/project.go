package repository

import (
	"errors"

	"abingblog-backend/internal/model"

	"gorm.io/gorm"
)

// ProjectRepo 项目数据访问层：只负责 SQL，不含业务逻辑
type ProjectRepo struct {
	db *gorm.DB
}

func NewProjectRepo(db *gorm.DB) *ProjectRepo { return &ProjectRepo{db: db} }

// List 分页查询项目；status 为空表示不按状态过滤（后台用）。
// 排序沿用 sort 升序 + id 降序：sort 定优先级，同权重时新项目靠前。
func (r *ProjectRepo) List(page, pageSize int, status string) ([]model.Project, int64, error) {
	query := r.db.Model(&model.Project{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	list := make([]model.Project, 0)
	err := query.Session(&gorm.Session{}).
		Order("sort ASC, id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Find(&list).Error
	return list, total, err
}

// GetByID 按 ID 查项目，不限状态（后台编辑用）
func (r *ProjectRepo) GetByID(id uint) (*model.Project, error) {
	var p model.Project
	err := r.db.First(&p, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

// GetBySlug 按 slug 查项目（唯一性预检查用）
func (r *ProjectRepo) GetBySlug(slug string) (*model.Project, error) {
	var p model.Project
	err := r.db.Where("slug = ?", slug).First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &p, err
}

func (r *ProjectRepo) Create(p *model.Project) error {
	return r.db.Create(p).Error
}

// Update 全量更新业务字段（用 map 保证空值也能覆盖）
func (r *ProjectRepo) Update(p *model.Project) error {
	return r.db.Model(p).Updates(map[string]any{
		"slug":       p.Slug,
		"name":       p.Name,
		"detail":     p.Detail,
		"stack":      p.Stack,
		"github_url": p.GithubURL,
		"demo_url":   p.DemoURL,
		"sort":       p.Sort,
		"status":     p.Status,
	}).Error
}

// SoftDelete 软删除项目（模型带 DeletedAt，Delete 自动打删除标记）
func (r *ProjectRepo) SoftDelete(id uint) error {
	return r.db.Delete(&model.Project{}, id).Error
}
