package handler

import (
	"net/http"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"
	"abingblog-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// CategoryHandler 分类接口层
type CategoryHandler struct {
	svc *service.CategoryService
}

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// List 公共分类列表（含文章数）
func (h *CategoryHandler) List(c *gin.Context) {
	list, err := h.svc.List()
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, list)
}

// Create 创建分类（后台）
func (h *CategoryHandler) Create(c *gin.Context) {
	var req service.CategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	cat, err := h.svc.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, toCategoryDTO(cat))
}

// Update 更新分类（后台）
func (h *CategoryHandler) Update(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的分类 ID")
		return
	}
	var req service.CategoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	cat, err := h.svc.Update(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, toCategoryDTO(cat))
}

// Delete 删除分类（后台）
func (h *CategoryHandler) Delete(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的分类 ID")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
