package service

import (
	"strings"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/model"
	"abingblog-backend/internal/repository"
)

// FriendLinkService 友链业务层：参数校验、归属校验、审核状态机
type FriendLinkService struct {
	repo *repository.FriendLinkRepo
}

func NewFriendLinkService(repo *repository.FriendLinkRepo) *FriendLinkService {
	return &FriendLinkService{repo: repo}
}

// FriendLinkReq 创建/更新友链的请求体（前后端契约）
// 注意：ID 从 URL 路径取、OwnerID 从 X-Visitor-ID 头取、时间戳由 GORM 维护，
// 三者都不属于前端该传的字段，所以不放进 DTO。
type FriendLinkReq struct {
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	URL         string `json:"url"`
	Description string `json:"description"`
	Sort        int    `json:"sort"`
	Status      int    `json:"status"` // 仅管理端审核时生效；访客提交一律视为待审核
}

// validate 公共校验：名称与地址必填
func (s *FriendLinkService) validate(req *FriendLinkReq) error {
	if strings.TrimSpace(req.Name) == "" {
		return errcode.New(errcode.CodeFriendNameEmpty, "友链名称不能为空")
	}
	if strings.TrimSpace(req.URL) == "" {
		return errcode.New(errcode.CodeFriendURLEmpty, "友链地址不能为空")
	}
	return nil
}

// List 前台友链列表：只返回已上架的
func (s *FriendLinkService) List() ([]model.FriendLink, error) {
	return s.repo.ListEnabled()
}

// ListMine 访客"我的友链"：不限状态，能看到待审核/已驳回的
func (s *FriendLinkService) ListMine(visitorID string) ([]model.FriendLink, error) {
	return s.repo.ListByOwner(visitorID)
}

// Create 访客提交友链：归属记为提交者，状态强制为待审核（忽略请求里的 status）
func (s *FriendLinkService) Create(visitorID string, req FriendLinkReq) (*model.FriendLink, error) {
	if err := s.validate(&req); err != nil {
		return nil, err
	}
	f := &model.FriendLink{
		Name:        strings.TrimSpace(req.Name),
		Avatar:      req.Avatar,
		URL:         strings.TrimSpace(req.URL),
		Description: req.Description,
		Status:      model.FriendLinkPending, // 提交即待审核，能否上架由管理端决定
		OwnerID:     visitorID,
	}
	if err := s.repo.Insert(f); err != nil {
		return nil, err
	}
	return f, nil
}

// UpdateMine 访客修改自己的友链：校验归属，改完打回待审核（防止改内容绕过审核）
func (s *FriendLinkService) UpdateMine(id uint, visitorID string, req FriendLinkReq) (*model.FriendLink, error) {
	if err := s.validate(&req); err != nil {
		return nil, err
	}
	f, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, errcode.New(errcode.CodeFriendNotFound, "友链不存在")
	}
	if f.OwnerID != visitorID {
		return nil, errcode.New(errcode.CodeFriendNotOwner, "只能修改自己提交的友链")
	}

	f.Name = strings.TrimSpace(req.Name)
	f.Avatar = req.Avatar
	f.URL = strings.TrimSpace(req.URL)
	f.Description = req.Description
	f.Status = model.FriendLinkPending // 内容变了，重新走审核
	if err := s.repo.Update(f); err != nil {
		return nil, err
	}
	return f, nil
}

// DeleteMine 访客删除自己的友链：校验归属后软删除
func (s *FriendLinkService) DeleteMine(id uint, visitorID string) error {
	f, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if f == nil {
		return errcode.New(errcode.CodeFriendNotFound, "友链不存在")
	}
	if f.OwnerID != visitorID {
		return errcode.New(errcode.CodeFriendNotOwner, "只能删除自己提交的友链")
	}
	return s.repo.Delete(id)
}

// AdminList 后台友链列表：status < 0 查全部，否则按状态筛选
func (s *FriendLinkService) AdminList(status int) ([]model.FriendLink, error) {
	return s.repo.ListAll(status)
}

// AdminUpdate 管理端编辑 + 审核：改 status 就是审核动作（1 通过 / 2 驳回）
func (s *FriendLinkService) AdminUpdate(id uint, req FriendLinkReq) (*model.FriendLink, error) {
	if err := s.validate(&req); err != nil {
		return nil, err
	}
	if req.Status != model.FriendLinkPending &&
		req.Status != model.FriendLinkApproved &&
		req.Status != model.FriendLinkRejected {
		return nil, errcode.New(errcode.CodeFriendBadStatus, "友链状态只能为 0 待审核 / 1 上架 / 2 驳回")
	}
	f, err := s.repo.FindByID(id)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, errcode.New(errcode.CodeFriendNotFound, "友链不存在")
	}

	f.Name = strings.TrimSpace(req.Name)
	f.Avatar = req.Avatar
	f.URL = strings.TrimSpace(req.URL)
	f.Description = req.Description
	f.Status = req.Status // 审核结果落库
	f.Sort = req.Sort     // 上架时的展示顺序由管理员排
	if err := s.repo.Update(f); err != nil {
		return nil, err
	}
	return f, nil
}

// AdminDelete 管理端删除任意友链（软删除，数据可恢复）
func (s *FriendLinkService) AdminDelete(id uint) error {
	f, err := s.repo.FindByID(id)
	if err != nil {
		return err
	}
	if f == nil {
		return errcode.New(errcode.CodeFriendNotFound, "友链不存在")
	}
	return s.repo.Delete(id)
}
