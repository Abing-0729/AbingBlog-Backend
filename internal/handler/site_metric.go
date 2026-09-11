package handler

import (
	"net/http"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"
	"abingblog-backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SiteMetricHandler struct{ svc *service.SiteMetricService }

func NewSiteMetricHandler(svc *service.SiteMetricService) *SiteMetricHandler {
	return &SiteMetricHandler{svc: svc}
}

func (h *SiteMetricHandler) RecordStart(c *gin.Context) {
	count, err := h.svc.RecordStart()
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.CodeInternal, "记录启动次数失败")
		return
	}
	response.OK(c, gin.H{"start_count": count})
}
