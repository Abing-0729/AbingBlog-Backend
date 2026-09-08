package repository

import (
	"errors"

	"abingblog-backend/internal/model"

	"gorm.io/gorm"
)

// CategoryRepo 分类数据访问层
type CategoryRepo struct {
	db *gorm.DB
}

func NewCategoryRepo(db *gorm.DB) *CategoryRepo { return &CategoryRepo{db: db} }

// CategoryWithCount 分类 + 已发布文章数（公共分类列表接口的返回结构）
type CategoryWithCount struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	Sort         int    `json:"sort"`
	ArticleCount int64  `json:"article_count"`
}

// ListWithCount 分类列表，附带每个分类下「已发布且未删除」文章的数量
func (r *CategoryRepo) ListWithCount() ([]CategoryWithCount, error) {
	list := make([]CategoryWithCount, 0) // 初始化为空切片，空列表也序列化为 [] 而不是 null
	err := r.db.Model(&model.Category{}).
		Select("categories.*, COUNT(articles.id) AS article_count").
		Joins("LEFT JOIN articles ON articles.category_id = categories.id AND articles.status = ? AND articles.deleted_at IS NULL", "published").
		Group("categories.id").
		Order("categories.sort ASC, categories.id ASC").
		Scan(&list).Error
	return list, err
}

func (r *CategoryRepo) GetByID(id uint) (*model.Category, error) {
	var c model.Category
	err := r.db.First(&c, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *CategoryRepo) GetByName(name string) (*model.Category, error) {
	var c model.Category
	err := r.db.Where("name = ?", name).First(&c).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &c, err
}

func (r *CategoryRepo) Create(c *model.Category) error {
	return r.db.Create(c).Error
}

func (r *CategoryRepo) Update(c *model.Category) error {
	return r.db.Model(c).Updates(map[string]any{
		"name": c.Name,
		"slug": c.Slug,
		"sort": c.Sort,
	}).Error
}

// Delete 删除分类。
// 软删除的文章行仍在表里且带着 category_id，外键会阻止删分类；
// 所以先永久清掉该分类下「已软删」的文章（用户本来就已经删了它们）——
// 顺序：清中间表关联 → 硬删文章 → 删分类，每一步都有外键依赖。
// 活跃文章由 service 层检查保证不存在。
func (r *CategoryRepo) Delete(id uint) error {
	if err := r.db.Exec(
		"DELETE FROM article_tags WHERE article_id IN (SELECT id FROM articles WHERE category_id = ? AND deleted_at IS NOT NULL)",
		id,
	).Error; err != nil {
		return err
	}
	if err := r.db.Unscoped().
		Where("category_id = ? AND deleted_at IS NOT NULL", id).
		Delete(&model.Article{}).Error; err != nil {
		return err
	}
	return r.db.Delete(&model.Category{}, id).Error
}

// CountByCategory 分类下的文章数（软删除的文章 GORM 自动排除）
func (r *CategoryRepo) CountByCategory(categoryID uint) (int64, error) {
	var n int64
	err := r.db.Model(&model.Article{}).Where("category_id = ?", categoryID).Count(&n).Error
	return n, err
}
