package handler

import (
	"net/http"
	"strconv"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"
	"abingblog-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// ArticleHandler 文章接口层：解析请求 → 调 service → 封装响应，不含业务逻辑
type ArticleHandler struct {
	svc *service.ArticleService
}

func NewArticleHandler(svc *service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

// List 公共文章列表：只返回已发布文章
func (h *ArticleHandler) List(c *gin.Context) {
	q := service.ArticleQuery{
		Page:         atoiDefault(c.Query("page"), 1),
		PageSize:     atoiDefault(c.Query("page_size"), 10),
		Status:       "published",
		CategorySlug: c.Query("category_slug"),
		Tag:          c.Query("tag"),
		Keyword:      c.Query("keyword"),
	}
	list, total, err := h.svc.List(q)
	if err != nil {
		response.Error(c, err)
		return
	}
	dtos := make([]articleDTO, 0, len(list))
	for i := range list {
		dtos = append(dtos, toArticleDTO(&list[i], false))
	}
	response.OK(c, gin.H{"list": dtos, "total": total, "page": q.Page, "page_size": q.PageSize})
}

// AdminList 后台文章列表：包含草稿，可用 status 筛选
func (h *ArticleHandler) AdminList(c *gin.Context) {
	q := service.ArticleQuery{
		Page:     atoiDefault(c.Query("page"), 1),
		PageSize: atoiDefault(c.Query("page_size"), 10),
		Status:   c.Query("status"),
	}
	list, total, err := h.svc.List(q)
	if err != nil {
		response.Error(c, err)
		return
	}
	dtos := make([]articleDTO, 0, len(list))
	for i := range list {
		dtos = append(dtos, toArticleDTO(&list[i], false))
	}
	response.OK(c, gin.H{"list": dtos, "total": total, "page": q.Page, "page_size": q.PageSize})
}

// Get 文章详情（公共）：只允许已发布
func (h *ArticleHandler) Get(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的文章 ID")
		return
	}
	a, err := h.svc.GetPublicByID(id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, toArticleDTO(a, true))
}

// Create 创建文章（后台）
func (h *ArticleHandler) Create(c *gin.Context) {
	var req service.CreateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	a, err := h.svc.Create(req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, toArticleDTO(a, true))
}

// Update 全量更新文章（后台）
func (h *ArticleHandler) Update(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的文章 ID")
		return
	}
	var req service.CreateArticleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	a, err := h.svc.Update(id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, toArticleDTO(a, true))
}

// Delete 软删除文章（后台）
func (h *ArticleHandler) Delete(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的文章 ID")
		return
	}
	if err := h.svc.Delete(id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, nil)
}

// UpdateStatus 发布/撤回（后台）
func (h *ArticleHandler) UpdateStatus(c *gin.Context) {
	id := pathParamUint(c)
	if id == 0 {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "无效的文章 ID")
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, errcode.CodeBadParam, "请求参数错误: "+err.Error())
		return
	}
	a, err := h.svc.UpdateStatus(id, req.Status)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, toArticleDTO(a, true))
}

// atoiDefault 解析 int 查询参数，失败/缺省时用默认值
func atoiDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

// pathParamUint 解析路径参数 :id，无效时返回 0
func pathParamUint(c *gin.Context) uint {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	return uint(id)
}
