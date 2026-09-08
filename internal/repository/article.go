package repository

import (
	"errors"
	"time"

	"abingblog-backend/internal/model"

	"gorm.io/gorm"
)

// ArticleRepo 文章数据访问层：只负责 SQL，不含业务逻辑
type ArticleRepo struct {
	db *gorm.DB
}

func NewArticleRepo(db *gorm.DB) *ArticleRepo { return &ArticleRepo{db: db} }

// buildQuery 组装公共查询条件（列表接口的 count 和 find 共用，保证条件一致）
func (r *ArticleRepo) buildQuery(status, categorySlug, tag, keyword string) *gorm.DB {
	query := r.db.Model(&model.Article{})
	if status != "" {
		query = query.Where("articles.status = ?", status)
	}
	if categorySlug != "" {
		query = query.Joins("JOIN categories ON categories.id = articles.category_id").
			Where("categories.slug = ?", categorySlug)
	}
	if tag != "" {
		query = query.Joins("JOIN article_tags ON article_tags.article_id = articles.id").
			Joins("JOIN tags ON tags.id = article_tags.tag_id").
			Where("tags.name = ?", tag)
	}
	if keyword != "" {
		query = query.Where("articles.title LIKE ?", "%"+keyword+"%")
	}
	return query
}

// List 分页查询文章；status 为空表示不按状态过滤
func (r *ArticleRepo) List(page, pageSize int, status, categorySlug, tag, keyword string) ([]model.Article, int64, error) {
	query := r.buildQuery(status, categorySlug, tag, keyword)

	// 带 JOIN 时用 DISTINCT 计数：多对多关联会把同一篇文章展开成多行
	var total int64
	if err := query.Distinct("articles.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// GROUP BY 主键去重（MySQL 允许按主键分组后取整行）；Preload 是独立查询，不受影响
	var list []model.Article
	err := query.Group("articles.id").
		Order("articles.published_at DESC, articles.id DESC").
		Offset((page - 1) * pageSize).Limit(pageSize).
		Preload("Category").Preload("Tags").
		Find(&list).Error
	return list, total, err
}

// GetPublishedByID 按 ID 查已发布文章（公共详情接口用）
func (r *ArticleRepo) GetPublishedByID(id uint) (*model.Article, error) {
	var a model.Article
	err := r.db.Preload("Category").Preload("Tags").
		Where("id = ? AND status = ?", id, "published").
		First(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &a, err
}

// GetByID 按 ID 查文章，不限状态（后台编辑用）
func (r *ArticleRepo) GetByID(id uint) (*model.Article, error) {
	var a model.Article
	err := r.db.Preload("Category").Preload("Tags").First(&a, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &a, err
}

// Create 创建文章；Omit("Tags") 让标签关联走后面的 ReplaceTags 显式处理
func (r *ArticleRepo) Create(a *model.Article) error {
	return r.db.Omit("Tags").Create(a).Error
}

// Update 全量更新业务字段（用 map 保证空值也能覆盖）
func (r *ArticleRepo) Update(a *model.Article) error {
	return r.db.Model(a).Omit("Tags").Updates(map[string]any{
		"title":        a.Title,
		"content":      a.Content,
		"summary":      a.Summary,
		"cover":        a.Cover,
		"category_id":  a.CategoryID,
		"status":       a.Status,
		"published_at": a.PublishedAt,
	}).Error
}

// UpdateStatus 发布/撤回：publishedAt 为 nil 表示撤回（清空发布时间）
func (r *ArticleRepo) UpdateStatus(id uint, status string, publishedAt *time.Time) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).
		Updates(map[string]any{"status": status, "published_at": publishedAt}).Error
}

// ReplaceTags 重建文章与标签的关联（先删后插，保证和传入列表完全一致）
func (r *ArticleRepo) ReplaceTags(a *model.Article, tags []model.Tag) error {
	return r.db.Model(a).Association("Tags").Replace(tags)
}

// SoftDelete 软删除：模型带 DeletedAt 字段时，GORM 的 Delete 自动打删除标记
func (r *ArticleRepo) SoftDelete(id uint) error {
	return r.db.Delete(&model.Article{}, id).Error
}

// IncrViewCount 浏览量 +1（v1 同步更新；后期改为 MQ 异步落库，见 docs/api-v1.md §5）
func (r *ArticleRepo) IncrViewCount(id uint) error {
	return r.db.Model(&model.Article{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}
