package handler

import (
	"net/http"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"
	"abingblog-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// ProjectHandler 项目接口层：解析请求 → 调 service → 封装响应
type ProjectHandler struct {
	svc *service.ProjectService
}

func NewProjectHandler(svc *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{svc: svc}
}

// List 公共项目列表：只返回已发布项目
func (h *ProjectHandler) List(c *gin.Context) {
	q := service.ProjectQuery{
		Page:     atoiDefault(c.Query("page"), 1),
		PageSize: atoiDefault(c.Query("page_size"), 10),
		Status:   "published",
	}
	list, total, err := h.svc.List(q)
	if err != nil {
		response.Error(c, err)
		return
	}
	dtos := make([]projectDTO, 0, len(list))
	for i := range list {
		dtos = append(dtos, toProjectDTO(&list[i]))
	}
	response.OK(c, gin.H{"list": dtos, "total": total, "page": q.Page, "page_size": q.PageSize})
}

// AdminList 后台项目列表：含草稿，可用 status 筛选
func (h *ProjectHandler) AdminList(c *gin.Context) {
	q := service.ProjectQuery{
		Page:     atoiDefault(c.Query("page"), 1),
		PageSize: atoiDefault(c.Query("page_size"), 10),
		Status:   c.Query("status"),
	}
	list, total, err := h.svc.List(q)
	if err != nil {
		response.Error(c, err)
		return
	}
	dtos := make([]projectDTO, 0, len(list))
	for i := range list {
		dtos = append(dtos, toProjectDTO(&list[i]))
	}
	response.OK(c, gin.H{"list": dtos, "total": total, "page": q.Page, "page_size": q.PageSize})
}

// Create 创建项目（后台）
func (h *ProjectHandler) Create(c *gin.Context) {
	var req service.ProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	p, err := h.svc.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, toProjectDTO(p))
}

// Update 全量更新项目（后台）
func (h *ProjectHandler) Update(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的项目 ID")
		return
	}
	var req service.ProjectReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	p, err := h.svc.Update(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, toProjectDTO(p))
}

// Delete 软删除项目（后台）
func (h *ProjectHandler) Delete(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的项目 ID")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}
