package handler

import (
	"time"

	"abingblog-backend/internal/model"
)

// 对外 DTO：数据库模型（model）和接口返回（DTO）分离，
// 表结构怎么变都不直接影响前端拿到的 JSON 结构

type categoryDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type tagDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type articleDTO struct {
	ID          uint         `json:"id"`
	Title       string       `json:"title"`
	Summary     string       `json:"summary"`
	Cover       string       `json:"cover"`
	Content     string       `json:"content,omitempty"` // 只有详情/后台接口带正文
	Category    *categoryDTO `json:"category,omitempty"`
	Tags        []tagDTO     `json:"tags"`
	Status      string       `json:"status"`
	ViewCount   int          `json:"view_count"`
	PublishedAt *time.Time   `json:"published_at,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

func toCategoryDTO(c *model.Category) *categoryDTO {
	if c == nil {
		return nil
	}
	return &categoryDTO{ID: c.ID, Name: c.Name, Slug: c.Slug}
}

func toTagDTOs(tags []model.Tag) []tagDTO {
	out := make([]tagDTO, 0, len(tags))
	for _, t := range tags {
		out = append(out, tagDTO{ID: t.ID, Name: t.Name})
	}
	return out
}

// toArticleDTO 模型 → DTO；withContent 控制是否带 Markdown 正文（列表不返回，省流量）
func toArticleDTO(a *model.Article, withContent bool) articleDTO {
	d := articleDTO{
		ID:          a.ID,
		Title:       a.Title,
		Summary:     a.Summary,
		Cover:       a.Cover,
		Category:    toCategoryDTO(a.Category),
		Tags:        toTagDTOs(a.Tags),
		Status:      a.Status,
		ViewCount:   a.ViewCount,
		PublishedAt: a.PublishedAt,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
	if withContent {
		d.Content = a.Content
	}
	return d
}

type projectDTO struct {
	ID        uint      `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	Detail    string    `json:"detail"`
	Stack     string    `json:"stack"`
	GithubURL string    `json:"github_url"`
	DemoURL   string    `json:"demo_url"`
	Sort      int       `json:"sort"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// toProjectDTO 模型 → DTO
func toProjectDTO(p *model.Project) projectDTO {
	return projectDTO{
		ID:        p.ID,
		Slug:      p.Slug,
		Name:      p.Name,
		Detail:    p.Detail,
		Stack:     p.Stack,
		GithubURL: p.GithubURL,
		DemoURL:   p.DemoURL,
		Sort:      p.Sort,
		Status:    p.Status,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}
