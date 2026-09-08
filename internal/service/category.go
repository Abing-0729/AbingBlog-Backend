package service

import (
	"strings"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/model"
	"abingblog-backend/internal/repository"
)

// CategoryService 分类业务逻辑层
type CategoryService struct {
	repo *repository.CategoryRepo
}

func NewCategoryService(repo *repository.CategoryRepo) *CategoryService {
	return &CategoryService{repo: repo}
}

// CategoryReq 创建/更新分类的请求体
type CategoryReq struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
	Sort int    `json:"sort"`
}

func (s *CategoryService) List() ([]repository.CategoryWithCount, error) {
	return s.repo.ListWithCount()
}

// Create 创建分类：名称必填且唯一；slug 未填时默认用名称
func (s *CategoryService) Create(req CategoryReq) (*model.Category, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.New(errcode.CodeNameEmpty, "分类名称不能为空")
	}
	// 预检查名称唯一性：给前端返回明确的业务错误，而不是数据库唯一键报错
	if exist, err := s.repo.GetByName(name); err != nil {
		return nil, err
	} else if exist != nil {
		return nil, errcode.New(errcode.CodeNameExists, "分类名称已存在")
	}
	slug := strings.TrimSpace(req.Slug)
	if slug == "" {
		slug = name
	}
	c := &model.Category{Name: name, Slug: slug, Sort: req.Sort}
	if err := s.repo.Create(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Update 更新分类：名称唯一性排除自身；slug 传空表示不变
func (s *CategoryService) Update(id uint, req CategoryReq) (*model.Category, error) {
	c, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, errcode.New(errcode.CodeCategoryNotFound, "分类不存在")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.New(errcode.CodeNameEmpty, "分类名称不能为空")
	}
	if exist, err := s.repo.GetByName(name); err != nil {
		return nil, err
	} else if exist != nil && exist.ID != id {
		return nil, errcode.New(errcode.CodeNameExists, "分类名称已存在")
	}
	c.Name = name
	c.Sort = req.Sort
	if slug := strings.TrimSpace(req.Slug); slug != "" {
		c.Slug = slug
	}
	if err := s.repo.Update(c); err != nil {
		return nil, err
	}
	return c, nil
}

// Delete 删除分类：分类下还有文章（含草稿）则拒绝，防止文章变成无分类状态
func (s *CategoryService) Delete(id uint) error {
	c, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if c == nil {
		return errcode.New(errcode.CodeCategoryNotFound, "分类不存在")
	}
	n, err := s.repo.CountByCategory(id)
	if err != nil {
		return err
	}
	if n > 0 {
		return errcode.New(errcode.CodeCategoryNotEmpty, "分类下存在文章，不可删除")
	}
	return s.repo.Delete(id)
}
