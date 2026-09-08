package handler

import (
	"net/http"

	"abingblog-backend/internal/errcode"
	"abingblog-backend/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// HealthHandler 健康检查：Docker / CD 部署时用它判断服务是否就绪
type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Check 检查 HTTP 服务 + 数据库连通性
func (h *HealthHandler) Check(c *gin.Context) {
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		response.Fail(c, http.StatusInternalServerError, errcode.CodeInternal, "数据库连接不可用")
		return
	}
	response.OK(c, gin.H{"status": "ok", "db": "ok"})
}
