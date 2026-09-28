package handler

import (
	"net/http"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"
	"abingblog-backend/internal/service"

	"github.com/gin-gonic/gin"
)

// VisitLogHandler 访客记录接口层（全部后台接口，JWT 由路由分组保证）
type VisitLogHandler struct {
	svc *service.VisitLogService
}

func NewVisitLogHandler(svc *service.VisitLogService) *VisitLogHandler {
	return &VisitLogHandler{svc: svc}
}

// List 访问明细（分页 + 搜索）
func (h *VisitLogHandler) List(c *gin.Context) {
	page := atoiDefault(c.Query("page"), 1)
	pageSize := atoiDefault(c.Query("page_size"), 20)
	list, total, err := h.svc.List(page, pageSize, c.Query("keyword"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, gin.H{
		"list": list, "total": total,
		"page": page, "page_size": pageSize,
	})
}

// Summary 访问统计汇总
func (h *VisitLogHandler) Summary(c *gin.Context) {
	stats, visitors, paths, err := h.svc.Summary()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.CodeInternal, "读取访问统计失败")
		return
	}
	response.OK(c, gin.H{
		"stats":     stats,
		"visitors":  visitors,
		"top_paths": paths,
	})
}
