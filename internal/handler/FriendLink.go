package handler

import (
	"net/http"
	"strings"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"
	"abingblog-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// FriendLinkHandler 友链接口层：解析请求 → 取访客身份 → 调 service → 封装响应
type FriendLinkHandler struct {
	svc *service.FriendLinkService
}

func NewFriendLinkHandler(svc *service.FriendLinkService) *FriendLinkHandler {
	return &FriendLinkHandler{svc: svc}
}

// visitorID 从 X-Visitor-ID 头取访客身份（前端 visitor.ts 自动附带）。
// 没有就报业务错误——归属校验全靠它，不能静默降级。
func visitorID(c *gin.Context) (string, error) {
	id := strings.TrimSpace(c.GetHeader("X-Visitor-ID"))
	if id == "" {
		return "", errcode.New(errcode.CodeFriendVisitorRequired, "缺少访客标识，请刷新页面后重试")
	}
	return id, nil
}

// List 公共友链列表：只返回已上架的，无需访客标识
func (h *FriendLinkHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"list": list})
}

// Mine 我的友链：按访客 UUID 查自己提交的全部（含待审核/已驳回）
func (h *FriendLinkHandler) Mine(c *gin.Context) {
	vid, err := visitorID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	list, err := h.svc.ListMine(vid)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"list": list})
}

// Create 访客提交友链：进待审核队列，管理员审核通过后前台可见
func (h *FriendLinkHandler) Create(c *gin.Context) {
	vid, err := visitorID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req service.FriendLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	f, err := h.svc.Create(vid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, f)
}

// UpdateMine 访客修改自己的友链：改完自动打回待审核
func (h *FriendLinkHandler) UpdateMine(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的友链 ID")
		return
	}
	vid, err := visitorID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	var req service.FriendLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	f, err := h.svc.UpdateMine(id, vid, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, f)
}

// DeleteMine 访客删除自己的友链（软删除）
func (h *FriendLinkHandler) DeleteMine(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的友链 ID")
		return
	}
	vid, err := visitorID(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	if err := h.svc.DeleteMine(id, vid); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// AdminList 后台友链列表：?status=0/1/2 筛选，缺省查全部
func (h *FriendLinkHandler) AdminList(c *gin.Context) {
	status := atoiDefault(c.Query("status"), -1) // -1 = 不过滤
	list, err := h.svc.AdminList(status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{"list": list})
}

// AdminUpdate 后台编辑 + 审核：body 里带 status 即审核动作（1 通过 / 2 驳回）
func (h *FriendLinkHandler) AdminUpdate(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的友链 ID")
		return
	}
	var req service.FriendLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	f, err := h.svc.AdminUpdate(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, f)
}

// AdminDelete 后台删除任意友链（软删除，数据可恢复）
func (h *FriendLinkHandler) AdminDelete(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的友链 ID")
		return
	}
	if err := h.svc.AdminDelete(id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
