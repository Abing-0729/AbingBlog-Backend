package service

import (
	"strings"
	"time"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/model"
	"abingblog-backend/internal/repository"
)

// ArticleService 文章业务逻辑层：参数校验、状态规则、数据组装
type ArticleService struct {
	article  *repository.ArticleRepo
	category *repository.CategoryRepo
	tag      *repository.TagRepo
}

func NewArticleService(article *repository.ArticleRepo, category *repository.CategoryRepo, tag *repository.TagRepo) *ArticleService {
	return &ArticleService{article: article, category: category, tag: tag}
}

// ArticleQuery 列表查询条件（公共接口固定 Status=published，后台可选）
type ArticleQuery struct {
	Page         int
	PageSize     int
	Status       string
	CategorySlug string
	Tag          string
	Keyword      string
}

// List 文章列表：默认分页参数，透传查询条件
func (s *ArticleService) List(q ArticleQuery) ([]model.Article, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 50 {
		q.PageSize = 10
	}
	return s.article.List(q.Page, q.PageSize, q.Status, q.CategorySlug, q.Tag, q.Keyword)
}

// GetByID 后台按 ID 查文章：不限状态（含草稿）、带正文、不自增浏览量。
// 后台编辑时用，区别于公共详情的 GetPublicByID。
func (s *ArticleService) GetByID(id uint) (*model.Article, error) {
	a, err := s.article.GetByID(id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errcode.New(errcode.CodeArticleNotFound, "文章不存在")
	}
	return a, nil
}

// GetPublicByID 公共详情：只允许已发布文章；浏览量 +1
func (s *ArticleService) GetPublicByID(id uint) (*model.Article, error) {
	a, err := s.article.GetPublishedByID(id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errcode.New(errcode.CodeArticleNotFound, "文章不存在")
	}
	// v1 先同步自增；后期改为 Redis 计数 + MQ 异步落库（docs/api-v1.md §5）
	_ = s.article.IncrViewCount(id)
	a.ViewCount++
	return a, nil
}

// CreateArticleReq 创建/更新文章的请求体（前后端契约的一部分）
type CreateArticleReq struct {
	Title      string `json:"title"`
	Content    string `json:"content"`
	Summary    string `json:"summary"`
	Cover      string `json:"cover"`
	CategoryID uint   `json:"category_id"` // 0 表示未分类
	TagIDs     []uint `json:"tag_ids"`
	Status     string `json:"status"` // draft / published，缺省 draft
}

// validate 公共校验：标题、状态、分类存在性
func (s *ArticleService) validate(req *CreateArticleReq) error {
	if strings.TrimSpace(req.Title) == "" {
		return errcode.New(errcode.CodeTitleEmpty, "文章标题不能为空")
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	if req.Status != "draft" && req.Status != "published" {
		return errcode.New(errcode.CodeBadStatus, "状态只能为 draft 或 published")
	}
	if req.CategoryID != 0 {
		c, err := s.category.GetByID(req.CategoryID)
		if err != nil {
			return err
		}
		if c == nil {
			return errcode.New(errcode.CodeCategoryNotFound, "分类不存在")
		}
	}
	return nil
}

// checkTags 校验标签都存在并返回标签实体
func (s *ArticleService) checkTags(tagIDs []uint) ([]model.Tag, error) {
	tags, err := s.tag.GetByIDs(tagIDs)
	if err != nil {
		return nil, err
	}
	if len(tags) != len(tagIDs) {
		return nil, errcode.New(errcode.CodeTagNotFound, "标签不存在")
	}
	return tags, nil
}

// Create 创建文章：发布时记录首次发布时间；标签关联在文章落库后重建
func (s *ArticleService) Create(req CreateArticleReq) (*model.Article, error) {
	if err := s.validate(&req); err != nil {
		return nil, err
	}
	tags, err := s.checkTags(req.TagIDs)
	if err != nil {
		return nil, err
	}

	a := &model.Article{
		Title:      req.Title,
		Content:    req.Content,
		Summary:    summaryOf(req.Summary, req.Content),
		Cover:      req.Cover,
		CategoryID: req.CategoryID,
		Status:     req.Status,
	}
	if req.Status == "published" {
		a.PublishedAt = new(time.Now())
	}
	if err := s.article.Create(a); err != nil {
		return nil, err
	}
	if err := s.article.ReplaceTags(a, tags); err != nil {
		return nil, err
	}
	a.Tags = tags
	return a, nil
}

// Update 全量更新：发布时补首次发布时间，撤回草稿则清空
func (s *ArticleService) Update(id uint, req CreateArticleReq) (*model.Article, error) {
	if err := s.validate(&req); err != nil {
		return nil, err
	}
	a, err := s.article.GetByID(id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errcode.New(errcode.CodeArticleNotFound, "文章不存在")
	}
	tags, err := s.checkTags(req.TagIDs)
	if err != nil {
		return nil, err
	}

	a.Title = req.Title
	a.Content = req.Content
	a.Summary = summaryOf(req.Summary, req.Content)
	a.Cover = req.Cover
	a.CategoryID = req.CategoryID
	a.Status = req.Status
	switch req.Status {
	case "published":
		if a.PublishedAt == nil { // 保留首次发布时间，重新发布不刷新
			a.PublishedAt = new(time.Now())
		}
	case "draft":
		a.PublishedAt = nil
	}

	if err := s.article.Update(a); err != nil {
		return nil, err
	}
	if err := s.article.ReplaceTags(a, tags); err != nil {
		return nil, err
	}
	a.Tags = tags
	return a, nil
}

// Delete 软删除文章（数据保留，可恢复）
func (s *ArticleService) Delete(id uint) error {
	a, err := s.article.GetByID(id)
	if err != nil {
		return err
	}
	if a == nil {
		return errcode.New(errcode.CodeArticleNotFound, "文章不存在")
	}
	return s.article.SoftDelete(id)
}

// UpdateStatus 发布/撤回草稿（独立接口，后台列表里一键切换）
func (s *ArticleService) UpdateStatus(id uint, status string) (*model.Article, error) {
	if status != "draft" && status != "published" {
		return nil, errcode.New(errcode.CodeBadStatus, "状态只能为 draft 或 published")
	}
	a, err := s.article.GetByID(id)
	if err != nil {
		return nil, err
	}
	if a == nil {
		return nil, errcode.New(errcode.CodeArticleNotFound, "文章不存在")
	}
	var publishedAt *time.Time
	if status == "published" && a.PublishedAt == nil {
		publishedAt = new(time.Now())
	}
	if err := s.article.UpdateStatus(id, status, publishedAt); err != nil {
		return nil, err
	}
	a.Status = status
	if publishedAt != nil {
		a.PublishedAt = publishedAt
	}
	return a, nil
}

// summaryOf 摘要：优先用请求里的；否则把正文压缩成一行并截取前 200 字符
func summaryOf(custom, content string) string {
	if strings.TrimSpace(custom) != "" {
		return custom
	}
	flat := strings.Join(strings.Fields(content), " ")
	runes := []rune(flat)
	if len(runes) > 200 {
		return string(runes[:200]) + "..."
	}
	return flat
}
