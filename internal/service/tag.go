package service

import (
	"strings"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/model"
	"abingblog-backend/internal/repository"
)

// TagService 标签业务逻辑层
type TagService struct {
	repo *repository.TagRepo
}

func NewTagService(repo *repository.TagRepo) *TagService {
	return &TagService{repo: repo}
}

// TagReq 创建/更新标签的请求体
type TagReq struct {
	Name string `json:"name"`
}

func (s *TagService) List() ([]repository.TagWithCount, error) {
	return s.repo.ListWithCount()
}

// Create 创建标签；同名标签直接返回已存在的（幂等）——方便写文章时随手打新标签
func (s *TagService) Create(req TagReq) (*model.Tag, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.New(errcode.CodeNameEmpty, "标签名称不能为空")
	}
	if exist, err := s.repo.GetByName(name); err != nil {
		return nil, err
	} else if exist != nil {
		return exist, nil
	}
	t := &model.Tag{Name: name}
	if err := s.repo.Create(t); err != nil {
		return nil, err
	}
	return t, nil
}

// Update 更新标签：名称唯一性排除自身
func (s *TagService) Update(id uint, req TagReq) (*model.Tag, error) {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, errcode.New(errcode.CodeTagNotFound, "标签不存在")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.New(errcode.CodeNameEmpty, "标签名称不能为空")
	}
	if exist, err := s.repo.GetByName(name); err != nil {
		return nil, err
	} else if exist != nil && exist.ID != id {
		return nil, errcode.New(errcode.CodeNameExists, "标签名称已存在")
	}
	t.Name = name
	if err := s.repo.Update(t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *TagService) Delete(id uint) error {
	t, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}
	if t == nil {
		return errcode.New(errcode.CodeTagNotFound, "标签不存在")
	}
	return s.repo.Delete(id)
}
