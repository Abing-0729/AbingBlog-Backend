package service

import (
	"strings"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/model"
	"abingblog-backend/internal/repository"
)

// ProjectService 项目业务逻辑层：参数校验、状态规则、slug 唯一性
type ProjectService struct {
	repo *repository.ProjectRepo
}

func NewProjectService(repo *repository.ProjectRepo) *ProjectService {
	return &ProjectService{repo: repo}
}

// ProjectQuery 列表查询条件（公共接口固定 Status=published，后台可选）
type ProjectQuery struct {
	Page     int
	PageSize int
	Status   string
}

// List 项目列表：默认分页参数
func (s *ProjectService) List(q ProjectQuery) ([]model.Project, int64, error) {
	if q.Page < 1 {
		q.Page = 1
	}
	if q.PageSize < 1 || q.PageSize > 50 {
		q.PageSize = 10
	}
	return s.repo.List(q.Page, q.PageSize, q.Status)
}

// ProjectReq 创建/更新项目的请求体（前后端契约的一部分）
type ProjectReq struct {
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Detail    string `json:"detail"`
	Stack     string `json:"stack"`
	GithubURL string `json:"github_url"`
	DemoURL   string `json:"demo_url"`
	Sort      int    `json:"sort"`
	Status    string `json:"status"` // draft / published，缺省 draft
}

// validate 公共校验：slug、name 必填，status 合法
func (s *ProjectService) validate(req *ProjectReq) error {
	if strings.TrimSpace(req.Slug) == "" {
		return errcode.New(errcode.CodeSlugEmpty, "项目 slug 不能为空")
	}
	if strings.TrimSpace(req.Name) == "" {
		return errcode.New(errcode.CodeNameEmpty, "项目名称不能为空")
	}
	if req.Status == "" {
		req.Status = "draft"
	}
	if req.Status != "draft" && req.Status != "published" {
		return errcode.New(errcode.CodeBadStatus, "状态只能为 draft 或 published")
	}
	return nil
}

// Create 创建项目：slug 唯一性预检查，给前端明确的业务错误
func (s *ProjectService) Create(req ProjectReq) (*model.Project, error) {
	if err := s.validate(&req); err != nil {
		return nil, err
	}
	slug := strings.TrimSpace(req.Slug)
	if exist, err := s.repo.GetBySlug(slug); err != nil {
		return nil, err
	} else if exist != nil {
		return nil, errcode.New(errcode.CodeSlugExists, "项目 slug 已存在")
	}

	p := &model.Project{
		Slug:      slug,
		Name:      strings.TrimSpace(req.Name),
		Detail:    req.Detail,
		Stack:     req.Stack,
		GithubURL: req.GithubURL,
		DemoURL:   req.DemoURL,
		Sort:      req.Sort,
		Status:    req.Status,
	}
	if err := s.repo.Create(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Update 全量更新：slug 唯一性排除自身
func (s *ProjectService) Update(id uint, req ProjectReq) (*model.Project, error) {
	if err := s.validate(&req); err != nil {
		return nil, err
	}
	p, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, errcode.New(errcode.CodeProjectNotFound, "项目不存在")
	}
	slug := strings.TrimSpace(req.Slug)
	if exist, err := s.repo.GetBySlug(slug); err != nil {
		return nil, err
	} else if exist != nil && exist.ID != id {
		return nil, errcode.New(errcode.CodeSlugExists, "项目 slug 已存在")
	}

	p.Slug = slug
	p.Name = strings.TrimSpace(req.Name)
	p.Detail = req.Detail
	p.Stack = req.Stack
	p.GithubURL = req.GithubURL
	p.DemoURL = req.DemoURL
	p.Sort = req.Sort
	p.Status = req.Status
	if err := s.repo.Update(p); err != nil {
		return nil, err
	}
	return p, nil
}

// Delete 软删除项目（数据保留，可恢复）
func (s *ProjectService) Delete(id uint) error {
	p, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if p == nil {
		return errcode.New(errcode.CodeProjectNotFound, "项目不存在")
	}
	return s.repo.SoftDelete(id)
}
