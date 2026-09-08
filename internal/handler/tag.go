package handler

import (
	"net/http"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"
	"abingblog-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// TagHandler 标签接口层
type TagHandler struct {
	svc *service.TagService
}

func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

// List 公共标签列表（含文章数）
func (h *TagHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// Create 创建标签（后台，同名幂等）
func (h *TagHandler) Create(c *gin.Context) {
	var req service.TagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	t, err := h.svc.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, tagDTO{ID: t.ID, Name: t.Name})
}

// Update 更新标签（后台）
func (h *TagHandler) Update(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的标签 ID")
		return
	}
	var req service.TagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	t, err := h.svc.Update(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, tagDTO{ID: t.ID, Name: t.Name})
}

// Delete 删除标签（后台）
func (h *TagHandler) Delete(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的标签 ID")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
