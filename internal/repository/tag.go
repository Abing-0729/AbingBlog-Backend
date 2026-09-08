package repository

import (
	"errors"

	"abingblog-backend/internal/model"

	"gorm.io/gorm"
)

// TagRepo 标签数据访问层
type TagRepo struct {
	db *gorm.DB
}

func NewTagRepo(db *gorm.DB) *TagRepo { return &TagRepo{db: db} }

// TagWithCount 标签 + 已发布文章数（公共标签列表接口的返回结构）
type TagWithCount struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	ArticleCount int64  `json:"article_count"`
}

// ListWithCount 标签列表，附带每个标签下「已发布且未删除」文章的数量
func (r *TagRepo) ListWithCount() ([]TagWithCount, error) {
	list := make([]TagWithCount, 0) // 初始化为空切片，空列表也序列化为 [] 而不是 null
	err := r.db.Model(&model.Tag{}).
		Select("tags.*, COUNT(article_tags.article_id) AS article_count").
		Joins("LEFT JOIN article_tags ON article_tags.tag_id = tags.id").
		Joins("LEFT JOIN articles ON articles.id = article_tags.article_id AND articles.status = ? AND articles.deleted_at IS NULL", "published").
		Group("tags.id").
		Order("tags.id ASC").
		Scan(&list).Error
	return list, err
}

func (r *TagRepo) GetByID(id uint) (*model.Tag, error) {
	var t model.Tag
	err := r.db.First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &t, err
}

func (r *TagRepo) GetByName(name string) (*model.Tag, error) {
	var t model.Tag
	err := r.db.Where("name = ?", name).First(&t).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &t, err
}

// GetByIDs 批量查标签（创建/更新文章时校验 tag_ids 是否存在）
func (r *TagRepo) GetByIDs(ids []uint) ([]model.Tag, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var tags []model.Tag
	err := r.db.Where("id IN ?", ids).Find(&tags).Error
	return tags, err
}

func (r *TagRepo) Create(t *model.Tag) error {
	return r.db.Create(t).Error
}

func (r *TagRepo) Update(t *model.Tag) error {
	return r.db.Model(t).Updates(map[string]any{"name": t.Name}).Error
}

// Delete 删除标签。
// 直接删中间表全部关联行（含指向已软删文章的），否则外键会阻止删标签；
// GORM 的 Association 操作默认会跳过软删文章，清不干净。
func (r *TagRepo) Delete(id uint) error {
	if err := r.db.Exec("DELETE FROM article_tags WHERE tag_id = ?", id).Error; err != nil {
		return err
	}
	return r.db.Delete(&model.Tag{}, id).Error
}
